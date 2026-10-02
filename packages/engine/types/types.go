package types

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	contracts "github.com/Giuseppe-Compagnone/lwn-contracts/generated"
)

// Clock represents simulation time. Implementations must be safe to use from
// the engine loop and from control operations that interrupt a wait.
type Clock interface {
	Now() time.Duration
	WaitUntil(ctx context.Context, target time.Duration, speed float64, wake <-chan struct{}) error
}

// RealClock maps simulation time to wall-clock time using the configured
// simulation speed.
type RealClock struct {
	startedAt time.Time
	speed     float64
}

func NewRealClock(speed float64) *RealClock {
	return &RealClock{startedAt: time.Now(), speed: speed}
}

func (c *RealClock) Now() time.Duration {
	return time.Duration(float64(time.Since(c.startedAt)) * c.speed)
}

func (c *RealClock) WaitUntil(
	ctx context.Context,
	target time.Duration,
	speed float64,
	wake <-chan struct{},
) error {
	if speed <= 0 {
		return fmt.Errorf("clock speed must be greater than zero")
	}

	delta := target - c.Now()
	if delta <= 0 {
		return nil
	}

	wallDuration := time.Duration(float64(delta) / speed)
	timer := time.NewTimer(wallDuration)
	defer timer.Stop()

	select {
	case <-timer.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	case <-wake:
		return nil
	}
}

// ManualClock is a deterministic clock intended for tests and step-by-step
// simulation control. Advance never moves time backwards.
type ManualClock struct {
	mu      sync.Mutex
	now     time.Duration
	changed chan struct{}
}

func NewManualClock(start time.Duration) *ManualClock {
	return &ManualClock{
		now:     start,
		changed: make(chan struct{}),
	}
}

func (c *ManualClock) Now() time.Duration {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.now
}

func (c *ManualClock) Advance(delta time.Duration) error {
	if delta < 0 {
		return fmt.Errorf("clock delta cannot be negative")
	}

	c.mu.Lock()
	c.now += delta
	close(c.changed)
	c.changed = make(chan struct{})
	c.mu.Unlock()

	return nil
}

func (c *ManualClock) Set(target time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if target < c.now {
		return fmt.Errorf("clock cannot move backwards")
	}

	c.now = target
	close(c.changed)
	c.changed = make(chan struct{})
	return nil
}

func (c *ManualClock) WaitUntil(
	ctx context.Context,
	target time.Duration,
	_ float64,
	wake <-chan struct{},
) error {
	for {
		c.mu.Lock()
		if c.now >= target {
			c.mu.Unlock()
			return nil
		}
		changed := c.changed
		c.mu.Unlock()

		select {
		case <-changed:
		case <-wake:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

type EventSink func(contracts.SimulationEvent)

type Options struct {
	Clock       Clock
	EventSink   EventSink
	EventBuffer int
}

type ScheduledEvent struct {
	ID        string
	At        time.Duration
	Type      contracts.SimulationEventType
	Message   string
	DeviceID  string
	GatewayID string
	Kind      ScheduledEventKind
}

type ScheduledEventKind string

const (
	ScheduledEventGeneric      ScheduledEventKind = "generic"
	ScheduledEventDeviceUplink ScheduledEventKind = "device-uplink"
)

type ValidationIssue struct {
	Field   string `json:"field"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

type ValidationErrors struct {
	Issues []ValidationIssue `json:"fields"`
}

func (e *ValidationErrors) Error() string {
	if e == nil || len(e.Issues) == 0 {
		return "validation failed"
	}

	parts := make([]string, 0, len(e.Issues))
	for _, issue := range e.Issues {
		parts = append(parts, fmt.Sprintf("%s: %s", issue.Field, issue.Message))
	}

	return strings.Join(parts, "; ")
}

func (e *ValidationErrors) Add(field string, code string, message string) {
	e.Issues = append(e.Issues, ValidationIssue{
		Field:   field,
		Code:    code,
		Message: message,
	})
}
