package engine

import (
	"testing"
	"time"

	contracts "github.com/Giuseppe-Compagnone/lwn-contracts/generated"
	"github.com/Giuseppe-Compagnone/lwn-engine/types"
)

func TestSchedulerReturnsEventsByTimeAndID(t *testing.T) {
	scheduler := NewScheduler()
	base := types.ScheduledEvent{
		Type:    contracts.DeviceRegistered,
		Message: "scheduled",
	}

	for _, event := range []types.ScheduledEvent{
		{ID: "b", At: 2 * time.Second, Type: base.Type, Message: base.Message},
		{ID: "a", At: 2 * time.Second, Type: base.Type, Message: base.Message},
		{ID: "c", At: time.Second, Type: base.Type, Message: base.Message},
	} {
		if err := scheduler.Schedule(event); err != nil {
			t.Fatalf("schedule event %s: %v", event.ID, err)
		}
	}

	first, ok := scheduler.PopDue(time.Second)
	if !ok || first.ID != "c" {
		t.Fatalf("expected c as first event, got %+v", first)
	}

	second, ok := scheduler.PopDue(2 * time.Second)
	if !ok || second.ID != "a" {
		t.Fatalf("expected a as second event, got %+v", second)
	}

	third, ok := scheduler.PopDue(2 * time.Second)
	if !ok || third.ID != "b" {
		t.Fatalf("expected b as third event, got %+v", third)
	}
}

func TestSchedulerCanCancelEvent(t *testing.T) {
	scheduler := NewScheduler()
	event := types.ScheduledEvent{
		ID:      "cancel-me",
		At:      time.Second,
		Type:    contracts.DeviceRegistered,
		Message: "scheduled",
	}

	if err := scheduler.Schedule(event); err != nil {
		t.Fatalf("schedule event: %v", err)
	}

	cancelled, ok := scheduler.Cancel(event.ID)
	if !ok || cancelled.ID != event.ID {
		t.Fatalf("expected event to be cancelled, got %+v", cancelled)
	}
	if scheduler.Len() != 0 {
		t.Fatalf("expected empty scheduler, got %d events", scheduler.Len())
	}
}
