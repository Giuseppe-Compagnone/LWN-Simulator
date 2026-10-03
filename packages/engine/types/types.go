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

// LifecycleClock lets the engine freeze and resume wall-clock based time.
// ManualClock deliberately does not implement it: tests control that clock by
// advancing it explicitly.
type LifecycleClock interface {
	Clock
	Start()
	Pause()
	Resume()
}

// RealClock maps simulation time to wall-clock time using the configured
// simulation speed.
type RealClock struct {
	mu        sync.Mutex
	startedAt time.Time
	elapsed   time.Duration
	speed     float64
	started   bool
	paused    bool
}

func NewRealClock(speed float64) *RealClock {
	return &RealClock{speed: speed}
}

func (c *RealClock) Now() time.Duration {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.nowLocked(time.Now())
}

func (c *RealClock) Start() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.startedAt = time.Now()
	c.started = true
	c.paused = false
}

func (c *RealClock) SetElapsed(elapsed time.Duration) {
	if elapsed < 0 {
		elapsed = 0
	}
	c.mu.Lock()
	c.elapsed = elapsed
	c.mu.Unlock()
}

func (c *RealClock) Pause() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.started || c.paused {
		return
	}
	c.elapsed = c.nowLocked(time.Now())
	c.paused = true
}

func (c *RealClock) Resume() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.started || !c.paused {
		return
	}
	c.startedAt = time.Now()
	c.paused = false
}

func (c *RealClock) nowLocked(now time.Time) time.Duration {
	if !c.started || c.paused {
		return c.elapsed
	}
	return c.elapsed + time.Duration(float64(now.Sub(c.startedAt))*c.speed)
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

type GatewayPacket struct {
	GatewayID       string
	Payload         []byte
	Frequency       int64
	Bandwidth       int64
	SpreadingFactor int
	Power           int
	DataRate        string
	ReceivedAt      time.Time
}

type GatewayAdapterEvent struct {
	GatewayID string
	State     contracts.GatewayConnectionState
	Error     string
	Timeout   bool
	Heartbeat bool
	At        time.Time
}

type GatewayAdapter interface {
	Start(context.Context) error
	Send(context.Context, GatewayPacket) error
	Packets() <-chan GatewayPacket
	Events() <-chan GatewayAdapterEvent
	Close() error
}

type GatewayAdapterFactory interface {
	NewGatewayAdapter(contracts.Gateway) (GatewayAdapter, error)
}

type GatewayAdapterFactoryCloser interface {
	Close() error
}

type GatewayRuntime struct {
	State              contracts.GatewayConnectionState
	LastHeartbeat      time.Duration
	ConnectionAttempts int64
	LastNetworkError   string
	IngressPackets     int64
	EgressPackets      int64
}

type Options struct {
	Clock                 Clock
	EventSink             EventSink
	EventBuffer           int
	GatewayAdapterFactory GatewayAdapterFactory
	Checkpoint            *EngineCheckpoint
}

// EngineCheckpoint is the durable simulation state. It intentionally keeps
// the contract snapshot and the engine runtime state together so a backend
// restart can continue from the same simulation timeline.
type EngineCheckpoint struct {
	Config    contracts.SimulationConfig
	State     contracts.SimulationState
	Metrics   contracts.SimulationMetrics
	Sessions  map[string]DeviceSession
	Radio     map[string]RadioTransmission
	Scheduled []ScheduledEvent
	EventLog  []contracts.SimulationEvent
}

type ScheduledEvent struct {
	ID               string
	At               time.Duration
	Type             contracts.SimulationEventType
	Message          string
	DeviceID         string
	GatewayID        string
	Kind             ScheduledEventKind
	Attempt          int
	PacketID         string
	FrameCounter     int64
	Confirmed        bool
	Window           contracts.SimulationRxWindow
	WindowDuration   time.Duration
	DataRate         int
	ChannelFrequency int64
	WindowBaseAt     time.Duration
	FragmentIndex    int
	FragmentCount    int
}

type ScheduledEventKind string

const (
	ScheduledEventGeneric          ScheduledEventKind = "generic"
	ScheduledEventDeviceUplink     ScheduledEventKind = "device-uplink"
	ScheduledEventJoinRequest      ScheduledEventKind = "join-request"
	ScheduledEventRX1Window        ScheduledEventKind = "rx1-window"
	ScheduledEventRX2Window        ScheduledEventKind = "rx2-window"
	ScheduledEventRadioComplete    ScheduledEventKind = "radio-transmission-complete"
	ScheduledEventGatewayHeartbeat ScheduledEventKind = "virtual-gateway-heartbeat"
)

type DeviceSession struct {
	Joined                 bool
	FrameCounterUp         int64
	FrameCounterDown       int64
	CurrentDataRate        int
	CurrentSpreadingFactor int
	DeviceAddress          string
	JoinEUI                string
	SecurityFingerprint    string
	PendingUplink          *PendingUplink
	PendingJoinRequest     *PendingJoinRequest
	LastRSSI               float64
	LastSNR                float64
	LastAirtime            time.Duration
	LastChannel            int64
	LastPayloadSize        int
	LastFPort              int
}

type PendingUplink struct {
	PacketID       string
	FrameCounter   int64
	Attempt        int
	Confirmed      bool
	GatewayIDs     []string
	TransmissionAt time.Duration
	FPort          int
	PayloadSize    int
	DataRate       int
	FragmentIndex  int
	FragmentCount  int
}

type PendingJoinRequest struct {
	PacketID       string
	JoinEUI        string
	GatewayIDs     []string
	TransmissionAt time.Duration
}

type RadioChannel struct {
	Frequency       int64
	Bandwidth       int64
	SpreadingFactor int
	DataRate        int
}

type RadioTransmission struct {
	PacketID         string
	DeviceID         string
	FrameCounter     int64
	Attempt          int
	Confirmed        bool
	StartAt          time.Duration
	EndAt            time.Duration
	ChannelFrequency int64
	Bandwidth        int64
	SpreadingFactor  int
	DataRate         int
	FPort            int
	PayloadSize      int
	FragmentIndex    int
	FragmentCount    int
	Airtime          time.Duration
	RSSI             float64
	SNR              float64
	GatewayIDs       []string
	Collision        bool
}

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
