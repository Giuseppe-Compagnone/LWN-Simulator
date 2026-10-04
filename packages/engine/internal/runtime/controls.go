package runtime

import (
	"fmt"
	"reflect"
	"time"

	contracts "github.com/Giuseppe-Compagnone/lwn-contracts/generated"
	"github.com/Giuseppe-Compagnone/lwn-engine/types"
	"github.com/google/uuid"
)

// QueueUplink injects a manual device transmission at the current simulation
// time. An OTAA device that is not joined sends a join request instead.
func (e *Engine) QueueUplink(deviceID string) (string, error) {
	e.mu.Lock()
	device, session, ok := e.deviceSessionLocked(deviceID)
	if !ok {
		e.mu.Unlock()
		return "", fmt.Errorf("device %s is not registered", deviceID)
	}
	if !device.Active {
		e.mu.Unlock()
		return "", fmt.Errorf("device %s is inactive", deviceID)
	}
	if !runtimeControlStatus(e.state.Status) {
		e.mu.Unlock()
		return "", fmt.Errorf("%w: cannot queue uplink from %q", ErrInvalidTransition, e.state.Status)
	}
	if session.PendingUplink != nil || session.PendingJoinRequest != nil {
		e.mu.Unlock()
		return "", fmt.Errorf("device %s already has a protocol transaction in progress", deviceID)
	}
	event := initialDeviceEvent(device, session, e.eventTimestampLocked())
	if err := e.scheduler.Schedule(event); err != nil {
		e.mu.Unlock()
		return "", err
	}
	logged := e.newEventLocked(contracts.DeviceUplinkScheduled, "manual device uplink scheduled", device.ID, "", event.ID)
	e.mu.Unlock()
	e.signalWake()
	e.publish(logged)
	return event.ID, nil
}

// UpdateDevice replaces a registered device while preserving its protocol
// session when identity and activation credentials have not changed.
func (e *Engine) UpdateDevice(updated contracts.Device) error {
	e.mu.Lock()
	current, exists := e.registry.Device(updated.ID)
	if !exists {
		e.mu.Unlock()
		return fmt.Errorf("device %s is not registered", updated.ID)
	}
	devices := e.registry.Devices()
	for index := range devices {
		if devices[index].ID == updated.ID {
			devices[index] = updated
			break
		}
	}
	registry, err := NewRegistry(devices, e.registry.Gateways())
	if err != nil {
		e.mu.Unlock()
		return fmt.Errorf("validate updated device: %w", err)
	}
	identityChanged := deviceIdentityChanged(current, updated)
	runtimeChanged := identityChanged || current.Class != updated.Class
	e.registry = registry
	e.metrics.ActiveDevices = int64(len(registry.ActiveDevices()))
	if !updated.Active || runtimeChanged {
		e.cancelDeviceRuntimeLocked(updated.ID)
	}
	if identityChanged {
		e.sessions[updated.ID] = newDeviceSession(updated)
	}
	var scheduled *types.ScheduledEvent
	if updated.Active && (!current.Active || runtimeChanged) && runtimeControlStatus(e.state.Status) {
		event := initialDeviceEvent(updated, e.sessions[updated.ID], e.eventTimestampLocked())
		if err := e.scheduler.Schedule(event); err != nil {
			e.mu.Unlock()
			return fmt.Errorf("schedule updated device: %w", err)
		}
		scheduled = &event
	}
	logged := e.newEventLocked(contracts.DeviceRegistered, "device runtime configuration updated", updated.ID, "", "")
	if scheduled != nil {
		logged.ScheduledEventID = &scheduled.ID
	}
	e.mu.Unlock()
	e.signalWake()
	e.publish(logged)
	return nil
}

// UpdateGateway replaces a gateway configuration. Virtual gateway schedules
// are rebuilt immediately; real adapters are restarted when their endpoint or
// active state changes.
func (e *Engine) UpdateGateway(updated contracts.Gateway) error {
	e.mu.Lock()
	current, exists := e.registry.Gateway(updated.ID)
	if !exists {
		e.mu.Unlock()
		return fmt.Errorf("gateway %s is not registered", updated.ID)
	}
	gateways := e.registry.Gateways()
	for index := range gateways {
		if gateways[index].ID == updated.ID {
			gateways[index] = updated
			break
		}
	}
	registry, err := NewRegistry(e.registry.Devices(), gateways)
	if err != nil {
		e.mu.Unlock()
		return fmt.Errorf("validate updated gateway: %w", err)
	}
	restart := current.Active != updated.Active || current.Type != updated.Type ||
		!reflect.DeepEqual(current.GatewayIPv4, updated.GatewayIPv4) ||
		!reflect.DeepEqual(current.GatewayPort, updated.GatewayPort) ||
		!reflect.DeepEqual(current.KeepAlive, updated.KeepAlive)
	e.registry = registry
	e.metrics.ActiveGateways = int64(len(registry.ActiveGateways()))
	var adapter types.GatewayAdapter
	if restart {
		e.cancelGatewayRuntimeLocked(updated.ID)
		adapter = e.gatewayAdapters[updated.ID]
		delete(e.gatewayAdapters, updated.ID)
	}
	if runtimeState := e.gatewayRuntime[updated.ID]; runtimeState != nil {
		if !updated.Active || updated.Type == contracts.Real {
			runtimeState.State = contracts.Disconnected
		} else {
			runtimeState.State = contracts.Connected
		}
	}
	runContext := e.runContext
	if restart && updated.Active && runtimeControlStatus(e.state.Status) {
		e.scheduleGatewayRuntimeLocked(updated)
	}
	logged := e.newEventLocked(contracts.GatewayRegistered, "gateway runtime configuration updated", "", updated.ID, "")
	e.mu.Unlock()
	if adapter != nil {
		_ = adapter.Close()
	}
	if restart && updated.Active && updated.Type == contracts.Real && runContext != nil {
		e.startGatewayAdapter(runContext, updated)
	}
	e.signalWake()
	e.publish(logged)
	return nil
}

func runtimeControlStatus(status contracts.SimulationStatus) bool {
	return status == contracts.SimulationStatusRunning || status == contracts.SimulationStatusPaused
}

func initialDeviceEvent(device contracts.Device, session *types.DeviceSession, at time.Duration) types.ScheduledEvent {
	kind := types.ScheduledEventDeviceUplink
	eventType := contracts.DeviceUplinkTransmitted
	message := "device uplink scheduled"
	if session == nil || !session.Joined {
		kind = types.ScheduledEventJoinRequest
		eventType = contracts.DeviceJoinRequestTransmitted
		message = "device join request scheduled"
	}
	return types.ScheduledEvent{ID: uuid.NewString(), At: at, Type: eventType, Message: message, DeviceID: device.ID, Kind: kind, Attempt: 1}
}

func (e *Engine) cancelDeviceRuntimeLocked(deviceID string) {
	for _, scheduled := range e.scheduler.Events() {
		if scheduled.DeviceID == deviceID {
			e.scheduler.Cancel(scheduled.ID)
		}
	}
	for packetID, transmission := range e.radio {
		if transmission.DeviceID == deviceID {
			delete(e.radio, packetID)
		}
	}
	if session := e.sessions[deviceID]; session != nil {
		session.PendingUplink = nil
		session.PendingJoinRequest = nil
		session.PendingClassADownlinks = nil
		session.PendingClassBDownlinks = nil
		session.PendingClassCDownlinks = nil
	}
}

func (e *Engine) cancelGatewayRuntimeLocked(gatewayID string) {
	for _, scheduled := range e.scheduler.Events() {
		if scheduled.GatewayID == gatewayID {
			e.scheduler.Cancel(scheduled.ID)
		}
	}
	for _, session := range e.sessions {
		if session == nil || session.ClassBGatewayID != gatewayID {
			continue
		}
		session.ClassBSynchronized = false
		session.ClassBGatewayID = ""
		session.ClassBNextPingEventID = ""
		session.ClassBBeaconTimeoutID = ""
	}
}

func (e *Engine) scheduleGatewayRuntimeLocked(gateway contracts.Gateway) {
	e.scheduleClassBBeaconLocked(gateway.ID, e.eventTimestampLocked()+classBRealGatewayInitialDelay)
	if gateway.Type != contracts.Virtual || gateway.KeepAlive == nil {
		return
	}
	_ = e.scheduler.Schedule(types.ScheduledEvent{
		ID: uuid.NewString(), At: e.eventTimestampLocked() + time.Duration(*gateway.KeepAlive)*time.Second,
		Type: contracts.GatewayHeartbeat, Message: "virtual gateway heartbeat scheduled",
		GatewayID: gateway.ID, Kind: types.ScheduledEventGatewayHeartbeat,
	})
}

func deviceIdentityChanged(left, right contracts.Device) bool {
	return left.DevEUI != right.DevEUI || left.Activation != right.Activation ||
		!reflect.DeepEqual(left.OOTAConfig, right.OOTAConfig) || !reflect.DeepEqual(left.ABPConfig, right.ABPConfig)
}
