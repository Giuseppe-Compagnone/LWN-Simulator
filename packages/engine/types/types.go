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

// SetSpeed changes the wall-clock multiplier while preserving the simulation
// time already accumulated. The current elapsed value is captured before the
// new multiplier is applied so changing speed never freezes or jumps the
// simulation clock.
func (c *RealClock) SetSpeed(speed float64) error {
	if speed <= 0 {
		return fmt.Errorf("clock speed must be greater than zero")
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	if c.started && !c.paused {
		now := time.Now()
		c.elapsed = c.nowLocked(now)
		c.startedAt = now
	}
	c.speed = speed
	return nil
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
	Kind            GatewayPacketKind
	Payload         []byte
	Frequency       int64
	Bandwidth       int64
	SpreadingFactor int
	RSSI            float64
	SNR             float64
	Power           int
	DataRate        string
	// TransmitAt requests a timed downlink. A zero value keeps the existing
	// immediate-transmission behavior used by ordinary gateway packets.
	TransmitAt time.Time
	ReceivedAt time.Time
}

type GatewayPacketKind string

const (
	GatewayPacketGeneric        GatewayPacketKind = "generic"
	GatewayPacketUplink         GatewayPacketKind = "uplink"
	GatewayPacketClassBBeacon   GatewayPacketKind = "class-b-beacon"
	GatewayPacketClassBDownlink GatewayPacketKind = "class-b-downlink"
	GatewayPacketDownlink       GatewayPacketKind = "downlink"
)

type GatewayAdapterEvent struct {
	GatewayID string
	State     contracts.GatewayConnectionState
	Error     string
	Timeout   bool
	Heartbeat bool
	At        time.Time
	Packet    *GatewayPacket
}

type GatewayAdapter interface {
	Start(context.Context) error
	Send(context.Context, GatewayPacket) error
	Packets() <-chan GatewayPacket
	Events() <-chan GatewayAdapterEvent
	Close() error
}

// GatewayUplinkAdapter is implemented by transports that can expose the
// simulator as a gateway to an external network server. Downlinks continue to
// use GatewayAdapter.Send; uplinks use this separate method so the transport
// can encode the Semtech PUSH_DATA envelope correctly.
type GatewayUplinkAdapter interface {
	SendUplink(context.Context, GatewayPacket) error
}

// VirtualGatewayAdapterFactory identifies factories that can transport
// in-process virtual gateways to an external Gateway Bridge.
type VirtualGatewayAdapterFactory interface {
	SupportsVirtualGateways() bool
}

type GatewayAdapterFactory interface {
	NewGatewayAdapter(contracts.Gateway) (GatewayAdapter, error)
}

type GatewayAdapterFactoryCloser interface {
	Close() error
}

// GatewayBridgeConfigurator allows the backend to apply the bridge settings
// selected for a simulation before the engine starts its gateway adapters.
// Implementations must not start or stop transports while configuring them.
type GatewayBridgeConfigurator interface {
	ConfigureGatewayBridge(contracts.GatewayBridgeConfig) error
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
	RunID          string
	Config         contracts.SimulationConfig
	Devices        []contracts.Device
	Gateways       []contracts.Gateway
	State          contracts.SimulationState
	Metrics        contracts.SimulationMetrics
	Sessions       map[string]DeviceSession
	Radio          map[string]RadioTransmission
	Scheduled      []ScheduledEvent
	GatewayPackets []GatewayPacket
	EventLog       []contracts.SimulationEvent
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
	FPendingPoll     bool
}

type ScheduledEventKind string

const (
	ScheduledEventGeneric             ScheduledEventKind = "generic"
	ScheduledEventDeviceUplink        ScheduledEventKind = "device-uplink"
	ScheduledEventJoinRequest         ScheduledEventKind = "join-request"
	ScheduledEventRX1Window           ScheduledEventKind = "rx1-window"
	ScheduledEventRX2Window           ScheduledEventKind = "rx2-window"
	ScheduledEventRadioComplete       ScheduledEventKind = "radio-transmission-complete"
	ScheduledEventGatewayHeartbeat    ScheduledEventKind = "virtual-gateway-heartbeat"
	ScheduledEventGatewayBeacon       ScheduledEventKind = "gateway-beacon"
	ScheduledEventClassBBeaconTimeout ScheduledEventKind = "class-b-beacon-timeout"
	ScheduledEventClassBPingSlot      ScheduledEventKind = "class-b-ping-slot"
	ScheduledEventClassCDownlink      ScheduledEventKind = "class-c-downlink"
)

// Class B follows the LoRaWAN beacon cadence. The one-second ping-slot period
// is the default periodicity used by the simulator when a device does not
// expose a network-specific ping-slot periodicity.
const (
	ClassBBeaconPeriod   = 128 * time.Second
	ClassBPingSlotPeriod = time.Second
)

type DeviceSession struct {
	Joined                    bool
	FrameCounterUp            int64
	LastExternalFrameCounter  uint32
	FrameCounterDown          int64
	CurrentDataRate           int
	CurrentSpreadingFactor    int
	DeviceAddress             string
	JoinEUI                   string
	SecurityFingerprint       string
	PendingUplink             *PendingUplink
	PendingJoinRequest        *PendingJoinRequest
	LastRSSI                  float64
	LastSNR                   float64
	LastAirtime               time.Duration
	LastChannel               int64
	LastPayloadSize           int
	LastFPort                 int
	ClassBSynchronized        bool
	LastBeaconAt              time.Duration
	NextPingSlotAt            time.Duration
	ClassBMissedBeacons       int64
	ClassBGatewayID           string
	ClassBNextPingEventID     string
	ClassBBeaconTimeoutID     string
	PendingClassBDownlinks    []ClassBDownlink
	PendingClassADownlinks    []Downlink
	PendingClassCDownlinks    []Downlink
	ClassCNextDownlinkEventID string
	NwkSKey                   []byte
	AppSKey                   []byte
	DevNonce                  uint16
	CurrentTxPower            int
	UnconfirmedRepetitions    int
	MaximumDutyCycle          float64
	RX1DataRateOffset         int
	RX2DataRate               int
	RX2Frequency              int64
	ReceiveDelay              time.Duration
	UplinkDwellTime           bool
	DownlinkDwellTime         bool
	MaximumEIRP               int
	PingSlotDataRate          int
	PingSlotFrequency         int64
	PingSlotPeriodicity       int
	BeaconFrequency           int64
	AdditionalChannels        []RadioChannel
}

// Downlink is queued for delivery according to the device class. An empty ID
// is replaced by the engine with a generated identifier.
type Downlink struct {
	ID          string
	DeviceID    string
	Payload     []byte
	FPort       int
	DataRate    int
	Confirmed   bool
	FPending    bool
	ACK         bool
	MACCommands []MACCommand
}

type MACCommandType string

const (
	MACLinkCheckAns       MACCommandType = "link-check-ans"
	MACLinkADRReq         MACCommandType = "link-adr-req"
	MACDutyCycleReq       MACCommandType = "duty-cycle-req"
	MACRXParamSetupReq    MACCommandType = "rx-param-setup-req"
	MACDevStatusReq       MACCommandType = "dev-status-req"
	MACNewChannelReq      MACCommandType = "new-channel-req"
	MACRXTimingSetupReq   MACCommandType = "rx-timing-setup-req"
	MACTXParamSetupReq    MACCommandType = "tx-param-setup-req"
	MACDLChannelReq       MACCommandType = "dl-channel-req"
	MACDeviceTimeAns      MACCommandType = "device-time-ans"
	MACPingSlotInfoAns    MACCommandType = "ping-slot-info-ans"
	MACPingSlotChannelReq MACCommandType = "ping-slot-channel-req"
	MACBeaconFreqReq      MACCommandType = "beacon-freq-req"
)

// MACCommand contains the semantic values of a LoRaWAN 1.0.x MAC command.
// Only fields used by the selected command are considered. Pointer fields
// distinguish an explicit zero from an omitted value.
type MACCommand struct {
	Type                 MACCommandType
	DataRate             *int
	TxPower              *int
	NbTrans              *int
	ChannelMask          *uint16
	ChannelMaskControl   *int
	MaxDutyCycleExponent *int
	RX1DataRateOffset    *int
	Frequency            *int64
	ChannelIndex         *int
	MinimumDataRate      *int
	MaximumDataRate      *int
	Delay                *time.Duration
	UplinkDwellTime      *bool
	DownlinkDwellTime    *bool
	MaximumEIRP          *int
	Margin               *int
	GatewayCount         *int
	BatteryLevel         *int
	PingSlotPeriodicity  *int
	DeviceTime           *time.Time
}

// ClassBDownlink is kept as a semantic alias for callers that use the
// Class-B-specific API.
type ClassBDownlink = Downlink

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
	AnySuccessful  bool
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

// DataRateProfile describes the radio modulation used by a regional data
// rate. Keeping this in the public types package lets backend and frontend
// adapters expose the same engine vocabulary without depending on runtime
// internals.
type DataRateProfile struct {
	DataRate        int
	SpreadingFactor int
	Bandwidth       int64
	MaximumPayload  int
}

type RegionalChannelGroup struct {
	InitialFrequency int64
	FrequencyStep    int64
	ChannelCount     int
	MinimumDataRate  int
	MaximumDataRate  int
}

type RegionalPlan struct {
	Region            contracts.DeviceRegion
	DefaultDataRate   int
	MinimumDataRate   int
	MaximumDataRate   int
	RX2Frequency      int64
	RX2DataRate       int
	BeaconFrequency   int64
	BeaconDataRate    int
	PingSlotFrequency int64
	PingSlotDataRate  int
	UplinkChannels    []RegionalChannelGroup
	DataRates         map[int]DataRateProfile
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
	FPendingPoll     bool
	Payload          []byte
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
