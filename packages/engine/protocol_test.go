package engine

import (
	"testing"
	"time"

	contracts "github.com/Giuseppe-Compagnone/lwn-contracts/generated"
	"github.com/Giuseppe-Compagnone/lwn-engine/types"
)

func TestPhaseThreeOTAAJoinOpensRX1AndStartsUplinks(t *testing.T) {
	clock := types.NewManualClock(0)
	device := validDevice("a1c6e32b-4f0d-4b50-9fc5-000000000001")
	device.PayloadConfig.UplinkInterval = 10
	gateway := validGateway("6f0f1f30-7dc5-4bb3-a6f5-11b2ed9a7c01")

	engine := newTestEngine(t, clock, device, gateway)
	defer stopEngine(t, engine)

	waitForEventType(t, engine.Events(), contracts.DeviceJoinRequestTransmitted)
	advanceAndWait(t, clock, time.Second, engine.Events(), contracts.JoinRequestReceived)
	joinAccept := waitForEventType(t, engine.Events(), contracts.JoinAcceptReceived)
	if joinAccept.RxWindow == nil || *joinAccept.RxWindow != contracts.RX1 {
		t.Fatalf("expected join accept in RX1, got %+v", joinAccept.RxWindow)
	}
	if !engine.Snapshot().Devices[0].Joined {
		t.Fatal("expected OTAA device to be joined")
	}

	if err := clock.Advance(10 * time.Second); err != nil {
		t.Fatalf("advance to first post-join uplink: %v", err)
	}
	uplink := waitForEventType(t, engine.Events(), contracts.DeviceUplinkTransmitted)
	if uplink.FrameCounter == nil || *uplink.FrameCounter != 1 {
		t.Fatalf("expected first uplink frame counter to be 1, got %+v", uplink.FrameCounter)
	}

	metrics := engine.Snapshot().Metrics
	if metrics.JoinRequests != 1 || metrics.JoinAccepts != 1 || metrics.TotalUplinks != 1 {
		t.Fatalf("unexpected OTAA metrics: %+v", metrics)
	}
}

func TestPhaseThreeConfirmedUplinkReceivesACKInRX1(t *testing.T) {
	clock := types.NewManualClock(0)
	device := validDevice("a1c6e32b-4f0d-4b50-9fc5-000000000001")
	configureABP(&device)
	device.PayloadConfig.MType = contracts.ConfirmedDataUp
	device.FrameConfig.Retransmission = intPtr(2)
	gateway := validGateway("6f0f1f30-7dc5-4bb3-a6f5-11b2ed9a7c01")

	engine := newTestEngine(t, clock, device, gateway)
	defer stopEngine(t, engine)

	uplink := waitForEventType(t, engine.Events(), contracts.DeviceUplinkTransmitted)
	if uplink.Confirmed == nil || !*uplink.Confirmed {
		t.Fatal("expected confirmed uplink metadata")
	}
	if !engine.Snapshot().Devices[0].PendingConfirmedUplink {
		t.Fatal("expected confirmed uplink to remain pending before ACK")
	}

	if err := clock.Advance(time.Second); err != nil {
		t.Fatalf("advance to RX1: %v", err)
	}
	ack := waitForEventType(t, engine.Events(), contracts.UplinkACKReceived)
	if ack.RxWindow == nil || *ack.RxWindow != contracts.RX1 {
		t.Fatalf("expected ACK in RX1, got %+v", ack.RxWindow)
	}

	snapshot := engine.Snapshot()
	if snapshot.Devices[0].PendingConfirmedUplink {
		t.Fatal("expected confirmed uplink to be cleared after ACK")
	}
	if snapshot.Devices[0].FrameCounterUp != 1 || snapshot.Devices[0].FrameCounterDown != 1 {
		t.Fatalf("unexpected frame counters: %+v", snapshot.Devices[0])
	}
	if snapshot.Metrics.AcknowledgedUplinks != 1 || snapshot.Metrics.Retransmissions != 0 {
		t.Fatalf("unexpected ACK metrics: %+v", snapshot.Metrics)
	}
}

func TestPhaseThreeConfirmedUplinkRetriesAfterRX2Timeout(t *testing.T) {
	clock := types.NewManualClock(0)
	device := validDevice("a1c6e32b-4f0d-4b50-9fc5-000000000001")
	configureABP(&device)
	device.PayloadConfig.MType = contracts.ConfirmedDataUp
	device.FrameConfig.Retransmission = intPtr(2)
	device.RX2Config.Duration = intPtr(1000)

	engine := newTestEngine(t, clock, device)
	defer stopEngine(t, engine)

	waitForEventType(t, engine.Events(), contracts.DeviceUplinkTransmitted)
	for attempt := 1; attempt <= 2; attempt++ {
		advanceAndWait(t, clock, time.Second, engine.Events(), contracts.RX1WindowOpened)
		advanceAndWait(t, clock, time.Second, engine.Events(), contracts.RX2WindowOpened)
		advanceAndWait(t, clock, time.Second, engine.Events(), contracts.UplinkRetryScheduled)
		transmitted := waitForEventType(t, engine.Events(), contracts.DeviceUplinkTransmitted)
		if transmitted.RetryAttempt == nil || int(*transmitted.RetryAttempt) != attempt+1 {
			t.Fatalf("expected retry attempt %d, got %+v", attempt+1, transmitted.RetryAttempt)
		}
	}

	advanceAndWait(t, clock, time.Second, engine.Events(), contracts.RX1WindowOpened)
	advanceAndWait(t, clock, time.Second, engine.Events(), contracts.RX2WindowOpened)
	waitForEventType(t, engine.Events(), contracts.PacketDropped)

	metrics := engine.Snapshot().Metrics
	if metrics.TotalUplinks != 1 || metrics.Retransmissions != 2 || metrics.DroppedUplinks != 1 {
		t.Fatalf("unexpected retransmission metrics: %+v", metrics)
	}
	if metrics.FrameCounterErrors != 0 || metrics.SuccessfulUplinks != 0 {
		t.Fatalf("unexpected frame counter/success metrics: %+v", metrics)
	}
}

func newTestEngine(t *testing.T, clock *types.ManualClock, device contracts.Device, gateways ...contracts.Gateway) *Engine {
	t.Helper()
	engine, err := New(contracts.SimulationConfig{Speed: 1}, []contracts.Device{device}, gateways, types.Options{Clock: clock, EventBuffer: 256})
	if err != nil {
		t.Fatalf("create engine: %v", err)
	}
	startEngine(t, engine)
	return engine
}

func advanceAndWait(t *testing.T, clock *types.ManualClock, delta time.Duration, events <-chan contracts.SimulationEvent, eventType contracts.SimulationEventType) contracts.SimulationEvent {
	t.Helper()
	if err := clock.Advance(delta); err != nil {
		t.Fatalf("advance clock: %v", err)
	}
	return waitForEventType(t, events, eventType)
}
