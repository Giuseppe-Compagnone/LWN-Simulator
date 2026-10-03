package runtime

import (
	"fmt"
	"math"
	"time"

	contracts "github.com/Giuseppe-Compagnone/lwn-contracts/generated"
	"github.com/Giuseppe-Compagnone/lwn-engine/geometry"
	"github.com/Giuseppe-Compagnone/lwn-engine/types"
	"github.com/google/uuid"
)

func (e *Engine) processGatewayBeaconLocked(scheduled types.ScheduledEvent) []contracts.SimulationEvent {
	gateway, ok := e.registry.Gateway(scheduled.GatewayID)
	if !ok || !gateway.Active || gateway.Type != contracts.Virtual {
		return nil
	}

	e.metrics.ClassBBeacons++
	events := []contracts.SimulationEvent{e.newEventLocked(
		contracts.GatewayBeaconTransmitted,
		"virtual gateway beacon transmitted",
		"",
		gateway.ID,
		scheduled.ID,
	)}

	for _, device := range e.registry.ActiveDevices() {
		if device.Class != contracts.ClassB {
			continue
		}
		session := e.sessions[device.ID]
		if session == nil || !session.Joined || !deviceIsCoveredByGateway(device, gateway) {
			continue
		}

		selectedGateway := e.classBGatewayForDevice(device)
		if selectedGateway == nil || selectedGateway.ID != gateway.ID {
			continue
		}

		if session.ClassBNextPingEventID != "" {
			e.scheduler.Cancel(session.ClassBNextPingEventID)
		}
		if session.ClassBBeaconTimeoutID != "" {
			e.scheduler.Cancel(session.ClassBBeaconTimeoutID)
		}

		session.ClassBSynchronized = true
		session.ClassBGatewayID = gateway.ID
		session.LastBeaconAt = scheduled.At
		session.NextPingSlotAt = 0
		e.metrics.ClassBSynchronizations++
		events = append(events, e.newEventLocked(
			contracts.DeviceBeaconSynchronized,
			"Class B device synchronized with gateway beacon",
			device.ID,
			gateway.ID,
			scheduled.ID,
		))

		if pingEvent := e.scheduleClassBPingSlotLocked(device, session, scheduled.At+types.ClassBPingSlotPeriod); pingEvent != nil {
			events = append(events, *pingEvent)
		}
		timeoutID := uuid.NewString()
		session.ClassBBeaconTimeoutID = timeoutID
		if err := e.scheduler.Schedule(types.ScheduledEvent{
			ID: timeoutID, At: scheduled.At + 2*types.ClassBBeaconPeriod,
			Type: contracts.DeviceBeaconMissed, Message: "Class B beacon synchronization timeout",
			DeviceID: device.ID, GatewayID: gateway.ID,
			Kind: types.ScheduledEventClassBBeaconTimeout, WindowBaseAt: scheduled.At,
		}); err != nil {
			session.ClassBBeaconTimeoutID = ""
		}
	}

	events = append(events, e.scheduleClassBBeaconLocked(gateway.ID, scheduled.At+types.ClassBBeaconPeriod))
	return events
}

func (e *Engine) processClassBBeaconTimeoutLocked(scheduled types.ScheduledEvent) []contracts.SimulationEvent {
	device, session, ok := e.deviceSessionLocked(scheduled.DeviceID)
	if !ok || device.Class != contracts.ClassB || session == nil || !session.ClassBSynchronized {
		return nil
	}
	if session.LastBeaconAt != scheduled.WindowBaseAt {
		return nil
	}

	if session.ClassBNextPingEventID != "" {
		e.scheduler.Cancel(session.ClassBNextPingEventID)
	}
	session.ClassBNextPingEventID = ""
	session.NextPingSlotAt = 0
	session.ClassBSynchronized = false
	session.ClassBGatewayID = ""
	session.ClassBMissedBeacons++
	e.metrics.ClassBMissedBeacons++
	return []contracts.SimulationEvent{e.newEventLocked(
		contracts.DeviceBeaconMissed,
		"Class B device lost beacon synchronization",
		device.ID,
		scheduled.GatewayID,
		scheduled.ID,
	)}
}

func (e *Engine) processClassBPingSlotLocked(scheduled types.ScheduledEvent) []contracts.SimulationEvent {
	device, session, ok := e.deviceSessionLocked(scheduled.DeviceID)
	if !ok || !device.Active || device.Class != contracts.ClassB || session == nil ||
		!session.ClassBSynchronized || session.ClassBGatewayID != scheduled.GatewayID {
		return nil
	}

	session.ClassBNextPingEventID = ""
	session.NextPingSlotAt = scheduled.At
	e.metrics.ClassBPingSlots++
	events := []contracts.SimulationEvent{e.newEventLocked(
		contracts.ClassBPingSlotOpened,
		"Class B ping slot opened",
		device.ID,
		scheduled.GatewayID,
		scheduled.ID,
	)}

	if len(session.PendingClassBDownlinks) > 0 {
		downlink := session.PendingClassBDownlinks[0]
		session.PendingClassBDownlinks = session.PendingClassBDownlinks[1:]
		session.FrameCounterDown++
		e.metrics.ClassBDownlinks++
		events = append(events, e.newClassBDownlinkEventLocked(
			contracts.ClassBDownlinkTransmitted,
			"Class B downlink transmitted in ping slot",
			device,
			session,
			downlink,
			scheduled.ID,
		))
	}

	if next := e.scheduleClassBPingSlotLocked(device, session, scheduled.At+types.ClassBPingSlotPeriod); next != nil {
		events = append(events, *next)
	}
	return events
}

func (e *Engine) scheduleClassBBeaconLocked(gatewayID string, at time.Duration) contracts.SimulationEvent {
	eventID := uuid.NewString()
	if err := e.scheduler.Schedule(types.ScheduledEvent{
		ID: eventID, At: at, Type: contracts.GatewayBeaconTransmitted,
		Message: "virtual gateway beacon scheduled", GatewayID: gatewayID,
		Kind: types.ScheduledEventGatewayBeacon,
	}); err != nil {
		return e.newEventLocked(contracts.SimulationFailed, fmt.Sprintf("schedule Class B beacon: %v", err), "", gatewayID, eventID)
	}
	return e.newEventLocked(contracts.GatewayBeaconScheduled, "virtual gateway beacon scheduled", "", gatewayID, eventID)
}

func (e *Engine) scheduleClassBPingSlotLocked(device contracts.Device, session *types.DeviceSession, at time.Duration) *contracts.SimulationEvent {
	eventID := uuid.NewString()
	if err := e.scheduler.Schedule(types.ScheduledEvent{
		ID: eventID, At: at, Type: contracts.ClassBPingSlotOpened,
		Message: "Class B ping slot scheduled", DeviceID: device.ID,
		GatewayID: session.ClassBGatewayID, Kind: types.ScheduledEventClassBPingSlot,
	}); err != nil {
		return nil
	}
	session.ClassBNextPingEventID = eventID
	session.NextPingSlotAt = at
	event := e.newEventLocked(contracts.ClassBPingSlotScheduled, "Class B ping slot scheduled", device.ID, session.ClassBGatewayID, eventID)
	return &event
}

func (e *Engine) classBGatewayForDevice(device contracts.Device) *contracts.Gateway {
	gateways := e.coveredGateways(device)
	if len(gateways) == 0 {
		return nil
	}

	deviceLatitude, deviceLongitude := locationCoordinates(device)
	selected := gateways[0]
	selectedDistance := math.Inf(1)
	for _, gateway := range gateways {
		gatewayLatitude, gatewayLongitude := gatewayCoordinates(gateway)
		distance := geometry.DistanceMeters(deviceLatitude, deviceLongitude, gatewayLatitude, gatewayLongitude)
		// Registry.Gateways is sorted by ID, so a strict comparison keeps the
		// result deterministic when two gateways are equally distant.
		if distance < selectedDistance {
			selected = gateway
			selectedDistance = distance
		}
	}
	return &selected
}

func deviceIsCoveredByGateway(device contracts.Device, gateway contracts.Gateway) bool {
	deviceLatitude, deviceLongitude := locationCoordinates(device)
	gatewayLatitude, gatewayLongitude := gatewayCoordinates(gateway)
	return geometry.DistanceMeters(deviceLatitude, deviceLongitude, gatewayLatitude, gatewayLongitude) <= float64(device.AdvancedConfig.AntennaRange)
}

func (e *Engine) newClassBDownlinkEventLocked(
	eventType contracts.SimulationEventType,
	message string,
	device contracts.Device,
	session *types.DeviceSession,
	downlink types.ClassBDownlink,
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

func pointerTo[T any](value T) *T {
	return &value
}
