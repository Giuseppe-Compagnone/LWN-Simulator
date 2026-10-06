package runtime

import (
	"testing"
	"time"

	contracts "github.com/Giuseppe-Compagnone/lwn-contracts/generated"
	"github.com/Giuseppe-Compagnone/lwn-engine/types"
)

func TestMACCommandsUpdateDeviceRuntimeState(t *testing.T) {
	device := validDevice("a1c6e32b-4f0d-4b50-9fc5-000000000001")
	configureABP(&device)
	session := newDeviceSession(device)
	dataRate, txPower, repetitions := 2, 4, 3
	exponent := 3
	rx2Rate, offset := 3, 1
	frequency := int64(869_525_000)
	delay := 4 * time.Second

	commands := []types.MACCommand{
		{Type: types.MACLinkADRReq, DataRate: &dataRate, TxPower: &txPower, NbTrans: &repetitions},
		{Type: types.MACDutyCycleReq, MaxDutyCycleExponent: &exponent},
		{Type: types.MACRXParamSetupReq, DataRate: &rx2Rate, RX1DataRateOffset: &offset, Frequency: &frequency},
		{Type: types.MACRXTimingSetupReq, Delay: &delay},
	}
	for _, command := range commands {
		if err := validateMACCommand(device.LocationConfig.Region, command); err != nil {
			t.Fatalf("validate command %s: %v", command.Type, err)
		}
	}
	applyMACCommands(device, session, commands)

	if session.CurrentDataRate != 2 || session.CurrentSpreadingFactor != 10 || session.CurrentTxPower != 4 || session.UnconfirmedRepetitions != 3 {
		t.Fatalf("LinkADRReq was not applied: %+v", session)
	}
	if session.MaximumDutyCycle != 0.125 || session.RX2DataRate != 3 || session.RX2Frequency != frequency || session.RX1DataRateOffset != 1 || session.ReceiveDelay != delay {
		t.Fatalf("MAC runtime settings were not applied: %+v", session)
	}
}

func TestQueueMACCommandUsesClassCDownlinkPath(t *testing.T) {
	clock := types.NewManualClock(0)
	device := validDevice("a1c6e32b-4f0d-4b50-9fc5-000000000001")
	configureABP(&device)
	device.Class = contracts.ClassC
	gateway := validGateway("6f0f1f30-7dc5-4bb3-a6f5-11b2ed9a7c01")
	engine := newTestEngine(t, clock, device, gateway)
	defer stopEngine(t, engine)

	dataRate, repetitions := 2, 3
	if _, err := engine.QueueMACCommand(device.ID, types.MACCommand{Type: types.MACLinkADRReq, DataRate: &dataRate, NbTrans: &repetitions}); err != nil {
		t.Fatalf("queue MAC command: %v", err)
	}
	dataRate = 5
	repetitions = 1
	if err := clock.Advance(time.Millisecond); err != nil {
		t.Fatalf("advance clock: %v", err)
	}
	waitForEventType(t, engine.Events(), contracts.DeviceDownlinkTransmitted)

	engine.mu.RLock()
	session := *engine.sessions[device.ID]
	engine.mu.RUnlock()
	if session.CurrentDataRate != 2 || session.UnconfirmedRepetitions != 3 {
		t.Fatalf("queued MAC command was not applied: %+v", session)
	}
}

func TestLinkADRRepetitionsReuseFrameCounterAndCompleteOnce(t *testing.T) {
	clock := types.NewManualClock(0)
	device := validDevice("a1c6e32b-4f0d-4b50-9fc5-000000000001")
	configureABP(&device)
	gateway := validGateway("6f0f1f30-7dc5-4bb3-a6f5-11b2ed9a7c01")
	engine, err := New(contracts.SimulationConfig{Speed: 1}, []contracts.Device{device}, []contracts.Gateway{gateway}, types.Options{Clock: clock, EventBuffer: 256})
	if err != nil {
		t.Fatalf("create engine: %v", err)
	}
	engine.sessions[device.ID].UnconfirmedRepetitions = 3
	startEngine(t, engine)
	defer stopEngine(t, engine)

	if err := clock.Advance(10 * time.Second); err != nil {
		t.Fatalf("advance clock: %v", err)
	}
	waitForEventCounts(t, engine.Events(), map[contracts.SimulationEventType]int{
		contracts.DeviceUplinkTransmitted: 3,
		contracts.GatewayPacketReceived:   3,
	})

	snapshot := engine.Snapshot()
	if snapshot.Metrics.TotalUplinks != 1 || snapshot.Metrics.Retransmissions != 2 || snapshot.Metrics.SuccessfulUplinks != 1 {
		t.Fatalf("unexpected repetition metrics: %+v", snapshot.Metrics)
	}
	if snapshot.Devices[0].FrameCounterUp != 1 {
		t.Fatalf("repetitions must reuse one frame counter: %+v", snapshot.Devices[0])
	}
}

func TestFPendingSchedulesOneClassAPollWithoutDuplicatingPeriodicTraffic(t *testing.T) {
	clock := types.NewManualClock(0)
	device := validDevice("a1c6e32b-4f0d-4b50-9fc5-000000000001")
	configureABP(&device)
	device.PayloadConfig.UplinkInterval = 60
	gateway := validGateway("6f0f1f30-7dc5-4bb3-a6f5-11b2ed9a7c01")
	engine := newTestEngine(t, clock, device, gateway)
	defer stopEngine(t, engine)

	if _, err := engine.QueueDownlink(types.Downlink{DeviceID: device.ID, Payload: []byte("more"), FPending: true, DataRate: -1}); err != nil {
		t.Fatalf("queue FPending downlink: %v", err)
	}
	if err := clock.Advance(10 * time.Second); err != nil {
		t.Fatalf("advance clock: %v", err)
	}
	waitForEventCounts(t, engine.Events(), map[contracts.SimulationEventType]int{
		contracts.DeviceDownlinkTransmitted: 1,
		contracts.DeviceUplinkTransmitted:   2,
	})

	engine.mu.RLock()
	periodicUplinks := 0
	for _, scheduled := range engine.scheduler.Events() {
		if scheduled.Kind == types.ScheduledEventDeviceUplink && !scheduled.FPendingPoll {
			periodicUplinks++
		}
	}
	engine.mu.RUnlock()
	if periodicUplinks != 1 {
		t.Fatalf("expected one periodic uplink chain after FPending, got %d", periodicUplinks)
	}
}
