package runtime

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"time"

	contracts "github.com/Giuseppe-Compagnone/lwn-contracts/generated"
	"github.com/Giuseppe-Compagnone/lwn-engine/lorawan"
	"github.com/Giuseppe-Compagnone/lwn-engine/types"
)

func (e *Engine) newDownlinkEventLocked(
	eventType contracts.SimulationEventType,
	message string,
	device contracts.Device,
	session *types.DeviceSession,
	downlink types.Downlink,
	scheduledEventID string,
) contracts.SimulationEvent {
	event := e.newPacketEventLocked(eventType, message, device.ID, session.ClassBGatewayID, scheduledEventID, downlink.ID)
	payloadSize := int64(len(downlink.Payload))
	fPort := downlink.FPort
	dataRate := downlink.DataRate
	if dataRate < 0 {
		dataRate = session.CurrentDataRate
	}
	frequency := int64(device.RX2Config.ChannelFrequency)
	event.PayloadSize = &payloadSize
	event.FPort = &fPort
	event.DataRate = &dataRate
	event.ChannelFrequency = &frequency
	event.Confirmed = pointerTo(false)
	frameCounter := session.FrameCounterDown
	event.FrameCounter = &frameCounter
	e.eventLog[len(e.eventLog)-1] = event
	return event
}

func (e *Engine) transmitQueuedClassADownlinkLocked(
	device contracts.Device,
	session *types.DeviceSession,
	scheduled types.ScheduledEvent,
) (contracts.SimulationEvent, bool) {
	if len(session.PendingClassADownlinks) == 0 {
		return contracts.SimulationEvent{}, false
	}
	gateways := e.classBCoveredGateways(device)
	if len(gateways) == 0 {
		return contracts.SimulationEvent{}, false
	}
	downlink := session.PendingClassADownlinks[0]
	session.PendingClassADownlinks = session.PendingClassADownlinks[1:]
	session.FrameCounterDown++
	gateway := gateways[0]
	if gateway.Type == contracts.Real {
		e.queueDataDownlinkLocked(gateway, device, session, downlink, scheduled.ChannelFrequency, scheduled.DataRate, scheduled.At+durationSeconds(device.RX1Config.Delay))
	}
	event := e.newDownlinkEventLocked(
		contracts.DeviceDownlinkTransmitted,
		"Class A downlink transmitted in receive window",
		device,
		session,
		downlink,
		scheduled.ID,
	)
	if event.GatewayID == nil {
		event.GatewayID = &gateway.ID
		e.eventLog[len(e.eventLog)-1] = event
	}
	return event, true
}

func (e *Engine) queueDataDownlinkLocked(
	gateway contracts.Gateway,
	device contracts.Device,
	session *types.DeviceSession,
	downlink types.Downlink,
	frequency int64,
	dataRate int,
	at time.Duration,
) {
	if frequency <= 0 {
		frequency = int64(device.RX2Config.ChannelFrequency)
	}
	if dataRate < 0 {
		dataRate = currentDeviceDataRate(session, device)
	}
	devAddr, err := hex.DecodeString(session.DeviceAddress)
	if err != nil || len(devAddr) != 4 || len(session.NwkSKey) != 16 || len(session.AppSKey) != 16 {
		return
	}
	address := binary.BigEndian.Uint32(devAddr)
	fPort := byte(downlink.FPort)
	payload, err := lorawan.BuildDataFrame(lorawan.DataFrameOptions{
		DevAddr:   address,
		FCnt:      uint32(session.FrameCounterDown),
		FPort:     &fPort,
		Payload:   downlink.Payload,
		Direction: 1,
		NwkSKey:   session.NwkSKey,
		AppSKey:   session.AppSKey,
	})
	if err != nil {
		return
	}
	target := time.Now().Add(classBRealGatewayTimingLead)
	if at > 0 {
		target = time.Now().Add(time.Duration(float64(at-e.clock.Now()) / e.state.Speed))
		if target.Before(time.Now()) {
			target = time.Now().Add(classBRealGatewayTimingLead)
		}
	}
	spreadingFactor := spreadingFactorForDataRate(dataRate)
	e.gatewayPackets = append(e.gatewayPackets, types.GatewayPacket{
		GatewayID:       gateway.ID,
		Kind:            GatewayPacketKindForDeviceClass(device.Class),
		Payload:         payload,
		Frequency:       frequency,
		Bandwidth:       classBBeaconBandwidth,
		SpreadingFactor: spreadingFactor,
		Power:           classBBeaconPower,
		DataRate:        fmt.Sprintf("SF%dBW125", spreadingFactor),
		TransmitAt:      target,
	})
}

func GatewayPacketKindForDeviceClass(class contracts.DeviceClass) types.GatewayPacketKind {
	if class == contracts.ClassB {
		return types.GatewayPacketClassBDownlink
	}
	return types.GatewayPacketDownlink
}
