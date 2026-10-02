package engine

import (
	"context"
	"testing"
	"time"

	contracts "github.com/Giuseppe-Compagnone/lwn-contracts/generated"
	"github.com/Giuseppe-Compagnone/lwn-engine/types"
)

func TestEngineLifecycleAndScheduledEvents(t *testing.T) {
	clock := types.NewManualClock(0)
	events := make(chan contracts.SimulationEvent, 32)
	engine, err := New(
		contracts.SimulationConfig{Speed: 1, Seed: int64Ptr(7)},
		[]contracts.Device{validDevice("a1c6e32b-4f0d-4b50-9fc5-000000000001")},
		[]contracts.Gateway{validGateway("6f0f1f30-7dc5-4bb3-a6f5-11b2ed9a7c01")},
		types.Options{
			Clock:       clock,
			EventSink:   func(event contracts.SimulationEvent) { events <- event },
			EventBuffer: 64,
		},
	)
	if err != nil {
		t.Fatalf("create engine: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := engine.Start(ctx); err != nil {
		t.Fatalf("start engine: %v", err)
	}

	if got := engine.Snapshot().State.Status; got != contracts.SimulationStatusRunning {
		t.Fatalf("expected running state, got %q", got)
	}
	for range 3 {
		readEvent(t, events)
	}

	if err := engine.Pause(); err != nil {
		t.Fatalf("pause engine: %v", err)
	}
	if got := engine.Snapshot().State.Status; got != contracts.SimulationStatusPaused {
		t.Fatalf("expected paused state, got %q", got)
	}
	readEvent(t, events)

	if err := engine.Resume(); err != nil {
		t.Fatalf("resume engine: %v", err)
	}
	readEvent(t, events)

	if err := engine.Schedule(types.ScheduledEvent{
		ID:      "event-1",
		At:      100 * time.Millisecond,
		Type:    contracts.DeviceRegistered,
		Message: "scheduled device event",
	}); err != nil {
		t.Fatalf("schedule event: %v", err)
	}
	readEvent(t, events)

	if err := clock.Advance(100 * time.Millisecond); err != nil {
		t.Fatalf("advance clock: %v", err)
	}
	for {
		event := readEvent(t, events)
		if event.Message == "scheduled device event" {
			if event.ScheduledEventID == nil || *event.ScheduledEventID != "event-1" {
				t.Fatalf("expected scheduled event reference, got %+v", event.ScheduledEventID)
			}
			break
		}
	}

	stopContext, stopCancel := context.WithTimeout(context.Background(), time.Second)
	defer stopCancel()
	if err := engine.Stop(stopContext); err != nil {
		t.Fatalf("stop engine: %v", err)
	}
	readEvent(t, events)

	if got := engine.Snapshot().State.Status; got != contracts.SimulationStatusStopped {
		t.Fatalf("expected stopped state, got %q", got)
	}
	if len(engine.EventLog()) != int(engine.Snapshot().State.EventSequence) {
		t.Fatal("event log and state sequence are inconsistent")
	}
}

func TestEngineRejectsInvalidLifecycleTransitions(t *testing.T) {
	engine, err := New(
		contracts.SimulationConfig{Speed: 1},
		nil,
		nil,
		types.Options{Clock: types.NewManualClock(0)},
	)
	if err != nil {
		t.Fatalf("create engine: %v", err)
	}

	if err := engine.Pause(); err == nil {
		t.Fatal("expected pause before start to fail")
	}
	if err := engine.Resume(); err == nil {
		t.Fatal("expected resume before start to fail")
	}
}

func readEvent(t *testing.T, events <-chan contracts.SimulationEvent) contracts.SimulationEvent {
	t.Helper()
	select {
	case event := <-events:
		return event
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for engine event")
		return contracts.SimulationEvent{}
	}
}
