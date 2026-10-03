package engine

import (
	"encoding/base64"
	"math"
	"time"

	contracts "github.com/Giuseppe-Compagnone/lwn-contracts/generated"
	"github.com/Giuseppe-Compagnone/lwn-engine/geometry"
	"github.com/Giuseppe-Compagnone/lwn-engine/types"
	"github.com/google/uuid"
)

const (
	radioTransmitPowerDBm = 14.0
	radioNoiseFloorDBm    = -120.0
	radioCodingRate       = 1
	radioBandwidthHz      = 125_000
)

func (e *Engine) startRadioTransmissionLocked(
	device contracts.Device,
	scheduled types.ScheduledEvent,
	packetID string,
	frameCounter int64,
	confirmed bool,
) []contracts.SimulationEvent {
	channel := selectRadioChannel(device, frameCounter)
	airtime := calculateAirtime(device, channel.SpreadingFactor, channel.Bandwidth)
	gateways := e.coveredGateways(device)
	rssi, snr := calculateSignal(device, gateways, channel.Frequency)
	transmission := &types.RadioTransmission{
		PacketID:         packetID,
		DeviceID:         device.ID,
		FrameCounter:     frameCounter,
		Attempt:          scheduled.Attempt,
		Confirmed:        confirmed,
		StartAt:          scheduled.At,
		EndAt:            scheduled.At + airtime,
		ChannelFrequency: channel.Frequency,
		Bandwidth:        channel.Bandwidth,
		SpreadingFactor:  channel.SpreadingFactor,
		Airtime:          airtime,
		RSSI:             rssi,
		SNR:              snr,
		GatewayIDs:       gatewayIDs(gateways),
	}
	for _, active := range e.radio {
		if active.EndAt > transmission.StartAt &&
			active.ChannelFrequency == transmission.ChannelFrequency &&
			active.Bandwidth == transmission.Bandwidth &&
			active.SpreadingFactor == transmission.SpreadingFactor &&
			sharesGateway(active.GatewayIDs, transmission.GatewayIDs) {
			active.Collision = true
			transmission.Collision = true
		}
	}
	e.radio[packetID] = transmission

	if session := e.sessions[device.ID]; session != nil {
		session.LastRSSI = rssi
		session.LastSNR = snr
		session.LastAirtime = airtime
		session.LastChannel = channel.Frequency
	}
	e.metrics.TotalTransmissions++
	e.metrics.AverageRSSI = runningAverage(e.metrics.AverageRSSI, e.metrics.TotalTransmissions, rssi)
	e.metrics.AverageSNR = runningAverage(e.metrics.AverageSNR, e.metrics.TotalTransmissions, snr)
	e.metrics.AverageAirtimeMilliseconds = runningAverage(e.metrics.AverageAirtimeMilliseconds, e.metrics.TotalTransmissions, float64(airtime.Milliseconds()))

	events := []contracts.SimulationEvent{e.newRadioEventLocked(
		contracts.RadioTransmissionStarted,
		"radio transmission started",
		device.ID,
		"",
		scheduled.ID,
		packetID,
		transmission,
		false,
		"",
	)}
	completion := types.ScheduledEvent{
		ID: uuid.NewString(), At: transmission.EndAt, Type: contracts.RadioTransmissionCompleted,
		Message: "radio transmission completed", DeviceID: device.ID,
		Kind: types.ScheduledEventRadioComplete, Attempt: scheduled.Attempt,
		PacketID: packetID, FrameCounter: frameCounter, Confirmed: confirmed,
	}
	if err := e.scheduler.Schedule(completion); err != nil {
		delete(e.radio, packetID)
		events = append(events, e.newRadioEventLocked(
			contracts.RadioPacketLost, "radio completion could not be scheduled", device.ID, "", scheduled.ID,
			packetID, transmission, true, contracts.RandomLoss,
		))
		return events
	}
	events = append(events, e.newEventLocked(contracts.SchedulerEventScheduled, "radio transmission completion scheduled", device.ID, "", completion.ID))
	return events
}

func (e *Engine) processRadioTransmissionCompletedLocked(scheduled types.ScheduledEvent) []contracts.SimulationEvent {
	transmission, ok := e.radio[scheduled.PacketID]
	if !ok {
		e.metrics.FrameCounterErrors++
		return []contracts.SimulationEvent{e.newEventLocked(contracts.FrameCounterRejected, "radio transmission is no longer active", scheduled.DeviceID, "", scheduled.ID)}
	}
	delete(e.radio, scheduled.PacketID)

	events := []contracts.SimulationEvent{e.newRadioEventLocked(
		contracts.RadioTransmissionCompleted,
		"radio transmission completed",
		transmission.DeviceID, "", scheduled.ID, transmission.PacketID, transmission, transmission.Collision, "",
	)}
	reason, lost := e.radioLossReason(transmission)
	if lost {
		e.metrics.RadioPacketLosses++
		if reason == contracts.Collision {
			e.metrics.Collisions++
			events = append(events, e.newRadioEventLocked(
				contracts.RadioCollisionDetected, "radio collision detected", transmission.DeviceID, "", scheduled.ID,
				transmission.PacketID, transmission, true, reason,
			))
		}
		if reason == contracts.LowSNR {
			e.metrics.LowSNRLosses++
		}
		events = append(events, e.newRadioEventLocked(
			contracts.RadioPacketLost, "radio packet lost", transmission.DeviceID, "", scheduled.ID,
			transmission.PacketID, transmission, true, reason,
		))
		return append(events, e.completeRadioLossLocked(transmission, scheduled.ID)...)
	}

	for _, gatewayID := range transmission.GatewayIDs {
		e.metrics.TotalPacketsReceived++
		events = append(events, e.newRadioEventLocked(
			gatewayEventType(transmission), "virtual gateway received radio packet", transmission.DeviceID,
			gatewayID, scheduled.ID, transmission.PacketID, transmission, false, "",
		))
	}
	return append(events, e.completeRadioSuccessLocked(transmission, scheduled.ID)...)
}

func (e *Engine) completeRadioLossLocked(transmission *types.RadioTransmission, scheduledID string) []contracts.SimulationEvent {
	device, session, ok := e.deviceSessionLocked(transmission.DeviceID)
	if !ok {
		return nil
	}
	if session.PendingJoinRequest != nil && session.PendingJoinRequest.PacketID == transmission.PacketID {
		session.PendingJoinRequest.GatewayIDs = nil
		return []contracts.SimulationEvent{e.scheduleWindowLocked(device, radioWindowSource(transmission), contracts.RX1, transmission.PacketID, -1, false)}
	}
	if !transmission.Confirmed {
		e.metrics.DroppedUplinks++
		e.metrics.PacketSuccessRate = e.packetSuccessRate()
		events := []contracts.SimulationEvent{
			e.newPacketEventLocked(contracts.PacketDropped, "uplink dropped by radio model", transmission.DeviceID, "", scheduledID, transmission.PacketID),
			e.newPacketEventLocked(contracts.MetricsUpdated, "simulation metrics updated", transmission.DeviceID, "", scheduledID, transmission.PacketID),
		}
		return append(events, e.scheduleNextUplinkAfterTransmissionLocked(device, transmission)...)
	}
	if session.PendingUplink == nil || session.PendingUplink.PacketID != transmission.PacketID {
		return nil
	}
	session.PendingUplink.GatewayIDs = nil
	return []contracts.SimulationEvent{e.scheduleWindowLocked(device, radioWindowSource(transmission), contracts.RX1, transmission.PacketID, transmission.FrameCounter, true)}
}

func (e *Engine) completeRadioSuccessLocked(transmission *types.RadioTransmission, scheduledID string) []contracts.SimulationEvent {
	device, session, ok := e.deviceSessionLocked(transmission.DeviceID)
	if !ok {
		return nil
	}
	if session.PendingJoinRequest != nil && session.PendingJoinRequest.PacketID == transmission.PacketID {
		session.PendingJoinRequest.GatewayIDs = append([]string(nil), transmission.GatewayIDs...)
		return []contracts.SimulationEvent{e.scheduleWindowLocked(device, radioWindowSource(transmission), contracts.RX1, transmission.PacketID, -1, false)}
	}
	if session.PendingUplink == nil || session.PendingUplink.PacketID != transmission.PacketID {
		if !transmission.Confirmed {
			e.metrics.SuccessfulUplinks++
			e.metrics.PacketSuccessRate = e.packetSuccessRate()
			events := []contracts.SimulationEvent{e.newPacketEventLocked(contracts.MetricsUpdated, "simulation metrics updated", transmission.DeviceID, "", scheduledID, transmission.PacketID)}
			return append(events, e.scheduleNextUplinkAfterTransmissionLocked(device, transmission)...)
		}
		return nil
	}
	session.PendingUplink.GatewayIDs = append([]string(nil), transmission.GatewayIDs...)
	if transmission.Confirmed {
		return []contracts.SimulationEvent{e.scheduleWindowLocked(device, radioWindowSource(transmission), contracts.RX1, transmission.PacketID, transmission.FrameCounter, true)}
	}
	e.metrics.SuccessfulUplinks++
	e.metrics.PacketSuccessRate = e.packetSuccessRate()
	events := []contracts.SimulationEvent{e.newPacketEventLocked(contracts.MetricsUpdated, "simulation metrics updated", transmission.DeviceID, "", scheduledID, transmission.PacketID)}
	return append(events, e.scheduleNextUplinkAfterTransmissionLocked(device, transmission)...)
}

func (e *Engine) scheduleNextUplinkAfterTransmissionLocked(device contracts.Device, transmission *types.RadioTransmission) []contracts.SimulationEvent {
	nextAt := transmission.StartAt + durationSecondsFloat(device.PayloadConfig.UplinkInterval)
	if nextAt < transmission.EndAt {
		nextAt = transmission.EndAt
	}
	return e.scheduleUplinkAtLocked(device, nextAt)
}

func (e *Engine) newRadioEventLocked(eventType contracts.SimulationEventType, message, deviceID, gatewayID, scheduledID, packetID string, transmission *types.RadioTransmission, collision bool, reason contracts.SimulationRadioLossReason) contracts.SimulationEvent {
	event := e.newProtocolEventLocked(eventType, message, deviceID, gatewayID, scheduledID, packetID, transmission.FrameCounter, transmission.Attempt, transmission.Confirmed, nil)
	channelFrequency := transmission.ChannelFrequency
	bandwidth := transmission.Bandwidth
	spreadingFactor := int32(transmission.SpreadingFactor)
	airtimeMilliseconds := transmission.Airtime.Milliseconds()
	rssi := transmission.RSSI
	snr := transmission.SNR
	event.ChannelFrequency = &channelFrequency
	event.Bandwidth = &bandwidth
	event.SpreadingFactor = &spreadingFactor
	event.AirtimeMilliseconds = &airtimeMilliseconds
	event.RSSI = &rssi
	event.SNR = &snr
	event.Collision = &collision
	if reason != "" {
		reasonValue := reason
		event.RadioLossReason = &reasonValue
	}
	e.eventLog[len(e.eventLog)-1] = event
	return event
}

func (e *Engine) radioLossReason(transmission *types.RadioTransmission) (contracts.SimulationRadioLossReason, bool) {
	if len(transmission.GatewayIDs) == 0 {
		return contracts.NoGateway, true
	}
	if transmission.Collision {
		return contracts.Collision, true
	}
	if transmission.SNR < -20 {
		return contracts.LowSNR, true
	}
	if transmission.SNR < -10 {
		probability := math.Min(0.8, (-10-transmission.SNR)/20)
		if e.rng.Float64() < probability {
			return contracts.RandomLoss, true
		}
	}
	return "", false
}

func sharesGateway(left, right []string) bool {
	for _, leftID := range left {
		for _, rightID := range right {
			if leftID == rightID {
				return true
			}
		}
	}
	return false
}

func gatewayEventType(transmission *types.RadioTransmission) contracts.SimulationEventType {
	if transmission.Confirmed || transmission.FrameCounter > 0 {
		return contracts.GatewayPacketReceived
	}
	return contracts.JoinRequestReceived
}

func radioWindowSource(transmission *types.RadioTransmission) types.ScheduledEvent {
	return types.ScheduledEvent{
		At: transmission.StartAt, DeviceID: transmission.DeviceID, PacketID: transmission.PacketID,
		FrameCounter: transmission.FrameCounter, Attempt: transmission.Attempt, Confirmed: transmission.Confirmed,
	}
}

func selectRadioChannel(device contracts.Device, frameCounter int64) types.RadioChannel {
	channels := regionalChannels(device.LocationConfig.Region, spreadingFactor(device))
	index := int(frameCounter)
	if index < 0 {
		index = 0
	}
	return channels[index%len(channels)]
}

func regionalChannels(region contracts.DeviceRegion, sf int) []types.RadioChannel {
	frequencies := []int64{868_100_000, 868_300_000, 868_500_000}
	switch region {
	case contracts.EU433:
		frequencies = []int64{433_175_000, 433_375_000, 433_575_000}
	case contracts.AS923:
		frequencies = []int64{923_200_000, 923_400_000, 923_600_000}
	case contracts.AU915:
		frequencies = []int64{915_200_000, 915_400_000, 915_600_000, 915_800_000}
	case contracts.US915:
		frequencies = []int64{902_300_000, 902_500_000, 902_700_000, 902_900_000}
	case contracts.CN470, contracts.CN779:
		frequencies = []int64{470_300_000, 470_500_000, 779_500_000}
	case contracts.IN865:
		frequencies = []int64{865_062_500, 865_402_500, 865_985_000}
	case contracts.KR920:
		frequencies = []int64{920_900_000, 921_100_000, 921_300_000}
	case contracts.RU864:
		frequencies = []int64{868_900_000, 869_100_000, 869_300_000}
	}
	channels := make([]types.RadioChannel, len(frequencies))
	for index, frequency := range frequencies {
		channels[index] = types.RadioChannel{Frequency: frequency, Bandwidth: radioBandwidthHz, SpreadingFactor: sf}
	}
	return channels
}

func spreadingFactor(device contracts.Device) int {
	dataRate := 0
	if device.RX2Config.DataRate != nil {
		dataRate = *device.RX2Config.DataRate
	}
	if dataRate < 0 {
		dataRate = 0
	}
	if dataRate > 5 {
		dataRate = 5
	}
	return 12 - dataRate
}

func calculateAirtime(device contracts.Device, sf int, bandwidth int64) time.Duration {
	payloadSize := len([]byte(device.PayloadConfig.Payload))
	if device.PayloadConfig.Base64Encoded {
		if decoded, err := base64.StdEncoding.DecodeString(device.PayloadConfig.Payload); err == nil {
			payloadSize = len(decoded)
		}
	}
	payloadSize += 13
	symbolDuration := math.Pow(2, float64(sf)) / float64(bandwidth)
	denominator := 4 * float64(sf-2)
	payloadSymbols := 8.0
	if denominator > 0 {
		payloadSymbols += math.Max(math.Ceil((8*float64(payloadSize)-4*float64(sf)+28+16)/(denominator))*float64(radioCodingRate+4), 0)
	}
	seconds := (8 + 4.25 + payloadSymbols) * symbolDuration
	return time.Duration(math.Max(1, math.Ceil(seconds*float64(time.Second))))
}

func calculateSignal(device contracts.Device, gateways []contracts.Gateway, frequency int64) (float64, float64) {
	if len(gateways) == 0 {
		return -120, -30
	}
	minimumDistance := math.MaxFloat64
	latitude, longitude := locationCoordinates(device)
	for _, gateway := range gateways {
		gatewayLatitude, gatewayLongitude := gatewayCoordinates(gateway)
		distance := geometry.DistanceMeters(latitude, longitude, gatewayLatitude, gatewayLongitude)
		if distance < minimumDistance {
			minimumDistance = distance
		}
	}
	distanceKilometers := math.Max(minimumDistance/1000, 0.001)
	pathLoss := 32.44 + 20*math.Log10(float64(frequency)/1_000_000) + 20*math.Log10(distanceKilometers)
	rssi := radioTransmitPowerDBm - pathLoss
	return rssi, rssi - radioNoiseFloorDBm
}

func runningAverage(previous float64, count int64, value float64) float64 {
	if count <= 1 {
		return value
	}
	return previous + (value-previous)/float64(count)
}
