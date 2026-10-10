package types

import (
	"context"
	"time"

	contracts "github.com/Giuseppe-Compagnone/lwn-contracts/generated"
)

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
