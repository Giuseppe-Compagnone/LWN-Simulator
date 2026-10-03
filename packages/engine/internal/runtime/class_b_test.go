package runtime

import (
	"testing"
	"time"

	contracts "github.com/Giuseppe-Compagnone/lwn-contracts/generated"
	"github.com/Giuseppe-Compagnone/lwn-engine/types"
)

func TestClassBBeaconSynchronizationAndPingSlotDownlink(t *testing.T) {
	clock := types.NewManualClock(0)
	device := validDevice("a1c6e32b-4f0d-4b50-9fc5-000000000011")
	configureABP(&device)
	device.Class = contracts.ClassB
	device.DevEUI = "70B3D57ED0000011"
	gateway := validGateway("6f0f1f30-7dc5-4bb3-a6f5-11b2ed9a7c11")

	engine, err := New(
		contracts.SimulationConfig{Speed: 1},
		[]contracts.Device{device},
		[]contracts.Gateway{gateway},
		types.Options{Clock: clock, EventBuffer: 512},
	)
	if err != nil {
		t.Fatalf("create Class B engine: %v", err)
	}
	startEngine(t, engine)
	defer stopEngine(t, engine)
	if _, err := engine.QueueClassBDownlink(types.ClassBDownlink{
		DeviceID: device.ID,
		Payload:  []byte("class-b-downlink"),
		FPort:    10,
		DataRate: 3,
	}); err != nil {
		t.Fatalf("queue Class B downlink before synchronization: %v", err)
	}

	waitForEventType(t, engine.Events(), contracts.GatewayBeaconTransmitted)
	waitForEventType(t, engine.Events(), contracts.DeviceBeaconSynchronized)
	if err := clock.Advance(time.Second); err != nil {
		t.Fatalf("advance to first Class B ping slot: %v", err)
	}
	waitForEventType(t, engine.Events(), contracts.ClassBPingSlotOpened)
	downlink := waitForEventType(t, engine.Events(), contracts.ClassBDownlinkTransmitted)
	if downlink.DeviceID == nil || *downlink.DeviceID != device.ID {
		t.Fatalf("downlink references the wrong device: %+v", downlink.DeviceID)
	}
	if downlink.PayloadSize == nil || *downlink.PayloadSize != int64(len("class-b-downlink")) {
		t.Fatalf("unexpected downlink payload size: %+v", downlink.PayloadSize)
	}
	if downlink.FPort == nil || *downlink.FPort != 10 || downlink.DataRate == nil || *downlink.DataRate != 3 {
		t.Fatalf("unexpected downlink radio metadata: %+v", downlink)
	}

	snapshot := engine.Snapshot()
	runtimeDevice := snapshot.Devices[0]
	if !runtimeDevice.ClassBSynchronized || runtimeDevice.LastBeaconTimestampMilliseconds != 0 || runtimeDevice.NextPingSlotTimestampMilliseconds <= 1000 {
		t.Fatalf("unexpected Class B runtime state: %+v", runtimeDevice)
	}
	if snapshot.Metrics.ClassBBeacons != 1 || snapshot.Metrics.ClassBSynchronizations != 1 ||
		snapshot.Metrics.ClassBPingSlots != 1 || snapshot.Metrics.ClassBDownlinks != 1 {
		t.Fatalf("unexpected Class B metrics: %+v", snapshot.Metrics)
	}
	if runtimeDevice.FrameCounterDown != 1 {
		t.Fatalf("expected Class B downlink to advance downlink frame counter: %+v", runtimeDevice)
	}
}

func TestClassBDeviceLosesSynchronizationAfterBeaconTimeout(t *testing.T) {
	clock := types.NewManualClock(0)
	device := validDevice("a1c6e32b-4f0d-4b50-9fc5-000000000012")
	configureABP(&device)
	device.Class = contracts.ClassB
	device.DevEUI = "70B3D57ED0000012"
	gateway := validGateway("6f0f1f30-7dc5-4bb3-a6f5-11b2ed9a7c12")

	engine := newTestEngine(t, clock, device, gateway)
	defer stopEngine(t, engine)
	waitForEventType(t, engine.Events(), contracts.GatewayBeaconTransmitted)
	waitForEventType(t, engine.Events(), contracts.DeviceBeaconSynchronized)

	engine.mu.Lock()
	for _, scheduled := range engine.scheduler.Events() {
		if (scheduled.Kind == types.ScheduledEventGatewayBeacon && scheduled.GatewayID == gateway.ID) ||
			(scheduled.Kind == types.ScheduledEventClassBPingSlot && scheduled.DeviceID == device.ID) {
			engine.scheduler.Cancel(scheduled.ID)
		}
	}
	engine.mu.Unlock()

	if err := clock.Advance(2 * types.ClassBBeaconPeriod); err != nil {
		t.Fatalf("advance beyond Class B beacon timeout: %v", err)
	}
	waitForEventType(t, engine.Events(), contracts.DeviceBeaconMissed)

	snapshot := engine.Snapshot()
	if snapshot.Devices[0].ClassBSynchronized {
		t.Fatalf("device should have lost Class B synchronization: %+v", snapshot.Devices[0])
	}
	if snapshot.Devices[0].ClassBMissedBeacons != 1 || snapshot.Metrics.ClassBMissedBeacons != 1 {
		t.Fatalf("unexpected missed beacon state: device=%+v metrics=%+v", snapshot.Devices[0], snapshot.Metrics)
	}
	if snapshot.State.Status != contracts.SimulationStatusRunning {
		t.Fatalf("simulation should remain running after beacon timeout: %q", snapshot.State.Status)
	}
}

func TestClassBUsesNearestCoveredGateway(t *testing.T) {
	clock := types.NewManualClock(0)
	device := validDevice("a1c6e32b-4f0d-4b50-9fc5-000000000013")
	configureABP(&device)
	device.Class = contracts.ClassB
	device.DevEUI = "70B3D57ED0000013"

	farGateway := validGateway("00000000-0000-4000-8000-000000000013")
	farGateway.Latitude = float32Ptr(37.505)
	farGateway.GatewayEUI = "A840410001000113"
	farGateway.MacAddress = "02:00:00:10:00:13"
	nearGateway := validGateway("ffffffff-ffff-4fff-bfff-ffffffffffff")
	nearGateway.GatewayEUI = "A840410001000114"
	nearGateway.MacAddress = "02:00:00:10:00:14"

	engine := newTestEngine(t, clock, device, farGateway, nearGateway)
	defer stopEngine(t, engine)
	synchronized := waitForEventType(t, engine.Events(), contracts.DeviceBeaconSynchronized)
	if synchronized.GatewayID == nil || *synchronized.GatewayID != nearGateway.ID {
		t.Fatalf("Class B selected the wrong gateway: %+v", synchronized.GatewayID)
	}
}

func TestClassBDownlinkSurvivesCheckpointRestore(t *testing.T) {
	firstClock := types.NewManualClock(0)
	device := validDevice("a1c6e32b-4f0d-4b50-9fc5-000000000014")
	configureABP(&device)
	device.Class = contracts.ClassB
	device.DevEUI = "70B3D57ED0000014"
	gateway := validGateway("6f0f1f30-7dc5-4bb3-a6f5-11b2ed9a7c14")

	first := newTestEngine(t, firstClock, device, gateway)
	waitForEventType(t, first.Events(), contracts.DeviceBeaconSynchronized)
	if _, err := first.QueueClassBDownlink(types.ClassBDownlink{DeviceID: device.ID, Payload: []byte("persisted")}); err != nil {
		t.Fatalf("queue Class B downlink: %v", err)
	}
	checkpoint := first.Checkpoint()
	stopEngine(t, first)

	secondClock := types.NewManualClock(0)
	second, err := New(
		contracts.SimulationConfig{Speed: 1},
		[]contracts.Device{device},
		[]contracts.Gateway{gateway},
		types.Options{Clock: secondClock, Checkpoint: &checkpoint},
	)
	if err != nil {
		t.Fatalf("restore Class B engine: %v", err)
	}
	startEngine(t, second)
	defer stopEngine(t, second)

	if err := secondClock.Advance(time.Second); err != nil {
		t.Fatalf("advance restored engine to ping slot: %v", err)
	}
	transmitted := waitForEventType(t, second.Events(), contracts.ClassBDownlinkTransmitted)
	if transmitted.PayloadSize == nil || *transmitted.PayloadSize != int64(len("persisted")) {
		t.Fatalf("restored downlink was not transmitted: %+v", transmitted)
	}
}

func TestClassBDownlinkValidation(t *testing.T) {
	clock := types.NewManualClock(0)
	device := validDevice("a1c6e32b-4f0d-4b50-9fc5-000000000015")
	configureABP(&device)
	device.Class = contracts.ClassB
	device.DevEUI = "70B3D57ED0000015"
	gateway := validGateway("6f0f1f30-7dc5-4bb3-a6f5-11b2ed9a7c15")
	engine, err := New(contracts.SimulationConfig{Speed: 1}, []contracts.Device{device}, []contracts.Gateway{gateway}, types.Options{Clock: clock})
	if err != nil {
		t.Fatalf("create Class B engine: %v", err)
	}
	startEngine(t, engine)
	defer stopEngine(t, engine)

	invalid := []types.ClassBDownlink{
		{DeviceID: device.ID, Payload: []byte("payload"), DataRate: -2},
		{DeviceID: device.ID, Payload: []byte("payload"), FPort: 224},
		{DeviceID: device.ID, Payload: make([]byte, 243), DataRate: 5},
		{DeviceID: device.ID, Payload: []byte("payload"), ID: "not-a-uuid"},
	}
	for index, downlink := range invalid {
		if _, err := engine.QueueClassBDownlink(downlink); err == nil {
			t.Fatalf("invalid downlink %d was accepted", index)
		}
	}
}
