package runtime

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/rand"
	"sync"
	"time"

	contracts "github.com/Giuseppe-Compagnone/lwn-contracts/generated"
	"github.com/Giuseppe-Compagnone/lwn-engine/types"
	"github.com/google/uuid"
)

type Engine struct {
	mu sync.RWMutex

	config                contracts.SimulationConfig
	registry              *Registry
	clock                 types.Clock
	sessions              map[string]*types.DeviceSession
	radio                 map[string]*types.RadioTransmission
	rng                   *rand.Rand
	gatewayAdapterFactory types.GatewayAdapterFactory
	gatewayAdapters       map[string]types.GatewayAdapter
	gatewayRuntime        map[string]*types.GatewayRuntime

	scheduler *Scheduler
	events    chan contracts.SimulationEvent
	eventLog  []contracts.SimulationEvent
	sink      types.EventSink

	state   contracts.SimulationState
	metrics contracts.SimulationMetrics

	runCancel    context.CancelFunc
	done         chan struct{}
	wake         chan struct{}
	processingAt time.Duration
	processing   bool
	restored     bool
}

func New(
	config contracts.SimulationConfig,
	devices []contracts.Device,
	gateways []contracts.Gateway,
	options types.Options,
) (*Engine, error) {
	if err := ValidateSimulationConfig(config); err != nil {
		return nil, fmt.Errorf("validate simulation config: %w", err)
	}

	registry, err := NewRegistry(devices, gateways)
	if err != nil {
		return nil, fmt.Errorf("validate runtime registry: %w", err)
	}

	clock := options.Clock
	if clock == nil {
		clock = types.NewRealClock(config.Speed)
	}

	bufferSize := options.EventBuffer
	if bufferSize <= 0 {
		bufferSize = 256
	}

	snapshot := registry.Snapshot()
	seed := int64(1)
	if config.Seed != nil {
		seed = *config.Seed
	}
	sessions := make(map[string]*types.DeviceSession, len(devices))
	gatewayRuntime := make(map[string]*types.GatewayRuntime, len(gateways))
	for _, device := range registry.Devices() {
		sessions[device.ID] = newDeviceSession(device)
	}
	for _, gateway := range registry.Gateways() {
		state := contracts.Connected
		if gateway.Type == contracts.Real {
			state = contracts.Disconnected
		}
		gatewayRuntime[gateway.ID] = &types.GatewayRuntime{State: state}
	}
	runtime := &Engine{
		config:                config,
		registry:              registry,
		clock:                 clock,
		sessions:              sessions,
		radio:                 make(map[string]*types.RadioTransmission),
		rng:                   rand.New(rand.NewSource(seed)),
		gatewayAdapterFactory: options.GatewayAdapterFactory,
		gatewayAdapters:       make(map[string]types.GatewayAdapter),
		gatewayRuntime:        gatewayRuntime,
		scheduler:             NewScheduler(),
		events:                make(chan contracts.SimulationEvent, bufferSize),
		eventLog:              make([]contracts.SimulationEvent, 0, bufferSize),
		sink:                  options.EventSink,
		state: contracts.SimulationState{
			Status:              contracts.SimulationStatusIdle,
			Speed:               config.Speed,
			Seed:                config.Seed,
			ElapsedMilliseconds: 0,
			EventSequence:       0,
			DeviceCount:         len(snapshot.Devices),
			GatewayCount:        len(snapshot.Gateways),
		},
		metrics: contracts.SimulationMetrics{
			ActiveDevices:  int64(len(registry.ActiveDevices())),
			ActiveGateways: int64(len(registry.ActiveGateways())),
		},
		wake: make(chan struct{}, 1),
	}
	if options.Checkpoint != nil {
		if err := runtime.restoreCheckpoint(options.Checkpoint); err != nil {
			return nil, fmt.Errorf("restore simulation checkpoint: %w", err)
		}
	}
	return runtime, nil
}

func (e *Engine) restoreCheckpoint(checkpoint *types.EngineCheckpoint) error {
	if checkpoint == nil {
		return nil
	}
	if checkpoint.Config.Speed > 0 {
		e.config = checkpoint.Config
		e.state.Speed = checkpoint.Config.Speed
		e.state.Seed = checkpoint.Config.Seed
	}
	if checkpoint.State.EventSequence > 0 || checkpoint.State.ElapsedMilliseconds > 0 {
		e.state.EventSequence = checkpoint.State.EventSequence
		e.state.ElapsedMilliseconds = checkpoint.State.ElapsedMilliseconds
	}
	e.metrics = checkpoint.Metrics
	e.state.DeviceCount = len(e.registry.Devices())
	e.state.GatewayCount = len(e.registry.Gateways())
	e.metrics.ActiveDevices = int64(len(e.registry.ActiveDevices()))
	e.metrics.ActiveGateways = int64(len(e.registry.ActiveGateways()))
	e.eventLog = append([]contracts.SimulationEvent(nil), checkpoint.EventLog...)
	for deviceID, session := range checkpoint.Sessions {
		if _, ok := e.sessions[deviceID]; ok {
			copySession := session
			if session.PendingUplink != nil {
				pending := *session.PendingUplink
				pending.GatewayIDs = append([]string(nil), pending.GatewayIDs...)
				copySession.PendingUplink = &pending
			}
			if session.PendingJoinRequest != nil {
				pending := *session.PendingJoinRequest
				pending.GatewayIDs = append([]string(nil), pending.GatewayIDs...)
				copySession.PendingJoinRequest = &pending
			}
			e.sessions[deviceID] = &copySession
		}
	}
	for packetID, transmission := range checkpoint.Radio {
		copyTransmission := transmission
		copyTransmission.GatewayIDs = append([]string(nil), transmission.GatewayIDs...)
		e.radio[packetID] = &copyTransmission
	}
	for _, scheduled := range checkpoint.Scheduled {
		if err := e.scheduler.Schedule(scheduled); err != nil {
			return err
		}
	}
	if realClock, ok := e.clock.(interface{ SetElapsed(time.Duration) }); ok {
		realClock.SetElapsed(time.Duration(checkpoint.State.ElapsedMilliseconds) * time.Millisecond)
	}
	e.restored = true
	return nil
}

func (e *Engine) Start(ctx context.Context) error {
	if ctx == nil {
		return fmt.Errorf("start context cannot be nil")
	}

	e.mu.Lock()
	if e.state.Status != contracts.SimulationStatusIdle {
		status := e.state.Status
		e.mu.Unlock()
		return fmt.Errorf("%w: cannot start from %q", ErrInvalidTransition, status)
	}
	if lifecycleClock, ok := e.clock.(types.LifecycleClock); ok {
		lifecycleClock.Start()
	}

	runCtx, cancel := context.WithCancel(ctx)
	e.runCancel = cancel
	e.done = make(chan struct{})
	e.state.Status = contracts.SimulationStatusRunning
	events := []contracts.SimulationEvent{
		e.newEventLocked(
			contracts.SimulationStarted,
			"simulation started",
			"",
			"",
			"",
		),
	}
	for _, device := range e.registry.Devices() {
		events = append(events, e.newEventLocked(
			contracts.DeviceRegistered,
			"device registered in runtime registry",
			device.ID,
			"",
			"",
		))
	}
	for _, gateway := range e.registry.Gateways() {
		events = append(events, e.newEventLocked(
			contracts.GatewayRegistered,
			"gateway registered in runtime registry",
			"",
			gateway.ID,
			"",
		))
	}
	if !e.restored {
		for _, gateway := range e.registry.ActiveVirtualGateways() {
			if gateway.KeepAlive == nil {
				continue
			}
			heartbeat := types.ScheduledEvent{
				ID: uuid.NewString(), At: time.Duration(*gateway.KeepAlive) * time.Second,
				Type: contracts.GatewayHeartbeat, Message: "virtual gateway heartbeat scheduled",
				GatewayID: gateway.ID, Kind: types.ScheduledEventGatewayHeartbeat,
			}
			if err := e.scheduler.Schedule(heartbeat); err != nil {
				e.state.Status = contracts.SimulationStatusFailed
				e.mu.Unlock()
				cancel()
				return fmt.Errorf("schedule heartbeat for gateway %s: %w", gateway.ID, err)
			}
		}
		for _, device := range e.registry.ActiveDevices() {
			session := e.sessions[device.ID]
			kind := types.ScheduledEventDeviceUplink
			scheduledType := contracts.DeviceUplinkTransmitted
			scheduledMessage := "device uplink scheduled"
			scheduledLogType := contracts.DeviceUplinkScheduled
			if !session.Joined {
				kind = types.ScheduledEventJoinRequest
				scheduledType = contracts.DeviceJoinRequestTransmitted
				scheduledMessage = "device join request scheduled"
				scheduledLogType = contracts.DeviceJoinRequestScheduled
			}
			scheduled := types.ScheduledEvent{
				ID:       uuid.NewString(),
				At:       0,
				Type:     scheduledType,
				Message:  scheduledMessage,
				DeviceID: device.ID,
				Kind:     kind,
				Attempt:  1,
			}
			if err := e.scheduler.Schedule(scheduled); err != nil {
				e.state.Status = contracts.SimulationStatusFailed
				e.mu.Unlock()
				cancel()
				return fmt.Errorf("schedule initial uplink for device %s: %w", device.ID, err)
			}
			events = append(events, e.newEventLocked(
				scheduledLogType,
				scheduledMessage,
				device.ID,
				"",
				scheduled.ID,
			))
		}
	}
	e.mu.Unlock()

	for _, event := range events {
		e.publish(event)
	}
	e.startGatewayAdapters(runCtx)
	go e.run(runCtx)
	return nil
}

func (e *Engine) Pause() error {
	e.mu.Lock()
	if e.state.Status != contracts.SimulationStatusRunning {
		status := e.state.Status
		e.mu.Unlock()
		return fmt.Errorf("%w: cannot pause from %q", ErrInvalidTransition, status)
	}
	e.state.Status = contracts.SimulationStatusPaused
	if lifecycleClock, ok := e.clock.(types.LifecycleClock); ok {
		lifecycleClock.Pause()
	}
	event := e.newEventLocked(
		contracts.SimulationPaused,
		"simulation paused",
		"",
		"",
		"",
	)
	e.mu.Unlock()

	e.signalWake()
	e.publish(event)
	return nil
}

func (e *Engine) Resume() error {
	e.mu.Lock()
	if e.state.Status != contracts.SimulationStatusPaused {
		status := e.state.Status
		e.mu.Unlock()
		return fmt.Errorf("%w: cannot resume from %q", ErrInvalidTransition, status)
	}
	e.state.Status = contracts.SimulationStatusRunning
	if lifecycleClock, ok := e.clock.(types.LifecycleClock); ok {
		lifecycleClock.Resume()
	}
	event := e.newEventLocked(
		contracts.SimulationResumed,
		"simulation resumed",
		"",
		"",
		"",
	)
	e.mu.Unlock()

	e.signalWake()
	e.publish(event)
	return nil
}

func (e *Engine) Stop(ctx context.Context) error {
	if ctx == nil {
		return fmt.Errorf("stop context cannot be nil")
	}

	e.mu.Lock()
	if e.state.Status == contracts.SimulationStatusStopped {
		e.mu.Unlock()
		return nil
	}
	if e.state.Status != contracts.SimulationStatusRunning && e.state.Status != contracts.SimulationStatusPaused {
		status := e.state.Status
		e.mu.Unlock()
		return fmt.Errorf("%w: cannot stop from %q", ErrInvalidTransition, status)
	}
	e.state.Status = contracts.SimulationStatusStopping
	cancel := e.runCancel
	done := e.done
	e.mu.Unlock()

	if cancel != nil {
		cancel()
	}
	e.signalWake()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (e *Engine) Schedule(event types.ScheduledEvent) error {
	e.mu.Lock()
	status := e.state.Status
	if status != contracts.SimulationStatusRunning && status != contracts.SimulationStatusPaused {
		e.mu.Unlock()
		return fmt.Errorf("%w: cannot schedule from %q", ErrInvalidTransition, status)
	}
	if err := e.scheduler.Schedule(event); err != nil {
		e.mu.Unlock()
		return err
	}
	logEvent := e.newEventLocked(
		contracts.SchedulerEventScheduled,
		"scheduled simulation event",
		event.DeviceID,
		event.GatewayID,
		event.ID,
	)
	e.mu.Unlock()

	e.signalWake()
	e.publish(logEvent)
	return nil
}

func (e *Engine) Cancel(eventID string) error {
	e.mu.Lock()
	status := e.state.Status
	if status != contracts.SimulationStatusRunning && status != contracts.SimulationStatusPaused {
		e.mu.Unlock()
		return fmt.Errorf("%w: cannot cancel from %q", ErrInvalidTransition, status)
	}
	if _, ok := e.scheduler.Cancel(eventID); !ok {
		e.mu.Unlock()
		return ErrEventNotFound
	}
	logEvent := e.newEventLocked(
		contracts.SchedulerEventCancelled,
		"cancelled scheduled simulation event",
		"",
		"",
		eventID,
	)
	e.mu.Unlock()

	e.signalWake()
	e.publish(logEvent)
	return nil
}

func (e *Engine) Snapshot() contracts.SimulationSnapshot {
	e.mu.RLock()
	defer e.mu.RUnlock()

	snapshot := e.registry.Snapshot()
	snapshot.State = e.state
	snapshot.Metrics = e.metrics
	for index, device := range e.registry.Devices() {
		if session := e.sessions[device.ID]; session != nil {
			snapshot.Devices[index].Joined = session.Joined
			snapshot.Devices[index].FrameCounterUp = session.FrameCounterUp
			snapshot.Devices[index].FrameCounterDown = session.FrameCounterDown
			snapshot.Devices[index].PendingConfirmedUplink = session.PendingUplink != nil
			snapshot.Devices[index].LastRSSI = session.LastRSSI
			snapshot.Devices[index].LastSNR = session.LastSNR
			snapshot.Devices[index].LastAirtimeMilliseconds = session.LastAirtime.Milliseconds()
			snapshot.Devices[index].LastChannelFrequency = session.LastChannel
			snapshot.Devices[index].CurrentDataRate = session.CurrentDataRate
			snapshot.Devices[index].CurrentSpreadingFactor = session.CurrentSpreadingFactor
			snapshot.Devices[index].LastPayloadSize = session.LastPayloadSize
			snapshot.Devices[index].LastFPort = session.LastFPort
		}
	}
	for index, gateway := range e.registry.Gateways() {
		if runtime := e.gatewayRuntime[gateway.ID]; runtime != nil {
			snapshot.Gateways[index].GatewayType = gateway.Type
			snapshot.Gateways[index].GatewayEUI = gateway.GatewayEUI
			snapshot.Gateways[index].GatewayState = runtime.State
			snapshot.Gateways[index].LastHeartbeatMilliseconds = runtime.LastHeartbeat.Milliseconds()
			snapshot.Gateways[index].ConnectionAttempts = runtime.ConnectionAttempts
			snapshot.Gateways[index].LastNetworkError = runtime.LastNetworkError
			snapshot.Gateways[index].IngressPackets = runtime.IngressPackets
			snapshot.Gateways[index].EgressPackets = runtime.EgressPackets
		}
	}
	if elapsed := e.clock.Now().Milliseconds(); elapsed > snapshot.State.ElapsedMilliseconds {
		snapshot.State.ElapsedMilliseconds = elapsed
	}
	return snapshot
}

// Checkpoint returns a consistent, serializable copy of the engine runtime.
// It is safe to call from an event sink while the simulation is running.
func (e *Engine) Checkpoint() types.EngineCheckpoint {
	e.mu.RLock()
	defer e.mu.RUnlock()

	state := e.state
	if elapsed := e.clock.Now().Milliseconds(); elapsed > state.ElapsedMilliseconds {
		state.ElapsedMilliseconds = elapsed
	}
	checkpoint := types.EngineCheckpoint{
		Config:    e.config,
		State:     state,
		Metrics:   e.metrics,
		Sessions:  make(map[string]types.DeviceSession, len(e.sessions)),
		Radio:     make(map[string]types.RadioTransmission, len(e.radio)),
		Scheduled: e.scheduler.Events(),
		EventLog:  append([]contracts.SimulationEvent(nil), e.eventLog...),
	}
	for deviceID, session := range e.sessions {
		if session == nil {
			continue
		}
		copySession := *session
		if session.PendingUplink != nil {
			pending := *session.PendingUplink
			pending.GatewayIDs = append([]string(nil), pending.GatewayIDs...)
			copySession.PendingUplink = &pending
		}
		if session.PendingJoinRequest != nil {
			pending := *session.PendingJoinRequest
			pending.GatewayIDs = append([]string(nil), pending.GatewayIDs...)
			copySession.PendingJoinRequest = &pending
		}
		checkpoint.Sessions[deviceID] = copySession
	}
	for packetID, transmission := range e.radio {
		if transmission == nil {
			continue
		}
		copyTransmission := *transmission
		copyTransmission.GatewayIDs = append([]string(nil), transmission.GatewayIDs...)
		checkpoint.Radio[packetID] = copyTransmission
	}
	return checkpoint
}

func (e *Engine) EventLog() []contracts.SimulationEvent {
	e.mu.RLock()
	defer e.mu.RUnlock()

	log := make([]contracts.SimulationEvent, len(e.eventLog))
	copy(log, e.eventLog)
	return log
}

func (e *Engine) Events() <-chan contracts.SimulationEvent {
	return e.events
}

func (e *Engine) run(ctx context.Context) {
	defer func() {
		e.stopGatewayAdapters()
		e.mu.Lock()
		if e.state.Status != contracts.SimulationStatusFailed {
			e.state.Status = contracts.SimulationStatusStopped
		}
		event := e.newEventLocked(
			contracts.SimulationStopped,
			"simulation stopped",
			"",
			"",
			"",
		)

		done := e.done
		e.runCancel = nil
		e.done = nil
		e.mu.Unlock()

		e.publish(event)
		close(done)
		close(e.events)
	}()

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		e.mu.RLock()
		status := e.state.Status
		next, hasNext := e.scheduler.Peek()
		speed := e.state.Speed
		e.mu.RUnlock()

		if status == contracts.SimulationStatusPaused || !hasNext {
			select {
			case <-ctx.Done():
				return
			case <-e.wake:
			}
			continue
		}

		if err := e.clock.WaitUntil(ctx, next.At, speed, e.wake); err != nil {
			if ctx.Err() != nil {
				return
			}
			continue
		}

		e.mu.Lock()
		// Pause may have won the lock while WaitUntil was returning. Re-check
		// the lifecycle state before consuming a due event so no scheduled work
		// is processed after the pause transition has been committed.
		if e.state.Status != contracts.SimulationStatusRunning {
			e.mu.Unlock()
			continue
		}
		due, ok := e.scheduler.PopDue(e.clock.Now())
		if !ok {
			e.mu.Unlock()
			continue
		}
		if elapsed := e.clock.Now().Milliseconds(); elapsed > e.state.ElapsedMilliseconds {
			e.state.ElapsedMilliseconds = elapsed
		}
		e.processingAt = due.At
		e.processing = true
		events := e.processScheduledEventLocked(due)
		e.processing = false
		e.mu.Unlock()
		for _, event := range events {
			e.publish(event)
		}
	}
}

func locationCoordinates(device contracts.Device) (float64, float64) {
	return float64(pointerValue(device.LocationConfig.Latitude)), float64(pointerValue(device.LocationConfig.Longitude))
}

func gatewayCoordinates(gateway contracts.Gateway) (float64, float64) {
	return float64(pointerValue(gateway.Latitude)), float64(pointerValue(gateway.Longitude))
}

func pointerValue[T any](value *T) T {
	if value == nil {
		var zero T
		return zero
	}
	return *value
}

func int32Pointer(value int) *int32 {
	converted := int32(value)
	return &converted
}

func newDeviceSession(device contracts.Device) *types.DeviceSession {
	dataRate := defaultUplinkDataRate(device.LocationConfig.Region)
	if configured := device.AdvancedConfig.UplinkDataRate; configured != nil {
		dataRate = clampDataRate(*configured)
	}
	session := &types.DeviceSession{
		Joined:                 device.Activation == contracts.ABP,
		FrameCounterUp:         initialFrameCounter(device.FrameConfig.FCntUp),
		FrameCounterDown:       initialFrameCounter(device.FrameConfig.FCntDown),
		CurrentDataRate:        dataRate,
		CurrentSpreadingFactor: spreadingFactorForDataRate(dataRate),
		LastFPort:              device.FrameConfig.FPort,
	}
	if device.ABPConfig != nil {
		session.DeviceAddress = device.ABPConfig.DevAddr
		session.SecurityFingerprint = securityFingerprint(device.DevEUI, device.ABPConfig.NwkSKey, device.ABPConfig.AppSKey)
	}
	if device.OOTAConfig != nil {
		session.JoinEUI = device.OOTAConfig.JoinEUI
		session.SecurityFingerprint = securityFingerprint(device.DevEUI, device.OOTAConfig.JoinEUI, device.OOTAConfig.AppKey)
	}
	return session
}

func securityFingerprint(values ...string) string {
	hash := sha256.New()
	for _, value := range values {
		_, _ = hash.Write([]byte(value))
	}
	return hex.EncodeToString(hash.Sum(nil))
}

func joinedDeviceAddress(session *types.DeviceSession) string {
	if session == nil || session.SecurityFingerprint == "" {
		return ""
	}
	return session.SecurityFingerprint[:8]
}

func initialFrameCounter(value **int) int64 {
	if value == nil || *value == nil {
		return 0
	}
	return int64(**value)
}

func (e *Engine) newEventLocked(
	eventType contracts.SimulationEventType,
	message string,
	deviceID string,
	gatewayID string,
	scheduledEventID string,
) contracts.SimulationEvent {
	e.state.EventSequence++
	event := contracts.SimulationEvent{
		ID:                    uuid.NewString(),
		Sequence:              e.state.EventSequence,
		Type:                  eventType,
		TimestampMilliseconds: e.eventTimestampLocked().Milliseconds(),
		Message:               message,
	}
	if deviceID != "" {
		event.DeviceID = &deviceID
	}
	if gatewayID != "" {
		event.GatewayID = &gatewayID
	}
	if scheduledEventID != "" {
		event.ScheduledEventID = &scheduledEventID
	}
	e.eventLog = append(e.eventLog, event)
	return event
}

func (e *Engine) eventTimestampLocked() time.Duration {
	if e.processing {
		return e.processingAt
	}
	return e.clock.Now()
}

func (e *Engine) newPacketEventLocked(
	eventType contracts.SimulationEventType,
	message string,
	deviceID string,
	gatewayID string,
	scheduledEventID string,
	packetID string,
) contracts.SimulationEvent {
	event := e.newEventLocked(eventType, message, deviceID, gatewayID, scheduledEventID)
	event.PacketID = &packetID
	e.eventLog[len(e.eventLog)-1] = event
	return event
}

func (e *Engine) publish(event contracts.SimulationEvent) {
	if e.sink != nil {
		e.sink(event)
	}

	select {
	case e.events <- event:
	default:
		// EventLog is authoritative. The channel is a non-blocking live view so
		// a slow consumer cannot stop the simulation loop.
	}
}

func (e *Engine) signalWake() {
	select {
	case e.wake <- struct{}{}:
	default:
	}
}
