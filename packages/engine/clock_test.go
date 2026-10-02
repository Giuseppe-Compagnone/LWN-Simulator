package engine

import (
	"context"
	"testing"
	"time"

	"github.com/Giuseppe-Compagnone/lwn-engine/types"
)

func TestManualClockDoesNotMoveBackwards(t *testing.T) {
	clock := types.NewManualClock(10 * time.Millisecond)

	if err := clock.Advance(5 * time.Millisecond); err != nil {
		t.Fatalf("advance clock: %v", err)
	}
	if got := clock.Now(); got != 15*time.Millisecond {
		t.Fatalf("expected 15ms, got %s", got)
	}

	if err := clock.Set(14 * time.Millisecond); err == nil {
		t.Fatal("expected backwards clock movement to fail")
	}
}

func TestManualClockWaitsUntilTarget(t *testing.T) {
	clock := types.NewManualClock(0)
	done := make(chan error, 1)

	go func() {
		done <- clock.WaitUntil(context.Background(), 100*time.Millisecond, 1, make(chan struct{}))
	}()

	select {
	case err := <-done:
		t.Fatalf("clock returned before target: %v", err)
	case <-time.After(10 * time.Millisecond):
	}

	if err := clock.Advance(100 * time.Millisecond); err != nil {
		t.Fatalf("advance clock: %v", err)
	}

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("wait until target: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("clock did not reach target")
	}
}
