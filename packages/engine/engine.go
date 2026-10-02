package engine

import (
	"context"
	"fmt"
	"sync"

	contracts "github.com/Giuseppe-Compagnone/lwn-contracts/generated"
	"github.com/Giuseppe-Compagnone/lwn-engine/types"
	"github.com/google/uuid"
)

type Engine struct {
	mu sync.RWMutex

	config   contracts.SimulationConfig
	registry *Registry
	clock    types.Clock

	scheduler *Scheduler
	events    chan contracts.SimulationEvent
	eventLog  []contracts.SimulationEvent
	sink      types.EventSink

	state contracts.SimulationState

	runCancel context.CancelFunc
	done      chan struct{}
	wake      chan struct{}
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
	return &Engine{
		config:    config,
		registry:  registry,
		clock:     clock,
		scheduler: NewScheduler(),
		events:    make(chan contracts.SimulationEvent, bufferSize),
		eventLog:  make([]contracts.SimulationEvent, 0, bufferSize),
		sink:      options.EventSink,
		state: contracts.SimulationState{
			Status:              contracts.SimulationStatusIdle,
			Speed:               config.Speed,
			Seed:                config.Seed,
			ElapsedMilliseconds: 0,
			EventSequence:       0,
			DeviceCount:         len(snapshot.Devices),
			GatewayCount:        len(snapshot.Gateways),
		},
		wake: make(chan struct{}, 1),
	}, nil
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
	e.mu.Unlock()

	for _, event := range events {
		e.publish(event)
	}
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
	if elapsed := e.clock.Now().Milliseconds(); elapsed > snapshot.State.ElapsedMilliseconds {
		snapshot.State.ElapsedMilliseconds = elapsed
	}
	return snapshot
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
		due, ok := e.scheduler.PopDue(e.clock.Now())
		if !ok {
			e.mu.Unlock()
			continue
		}
		if elapsed := e.clock.Now().Milliseconds(); elapsed > e.state.ElapsedMilliseconds {
			e.state.ElapsedMilliseconds = elapsed
		}
		event := e.newEventLocked(due.Type, due.Message, due.DeviceID, due.GatewayID, due.ID)
		e.mu.Unlock()
		e.publish(event)
	}
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
		TimestampMilliseconds: e.clock.Now().Milliseconds(),
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
