package runtime

import (
	"context"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"

	contracts "github.com/Giuseppe-Compagnone/lwn-contracts/generated"
	"github.com/Giuseppe-Compagnone/lwn-engine/lorawan"
	"github.com/Giuseppe-Compagnone/lwn-engine/regional"
	"github.com/Giuseppe-Compagnone/lwn-engine/types"
	"github.com/google/uuid"
)

func (e *Engine) startGatewayAdapters(ctx context.Context) {
	if e.gatewayAdapterFactory == nil {
		return
	}

	for _, gateway := range e.registry.ActiveGateways() {
		if gateway.Type == contracts.Virtual {
			virtualFactory, ok := e.gatewayAdapterFactory.(types.VirtualGatewayAdapterFactory)
			if !ok || !virtualFactory.SupportsVirtualGateways() {
				continue
			}
		} else if gateway.Type != contracts.Real {
			continue
		}
		e.startGatewayAdapter(ctx, gateway)
	}
}

func (e *Engine) startGatewayAdapter(ctx context.Context, gateway contracts.Gateway) {
	if e.gatewayAdapterFactory == nil || ctx == nil {
		return
	}
	adapter, err := e.gatewayAdapterFactory.NewGatewayAdapter(gateway)
	if err != nil {
		e.publishGatewayError(gateway.ID, fmt.Errorf("create gateway adapter: %w", err), false)
		return
	}

	e.mu.Lock()
	e.gatewayAdapters[gateway.ID] = adapter
	connecting := e.updateGatewayRuntimeLocked(gateway.ID, contracts.Connecting, "", true)
	e.mu.Unlock()
	e.publish(connecting)

	e.gatewayWG.Add(1)
	go func() {
		defer e.gatewayWG.Done()
		e.consumeGatewayAdapter(ctx, gateway.ID, adapter)
	}()
	if err := adapter.Start(ctx); err != nil {
		e.publishGatewayError(gateway.ID, fmt.Errorf("start gateway adapter: %w", err), false)
		_ = adapter.Close()
	}
}

func (e *Engine) consumeGatewayAdapter(ctx context.Context, gatewayID string, adapter types.GatewayAdapter) {
	packets := adapter.Packets()
	adapterEvents := adapter.Events()

	for packets != nil || adapterEvents != nil {
		select {
		case <-ctx.Done():
			return
		case event, ok := <-adapterEvents:
			if !ok {
				adapterEvents = nil
				continue
			}
			if event.GatewayID == "" {
				event.GatewayID = gatewayID
			}
			e.handleGatewayAdapterEvent(event)
		case packet, ok := <-packets:
			if !ok {
				packets = nil
				continue
			}
			if packet.GatewayID == "" {
				packet.GatewayID = gatewayID
			}
			e.handleGatewayPacket(packet)
		}
	}
}

func (e *Engine) handleGatewayAdapterEvent(adapterEvent types.GatewayAdapterEvent) {
	e.mu.Lock()

	runtime, ok := e.gatewayRuntime[adapterEvent.GatewayID]
	if !ok {
		e.mu.Unlock()
		return
	}

	if adapterEvent.Heartbeat {
		runtime.State = contracts.Connected
		runtime.LastHeartbeat = e.clock.Now()
		e.metrics.GatewayHeartbeats++
		event := e.newGatewayEventLocked(
			contracts.GatewayHeartbeat,
			"gateway heartbeat received",
			adapterEvent.GatewayID,
			contracts.Connected,
			nil,
			"",
			"",
		)
		e.mu.Unlock()
		e.publish(event)
		return
	}

	wasConnected := runtime.State == contracts.Connected
	var pendingPackets []types.GatewayPacket
	eventType := contracts.GatewayNetworkError
	message := "gateway network error"
	events := make([]contracts.SimulationEvent, 0, 2)
	switch adapterEvent.State {
	case contracts.Connected:
		runtime.State = contracts.Connected
		runtime.LastNetworkError = ""
		e.metrics.GatewayConnections++
		eventType = contracts.GatewayConnected
		message = "gateway connected"
		if !wasConnected {
			pendingPackets = e.takeGatewayPacketsLocked(adapterEvent.GatewayID)
		}
	case contracts.Disconnected:
		runtime.State = contracts.Disconnected
		eventType = contracts.GatewayDisconnected
		message = "gateway disconnected"
	case contracts.Reconnecting:
		runtime.State = contracts.Reconnecting
		runtime.ConnectionAttempts++
		e.metrics.GatewayReconnects++
		eventType = contracts.GatewayReconnecting
		message = "gateway reconnecting"
	case contracts.Error:
		runtime.State = contracts.Error
		runtime.LastNetworkError = adapterEvent.Error
		e.metrics.GatewayNetworkErrors++
		if adapterEvent.Timeout {
			e.metrics.GatewayTimeouts++
		}
		if adapterEvent.Packet != nil {
			copyPacket := *adapterEvent.Packet
			copyPacket.Payload = append([]byte(nil), adapterEvent.Packet.Payload...)
			e.gatewayPackets = append(e.gatewayPackets, copyPacket)
		}
	default:
		e.mu.Unlock()
		return
	}

	events = append(events, e.newGatewayEventLocked(
		eventType,
		message,
		adapterEvent.GatewayID,
		adapterEvent.State,
		nil,
		"",
		adapterEvent.Error,
	))
	if adapterEvent.State == contracts.Connected && !wasConnected {
		if beacon := e.rescheduleRealGatewayBeaconLocked(adapterEvent.GatewayID); beacon != nil {
			events = append(events, *beacon)
		}
	}
	e.mu.Unlock()
	for _, event := range events {
		e.publish(event)
	}
	if len(pendingPackets) > 0 {
		e.dispatchGatewayPackets(pendingPackets)
	}
}

func (e *Engine) handleGatewayPacket(packet types.GatewayPacket) {
	e.mu.Lock()
	runtime, ok := e.gatewayRuntime[packet.GatewayID]
	if !ok {
		e.mu.Unlock()
		return
	}
	runtime.IngressPackets++
	e.metrics.GatewayIngressPackets++
	payloadSize := int64(len(packet.Payload))
	event := e.newGatewayEventLocked(
		contracts.GatewayPacketIngress,
		"packet received from gateway transport",
		packet.GatewayID,
		contracts.Connected,
		&payloadSize,
		"",
		"",
	)
	if packet.Frequency > 0 {
		event.ChannelFrequency = &packet.Frequency
	}
	if packet.Bandwidth > 0 {
		event.Bandwidth = &packet.Bandwidth
	}
	if packet.SpreadingFactor > 0 {
		spreadingFactor := int32(packet.SpreadingFactor)
		event.SpreadingFactor = &spreadingFactor
	}
	if packet.RSSI != 0 {
		rssi := packet.RSSI
		event.RSSI = &rssi
	}
	if packet.SNR != 0 {
		snr := packet.SNR
		event.SNR = &snr
	}
	events := []contracts.SimulationEvent{event}
	if packet.Kind == types.GatewayPacketDownlink || packet.Kind == types.GatewayPacketClassBDownlink {
		if decoded, err := lorawan.Parse(packet.Payload); err == nil {
			events = append(events, e.processExternalDownlinkLocked(packet, decoded)...)
		} else {
			events = append(events, e.newPacketEventLocked(
				contracts.DeviceDownlinkDropped,
				"downlink received from gateway transport is not a valid LoRaWAN frame",
				"", packet.GatewayID, "", uuid.NewString(),
			))
		}
	} else if decoded, err := lorawan.Parse(packet.Payload); err == nil {
		events = append(events, e.processRealUplinkLocked(packet, decoded)...)
	}
	pendingPackets := e.takeGatewayPacketsLocked(packet.GatewayID)
	e.mu.Unlock()
	for _, event := range events {
		e.publish(event)
	}
	if len(pendingPackets) > 0 {
		e.dispatchGatewayPackets(pendingPackets)
	}
}

func (e *Engine) processExternalDownlinkLocked(packet types.GatewayPacket, decoded lorawan.Packet) []contracts.SimulationEvent {
	if decoded.MType != lorawan.MTypeUnconfirmedDataDown && decoded.MType != lorawan.MTypeConfirmedDataDown {
		return []contracts.SimulationEvent{e.newPacketEventLocked(
			contracts.DeviceDownlinkDropped,
			"external gateway downlink is not a data frame",
			"", packet.GatewayID, "", uuid.NewString(),
		)}
	}
	device, session := e.realDeviceSessionLocked(decoded.DevAddr)
	if session == nil {
		return []contracts.SimulationEvent{e.newPacketEventLocked(
			contracts.DeviceDownlinkDropped,
			"external gateway downlink targets an unknown device address",
			"", packet.GatewayID, "", uuid.NewString(),
		)}
	}
	if len(session.NwkSKey) != 16 || !lorawan.VerifyDataMIC(decoded, session.NwkSKey, 1) {
		e.metrics.FrameCounterErrors++
		return []contracts.SimulationEvent{e.newPacketEventLocked(
			contracts.DeviceDownlinkDropped,
			"external gateway downlink has an invalid MIC",
			device.ID, packet.GatewayID, "", uuid.NewString(),
		)}
	}
	if decoded.FCnt <= uint32(session.FrameCounterDown) {
		e.metrics.FrameCounterErrors++
		return []contracts.SimulationEvent{e.newProtocolEventLocked(
			contracts.FrameCounterRejected,
			"external gateway downlink frame counter is not newer",
			device.ID, packet.GatewayID, "", uuid.NewString(), int64(decoded.FCnt), 1, decoded.Confirmed, nil,
		)}
	}
	downlink := types.Downlink{
		ID:        uuid.NewString(),
		DeviceID:  device.ID,
		DataRate:  gatewayPacketDataRate(device, packet),
		Confirmed: decoded.Confirmed,
		FPending:  decoded.FPending,
		ACK:       decoded.ACK,
	}
	if decoded.FPort != nil {
		downlink.FPort = int(*decoded.FPort)
		key := session.AppSKey
		if *decoded.FPort == 0 {
			key = session.NwkSKey
		}
		payload, err := lorawan.DecryptPayload(decoded, key, 1)
		if err != nil {
			return []contracts.SimulationEvent{e.newPacketEventLocked(
				contracts.DeviceDownlinkDropped,
				"external gateway downlink payload could not be decrypted",
				device.ID, packet.GatewayID, "", uuid.NewString(),
			)}
		}
		downlink.Payload = payload
	}
	session.FrameCounterDown = int64(decoded.FCnt)
	event := e.newDownlinkEventLocked(
		contracts.DeviceDownlinkTransmitted,
		"downlink received from external gateway transport",
		device, session, downlink, "",
	)
	event.GatewayID = &packet.GatewayID
	if packet.Frequency > 0 {
		event.ChannelFrequency = &packet.Frequency
	}
	if packet.SpreadingFactor > 0 {
		spreadingFactor := int32(packet.SpreadingFactor)
		event.SpreadingFactor = &spreadingFactor
	}
	if packet.Bandwidth > 0 {
		event.Bandwidth = &packet.Bandwidth
	}
	e.eventLog[len(e.eventLog)-1] = event
	events := []contracts.SimulationEvent{event}
	events = append(events, e.applyDownlinkEffectsLocked(device, session, downlink, e.clock.Now())...)
	return events
}

func (e *Engine) processRealUplinkLocked(packet types.GatewayPacket, decoded lorawan.Packet) []contracts.SimulationEvent {
	if decoded.MType == lorawan.MTypeJoinRequest {
		for _, device := range e.registry.ActiveDevices() {
			if device.OOTAConfig == nil || !sameEUI(device.OOTAConfig.JoinEUI, decoded.JoinEUI) || !sameEUI(device.DevEUI, decoded.DevEUI) {
				continue
			}
			appKey, keyErr := lorawan.DecodeHex(device.OOTAConfig.AppKey)
			if keyErr != nil || !lorawan.VerifyJoinRequestMIC(decoded, appKey) {
				return []contracts.SimulationEvent{e.newPacketEventLocked(
					contracts.PacketDropped,
					"OTAA join request from real gateway has an invalid MIC",
					"", packet.GatewayID, "", uuid.NewString(),
				)}
			}
			session := e.sessions[device.ID]
			if session == nil {
				return nil
			}
			e.metrics.JoinRequests++
			if session.Joined {
				return []contracts.SimulationEvent{e.newProtocolEventLocked(
					contracts.JoinRequestReceived,
					"OTAA join request received from real gateway",
					device.ID, packet.GatewayID, "", uuid.NewString(), -1, 1, false, nil,
				)}
			}
			session.DevNonce = decoded.DevNonce
			session.Joined = true
			session.JoinEUI = device.OOTAConfig.JoinEUI
			session.DeviceAddress = joinedDeviceAddress(session)
			deriveOTAAKeys(device, session)
			joinAcceptQueued := e.queueRealJoinAcceptLocked(packet, device, session, appKey)
			e.metrics.JoinAccepts++
			packetID := uuid.NewString()
			events := []contracts.SimulationEvent{
				e.newProtocolEventLocked(contracts.JoinRequestReceived, "OTAA join request received from real gateway", device.ID, packet.GatewayID, "", packetID, -1, 1, false, nil),
				e.newProtocolEventLocked(contracts.JoinAcceptReceived, "OTAA join accepted for real gateway uplink", device.ID, packet.GatewayID, "", packetID, -1, 1, false, nil),
			}
			if !joinAcceptQueued {
				return append(events, e.newPacketEventLocked(contracts.PacketDropped, "OTAA join accept could not be queued for the real gateway", device.ID, packet.GatewayID, "", packetID))
			}
			return events
		}
		return nil
	}
	if decoded.MType != lorawan.MTypeUnconfirmedDataUp && decoded.MType != lorawan.MTypeConfirmedDataUp {
		return nil
	}
	device, session := e.realDeviceSessionLocked(decoded.DevAddr)
	if session == nil {
		return nil
	}
	if len(session.NwkSKey) == 16 && !lorawan.VerifyDataMIC(decoded, session.NwkSKey, 0) {
		return []contracts.SimulationEvent{e.newPacketEventLocked(
			contracts.PacketDropped,
			"real gateway uplink has an invalid MIC",
			device.ID, packet.GatewayID, "", uuid.NewString(),
		)}
	}
	if decoded.FPort != nil && len(decoded.FRMPayload) > 0 {
		key := session.AppSKey
		if *decoded.FPort == 0 {
			key = session.NwkSKey
		}
		if _, err := lorawan.DecryptPayload(decoded, key, 0); err != nil {
			return []contracts.SimulationEvent{e.newPacketEventLocked(
				contracts.PacketDropped,
				"real gateway uplink payload could not be decrypted",
				device.ID, packet.GatewayID, "", uuid.NewString(),
			)}
		}
		if dataRate := gatewayPacketDataRate(device, packet); dataRate >= 0 {
			maximum := maximumPayloadSize(device.LocationConfig.Region, dataRate)
			if len(decoded.FRMPayload) > maximum {
				return []contracts.SimulationEvent{e.newPacketEventLocked(
					contracts.PacketDropped,
					fmt.Sprintf("real gateway uplink payload exceeds the regional maximum of %d bytes", maximum),
					device.ID, packet.GatewayID, "", uuid.NewString(),
				)}
			}
		}
	}
	if !device.FrameConfig.DisableFrameCounterValidation && decoded.FCnt <= session.LastExternalFrameCounter {
		frameCounter := int64(decoded.FCnt)
		e.metrics.FrameCounterErrors++
		return []contracts.SimulationEvent{e.newProtocolEventLocked(
			contracts.FrameCounterRejected,
			"real gateway uplink rejected because its frame counter is not newer",
			device.ID, packet.GatewayID, "", uuid.NewString(), frameCounter, 1, decoded.Confirmed, nil,
		)}
	}
	frameCounter := decoded.FCnt
	session.LastExternalFrameCounter = decoded.FCnt
	if int64(frameCounter) > session.FrameCounterUp {
		session.FrameCounterUp = int64(frameCounter)
	}
	e.metrics.TotalUplinks++
	e.metrics.SuccessfulUplinks++
	e.metrics.TotalPacketsReceived++
	e.metrics.PacketSuccessRate = e.packetSuccessRate()
	if packet.RSSI != 0 {
		session.LastRSSI = packet.RSSI
	}
	if packet.SNR != 0 {
		session.LastSNR = packet.SNR
	}
	if packet.Frequency > 0 {
		session.LastChannel = packet.Frequency
	}
	session.LastPayloadSize = len(decoded.FRMPayload)
	if decoded.FPort != nil {
		session.LastFPort = int(*decoded.FPort)
	}
	if dataRate := gatewayPacketDataRate(device, packet); dataRate >= 0 {
		session.CurrentDataRate = dataRate
		session.CurrentSpreadingFactor = spreadingFactorForDataRate(device.LocationConfig.Region, dataRate)
	}
	packetID := uuid.NewString()
	events := []contracts.SimulationEvent{e.newProtocolEventLocked(
		contracts.DeviceUplinkTransmitted,
		"uplink received from real gateway",
		device.ID, packet.GatewayID, "", packetID, int64(frameCounter), 1, decoded.Confirmed, nil,
	)}
	if len(session.PendingClassADownlinks) > 0 {
		downlink := session.PendingClassADownlinks[0]
		session.PendingClassADownlinks = session.PendingClassADownlinks[1:]
		downlink.ACK = decoded.Confirmed
		if len(session.PendingClassADownlinks) > 0 {
			downlink.FPending = true
		}
		dataRate := gatewayPacketDataRate(device, packet)
		if dataRate < 0 {
			dataRate = currentDeviceDataRate(session, device)
		}
		dataRate = regional.ClampDataRate(device.LocationConfig.Region, dataRate-session.RX1DataRateOffset)
		downlink.DataRate = dataRate
		session.FrameCounterDown++
		if gateway, ok := e.registry.Gateway(packet.GatewayID); ok {
			e.queueDataDownlinkLocked(gateway, device, session, downlink, packet.Frequency, dataRate, e.clock.Now()+session.ReceiveDelay)
		}
		downlinkEvent := e.newDownlinkEventLocked(
			contracts.DeviceDownlinkTransmitted,
			"Class A downlink scheduled through real gateway",
			device,
			session,
			downlink,
			"",
		)
		downlinkEvent.GatewayID = &packet.GatewayID
		downlinkEvent.ChannelFrequency = &packet.Frequency
		e.eventLog[len(e.eventLog)-1] = downlinkEvent
		events = append(events, downlinkEvent)
		events = append(events, e.applyDownlinkEffectsLocked(device, session, downlink, e.clock.Now()+session.ReceiveDelay)...)
		return events
	}
	if decoded.Confirmed {
		if ack := e.queueRealACKLocked(packet, device, session, decoded); ack {
			events = append(events, e.newProtocolEventLocked(
				contracts.UplinkACKReceived,
				"ACK scheduled for real gateway uplink",
				device.ID, packet.GatewayID, "", packetID, int64(frameCounter), 1, true, nil,
			))
		}
	}
	return events
}

func (e *Engine) realDeviceSessionLocked(devAddr uint32) (contracts.Device, *types.DeviceSession) {
	for _, device := range e.registry.ActiveDevices() {
		session := e.sessions[device.ID]
		if session == nil {
			continue
		}
		if parsed, err := hex.DecodeString(session.DeviceAddress); err == nil && len(parsed) == 4 {
			value := binary.BigEndian.Uint32(parsed)
			if value == devAddr {
				return device, session
			}
		}
	}
	return contracts.Device{}, nil
}

func (e *Engine) queueRealACKLocked(packet types.GatewayPacket, device contracts.Device, session *types.DeviceSession, uplink lorawan.Packet) bool {
	if len(session.NwkSKey) != 16 || len(session.AppSKey) != 16 {
		return false
	}
	devAddrBytes, err := hex.DecodeString(session.DeviceAddress)
	if err != nil || len(devAddrBytes) != 4 {
		return false
	}
	devAddr := binary.BigEndian.Uint32(devAddrBytes)
	payload, err := lorawan.BuildDataFrame(lorawan.DataFrameOptions{
		DevAddr:   devAddr,
		FCnt:      uint32(session.FrameCounterDown + 1),
		Direction: 1,
		ACK:       true,
		NwkSKey:   session.NwkSKey,
		AppSKey:   session.AppSKey,
	})
	if err != nil {
		return false
	}
	delay := durationSeconds(device.RX1Config.Delay)
	transmitAt := time.Now().Add(time.Duration(float64(delay) / e.state.Speed))
	dataRate := gatewayPacketDataRate(device, packet)
	if dataRate < 0 {
		dataRate = currentDeviceDataRate(session, device)
	}
	spreadingFactor := spreadingFactorForDataRate(device.LocationConfig.Region, dataRate)
	bandwidth := gatewayPacketBandwidth(packet)
	e.gatewayPackets = append(e.gatewayPackets, types.GatewayPacket{
		GatewayID:       packet.GatewayID,
		Kind:            types.GatewayPacketDownlink,
		Payload:         payload,
		Frequency:       packet.Frequency,
		Bandwidth:       bandwidth,
		SpreadingFactor: spreadingFactor,
		Power:           classBBeaconPower,
		DataRate:        fmt.Sprintf("SF%dBW%d", spreadingFactor, bandwidth/1000),
		TransmitAt:      transmitAt,
	})
	session.FrameCounterDown++
	_ = uplink
	return true
}

func (e *Engine) queueRealJoinAcceptLocked(packet types.GatewayPacket, device contracts.Device, session *types.DeviceSession, appKey []byte) bool {
	devAddrBytes, err := hex.DecodeString(session.DeviceAddress)
	if err != nil || len(devAddrBytes) != 4 {
		return false
	}
	devAddr := binary.BigEndian.Uint32(devAddrBytes)
	payload, err := lorawan.BuildJoinAccept(lorawan.JoinAcceptOptions{
		AppNonce:          0,
		NetID:             0,
		DevAddr:           devAddr,
		RX2DataRate:       intValue(device.RX2Config.DataRate),
		RX1DataRateOffset: intValue(device.RX1Config.DataRateOffset),
		RXDelay:           intValue(device.RX1Config.Delay),
		AppKey:            appKey,
	})
	if err != nil {
		return false
	}
	delay := durationSeconds(device.RX1Config.Delay)
	transmitAt := time.Now().Add(time.Duration(float64(delay) / e.state.Speed))
	bandwidth := gatewayPacketBandwidth(packet)
	e.gatewayPackets = append(e.gatewayPackets, types.GatewayPacket{
		GatewayID:       packet.GatewayID,
		Kind:            types.GatewayPacketDownlink,
		Payload:         payload,
		Frequency:       packet.Frequency,
		Bandwidth:       bandwidth,
		SpreadingFactor: spreadingFactorForDataRate(device.LocationConfig.Region, currentDeviceDataRate(session, device)),
		Power:           classBBeaconPower,
		DataRate:        fmt.Sprintf("SF%dBW%d", spreadingFactorForDataRate(device.LocationConfig.Region, currentDeviceDataRate(session, device)), bandwidth/1000),
		TransmitAt:      transmitAt,
	})
	return true
}

func sameEUI(value string, numeric uint64) bool {
	parsed, err := hex.DecodeString(value)
	if err != nil || len(parsed) != 8 {
		return false
	}
	return binary.BigEndian.Uint64(parsed) == numeric
}

func (e *Engine) processVirtualGatewayHeartbeatLocked(scheduled types.ScheduledEvent) []contracts.SimulationEvent {
	gateway, ok := e.registry.Gateway(scheduled.GatewayID)
	if !ok || !gateway.Active || gateway.Type != contracts.Virtual || gateway.KeepAlive == nil {
		return nil
	}
	if runtime := e.gatewayRuntime[gateway.ID]; runtime != nil {
		runtime.State = contracts.Connected
		runtime.LastHeartbeat = scheduled.At
	}
	e.metrics.GatewayHeartbeats++
	event := e.newGatewayEventLocked(
		contracts.GatewayHeartbeat,
		"virtual gateway heartbeat received",
		gateway.ID,
		contracts.Connected,
		nil,
		"",
		"",
	)
	next := types.ScheduledEvent{
		ID: uuid.NewString(), At: scheduled.At + time.Duration(*gateway.KeepAlive)*time.Second,
		Type: contracts.GatewayHeartbeat, Message: "virtual gateway heartbeat scheduled",
		GatewayID: gateway.ID, Kind: types.ScheduledEventGatewayHeartbeat,
	}
	if err := e.scheduler.Schedule(next); err != nil {
		return []contracts.SimulationEvent{event, e.newEventLocked(contracts.PacketDropped, "virtual gateway heartbeat could not be scheduled", "", gateway.ID, next.ID)}
	}
	return []contracts.SimulationEvent{event}
}

func parseGatewayDataRate(value string) int {
	if !strings.HasPrefix(value, "SF") {
		return -1
	}
	parts := strings.SplitN(strings.TrimPrefix(value, "SF"), "BW", 2)
	if len(parts) != 2 {
		return -1
	}
	sf, err := strconv.Atoi(parts[0])
	if err != nil || sf < 7 || sf > 12 {
		return -1
	}
	return 12 - sf
}

func gatewayPacketDataRate(device contracts.Device, packet types.GatewayPacket) int {
	spreadingFactor := packet.SpreadingFactor
	if spreadingFactor == 0 {
		if parsed := parseGatewayDataRate(packet.DataRate); parsed >= 0 {
			spreadingFactor = 12 - parsed
		}
	}
	bandwidth := packet.Bandwidth
	if bandwidth == 0 {
		bandwidth = radioBandwidthHz
	}
	if dataRate, ok := regional.DataRateForModulation(device.LocationConfig.Region, spreadingFactor, bandwidth); ok {
		return dataRate
	}
	return -1
}

func gatewayPacketBandwidth(packet types.GatewayPacket) int64 {
	if packet.Bandwidth > 0 {
		return packet.Bandwidth
	}
	return radioBandwidthHz
}

func (e *Engine) publishGatewayError(gatewayID string, err error, timeout bool) {
	e.handleGatewayAdapterEvent(types.GatewayAdapterEvent{
		GatewayID: gatewayID,
		State:     contracts.Error,
		Error:     err.Error(),
		Timeout:   timeout,
	})
}

func (e *Engine) updateGatewayRuntimeLocked(
	gatewayID string,
	state contracts.GatewayConnectionState,
	errorMessage string,
	countAttempt bool,
) contracts.SimulationEvent {
	runtime := e.gatewayRuntime[gatewayID]
	if runtime == nil {
		runtime = &types.GatewayRuntime{}
		e.gatewayRuntime[gatewayID] = runtime
	}
	runtime.State = state
	runtime.LastNetworkError = errorMessage
	if countAttempt {
		runtime.ConnectionAttempts++
	}
	return e.newGatewayEventLocked(
		contracts.GatewayConnecting,
		"gateway connection starting",
		gatewayID,
		state,
		nil,
		"",
		errorMessage,
	)
}

func (e *Engine) newGatewayEventLocked(
	eventType contracts.SimulationEventType,
	message string,
	gatewayID string,
	state contracts.GatewayConnectionState,
	payloadSize *int64,
	packetID string,
	errorMessage string,
) contracts.SimulationEvent {
	event := e.newEventLocked(eventType, message, "", gatewayID, "")
	if packetID != "" {
		event.PacketID = &packetID
	}
	event.GatewayState = &state
	if payloadSize != nil {
		event.PayloadSize = payloadSize
	}
	if errorMessage != "" {
		event.Error = &errorMessage
	}
	e.eventLog[len(e.eventLog)-1] = event
	return event
}

func (e *Engine) stopGatewayAdapters() {
	e.mu.Lock()
	adapters := make([]types.GatewayAdapter, 0, len(e.gatewayAdapters))
	for gatewayID, adapter := range e.gatewayAdapters {
		adapters = append(adapters, adapter)
		delete(e.gatewayAdapters, gatewayID)
	}
	e.mu.Unlock()

	for _, adapter := range adapters {
		_ = adapter.Close()
	}
	if closer, ok := e.gatewayAdapterFactory.(types.GatewayAdapterFactoryCloser); ok {
		_ = closer.Close()
	}
}

func (e *Engine) SendGatewayPacket(ctx context.Context, gatewayID string, payload []byte) error {
	if ctx == nil {
		return fmt.Errorf("send gateway packet context cannot be nil")
	}
	return e.sendGatewayPacket(ctx, types.GatewayPacket{
		GatewayID: gatewayID,
		Payload:   append([]byte(nil), payload...),
	})
}

func (e *Engine) sendGatewayPacket(ctx context.Context, packet types.GatewayPacket) error {
	if ctx == nil {
		return fmt.Errorf("send gateway packet context cannot be nil")
	}

	e.mu.RLock()
	adapter, ok := e.gatewayAdapters[packet.GatewayID]
	gateway, gatewayExists := e.registry.Gateway(packet.GatewayID)
	e.mu.RUnlock()
	// Virtual gateways are externalized in phase 2 only for uplinks. Their
	// downlinks and Class-B beacons are still simulated in-process; the
	// Gateway Bridge receive path is introduced with the downlink phase.
	if gatewayExists && gateway.Type == contracts.Virtual && packet.Kind != types.GatewayPacketUplink {
		return nil
	}
	if !ok || adapter == nil {
		return fmt.Errorf("%w: %s", ErrGatewayTransportUnavailable, packet.GatewayID)
	}

	var err error
	if packet.Kind == types.GatewayPacketUplink {
		uplinkAdapter, ok := adapter.(types.GatewayUplinkAdapter)
		if !ok {
			return fmt.Errorf("gateway %s does not support uplink transport", packet.GatewayID)
		}
		err = uplinkAdapter.SendUplink(ctx, packet)
	} else {
		err = adapter.Send(ctx, packet)
	}
	if err != nil {
		e.mu.Lock()
		e.gatewayPackets = append(e.gatewayPackets, cloneGatewayPackets([]types.GatewayPacket{packet})...)
		e.mu.Unlock()
		e.publishGatewayError(packet.GatewayID, err, false)
		return err
	}

	e.mu.Lock()
	runtime := e.gatewayRuntime[packet.GatewayID]
	if runtime != nil {
		runtime.EgressPackets++
	}
	e.metrics.GatewayEgressPackets++
	payloadSize := int64(len(packet.Payload))
	event := e.newGatewayEventLocked(
		contracts.GatewayPacketEgress,
		"packet sent to gateway transport",
		packet.GatewayID,
		contracts.Connected,
		&payloadSize,
		"",
		"",
	)
	e.mu.Unlock()
	e.publish(event)
	return nil
}

func (e *Engine) drainGatewayPacketsLocked() []types.GatewayPacket {
	if len(e.gatewayPackets) == 0 {
		return nil
	}
	packets := append([]types.GatewayPacket(nil), e.gatewayPackets...)
	e.gatewayPackets = e.gatewayPackets[:0]
	return packets
}

func (e *Engine) takeGatewayPacketsLocked(gatewayID string) []types.GatewayPacket {
	if len(e.gatewayPackets) == 0 {
		return nil
	}
	packets := make([]types.GatewayPacket, 0)
	remaining := e.gatewayPackets[:0]
	for _, packet := range e.gatewayPackets {
		if packet.GatewayID == gatewayID {
			packets = append(packets, packet)
			continue
		}
		remaining = append(remaining, packet)
	}
	e.gatewayPackets = remaining
	return packets
}

func (e *Engine) dispatchGatewayPackets(packets []types.GatewayPacket) {
	for _, packet := range packets {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		_ = e.sendGatewayPacket(ctx, packet)
		cancel()
	}
}
