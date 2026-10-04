package runtime

import (
	"testing"
	"time"

	contracts "github.com/Giuseppe-Compagnone/lwn-contracts/generated"
	"github.com/Giuseppe-Compagnone/lwn-engine/types"
)

func TestUpdateDeviceAppliesRuntimeConfigurationAndPersistsIt(t *testing.T) {
	device := validDevice("a1c6e32b-4f0d-4b50-9fc5-000000000001")
	configureABP(&device)
	gateway := validGateway("6f0f1f30-7dc5-4bb3-a6f5-11b2ed9a7c01")
	engine, err := New(contracts.SimulationConfig{Speed: 1}, []contracts.Device{device}, []contracts.Gateway{gateway}, types.Options{Clock: types.NewManualClock(0)})
	if err != nil {
		t.Fatalf("create engine: %v", err)
	}

	updated := device
	updated.Name = "Updated sensor"
	updated.PayloadConfig.Payload = "new-payload"
	updated.PayloadConfig.MType = contracts.ConfirmedDataUp
	updated.LocationConfig.Latitude = float32Ptr(37.51)
	if err := engine.UpdateDevice(updated); err != nil {
		t.Fatalf("update device: %v", err)
	}
	if engine.sessions[device.ID].FrameCounterUp != 0 {
		t.Fatal("non-identity update unexpectedly reset protocol session")
	}
	checkpoint := engine.Checkpoint()
	if len(checkpoint.Devices) != 1 || checkpoint.Devices[0].Name != "Updated sensor" || checkpoint.Devices[0].PayloadConfig.Payload != "new-payload" {
		t.Fatalf("checkpoint did not retain updated hardware: %+v", checkpoint.Devices)
	}

	restored, err := New(contracts.SimulationConfig{Speed: 1}, []contracts.Device{device}, []contracts.Gateway{gateway}, types.Options{Clock: types.NewManualClock(0), Checkpoint: &checkpoint})
	if err != nil {
		t.Fatalf("restore engine: %v", err)
	}
	if got, _ := restored.registry.Device(device.ID); got.Name != "Updated sensor" || got.PayloadConfig.MType != contracts.ConfirmedDataUp {
		t.Fatalf("restored runtime lost update: %+v", got)
	}
}

func TestRuntimeControlsToggleEntitiesAndQueueManualUplink(t *testing.T) {
	clock := types.NewManualClock(0)
	device := validDevice("a1c6e32b-4f0d-4b50-9fc5-000000000001")
	configureABP(&device)
	device.Active = false
	gateway := validGateway("6f0f1f30-7dc5-4bb3-a6f5-11b2ed9a7c01")
	engine := newTestEngine(t, clock, device, gateway)
	defer stopEngine(t, engine)

	if _, err := engine.QueueUplink(device.ID); err == nil {
		t.Fatal("expected inactive device uplink to be rejected")
	}
	device.Active = true
	if err := engine.UpdateDevice(device); err != nil {
		t.Fatalf("activate device: %v", err)
	}
	if _, err := engine.QueueUplink(device.ID); err != nil {
		t.Fatalf("queue manual uplink: %v", err)
	}
	if err := clock.Advance(2 * time.Second); err != nil {
		t.Fatalf("advance clock: %v", err)
	}
	waitForEventType(t, engine.Events(), contracts.DeviceUplinkTransmitted)

	gateway.Active = false
	if err := engine.UpdateGateway(gateway); err != nil {
		t.Fatalf("deactivate gateway: %v", err)
	}
	snapshot := engine.Snapshot()
	if snapshot.Metrics.ActiveDevices != 1 || snapshot.Metrics.ActiveGateways != 0 {
		t.Fatalf("runtime active counts were not updated: %+v", snapshot.Metrics)
	}
	gateway.Active = true
	if err := engine.UpdateGateway(gateway); err != nil {
		t.Fatalf("reactivate gateway: %v", err)
	}
	engine.mu.RLock()
	hasBeacon, hasHeartbeat := false, false
	for _, scheduled := range engine.scheduler.Events() {
		if scheduled.GatewayID != gateway.ID {
			continue
		}
		hasBeacon = hasBeacon || scheduled.Kind == types.ScheduledEventGatewayBeacon
		hasHeartbeat = hasHeartbeat || scheduled.Kind == types.ScheduledEventGatewayHeartbeat
	}
	engine.mu.RUnlock()
	if !hasBeacon || !hasHeartbeat {
		t.Fatalf("reactivated virtual gateway did not rebuild schedules: beacon=%v heartbeat=%v", hasBeacon, hasHeartbeat)
	}
}

func TestUpdateDeviceResetsSessionWhenCredentialsChange(t *testing.T) {
	device := validDevice("a1c6e32b-4f0d-4b50-9fc5-000000000001")
	configureABP(&device)
	engine, err := New(contracts.SimulationConfig{Speed: 1}, []contracts.Device{device}, nil, types.Options{Clock: types.NewManualClock(0)})
	if err != nil {
		t.Fatalf("create engine: %v", err)
	}
	engine.sessions[device.ID].FrameCounterUp = 42
	device.ABPConfig.AppSKey = "11111111111111111111111111111111"
	if err := engine.UpdateDevice(device); err != nil {
		t.Fatalf("update credentials: %v", err)
	}
	if engine.sessions[device.ID].FrameCounterUp == 42 {
		t.Fatal("credential update did not reset protocol session")
	}
}
