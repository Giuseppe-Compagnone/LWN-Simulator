package services

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"

	contracts "github.com/Giuseppe-Compagnone/lwn-contracts/generated"
	"github.com/Giuseppe-Compagnone/lwn-engine"
	"github.com/Giuseppe-Compagnone/lwn-engine/types"
	"lwn-simulator-backend/internal/apperrors"
)

var ErrSimulationNotStarted = errors.New("simulation has not been started")

type SimulationDeviceSource interface {
	GetDevices(contracts.GetDevicesRequest) (contracts.GetDevicesResponse, error)
}

type SimulationGatewaySource interface {
	GetGateways(contracts.GetGatewaysRequest) (contracts.GetGatewaysResponse, error)
}

type SimulationUpdate struct {
	Snapshot contracts.SimulationSnapshot
	Event    contracts.SimulationEvent
}

type SimulationSubscription struct {
	Updates <-chan SimulationUpdate
	Close   func()
}

type SimulationService struct {
	devices         SimulationDeviceSource
	gateways        SimulationGatewaySource
	options         types.Options
	checkpointStore SimulationCheckpointStore

	operationMu sync.Mutex
	mu          sync.RWMutex
	engine      *engine.Engine
	starting    bool
	nextID      uint64
	subscribers map[uint64]chan SimulationUpdate
}

func NewSimulationService(
	devices SimulationDeviceSource,
	gateways SimulationGatewaySource,
	options types.Options,
	stores ...SimulationCheckpointStore,
) *SimulationService {
	var checkpointStore SimulationCheckpointStore
	if len(stores) > 0 {
		checkpointStore = stores[0]
	}
	return &SimulationService{
		devices:         devices,
		gateways:        gateways,
		options:         options,
		checkpointStore: checkpointStore,
		subscribers:     make(map[uint64]chan SimulationUpdate),
	}
}

// RestorePersisted resumes a simulation that was active when the backend
// stopped. A deliberately stopped or failed simulation is not restarted.
func (s *SimulationService) RestorePersisted(ctx context.Context) error {
	if s.checkpointStore == nil {
		return nil
	}
	checkpoint, err := s.checkpointStore.Load()
	if err != nil || checkpoint == nil {
		return err
	}
	if checkpoint.State.Status == contracts.SimulationStatusStopped || checkpoint.State.Status == contracts.SimulationStatusFailed {
		return s.checkpointStore.Clear()
	}
	_, err = s.Start(ctx, checkpoint.Config)
	return err
}

func (s *SimulationService) Start(
	ctx context.Context,
	config contracts.SimulationConfig,
) (contracts.SimulationSnapshot, error) {
	if ctx == nil {
		return contracts.SimulationSnapshot{}, fmt.Errorf("start context cannot be nil")
	}
	s.operationMu.Lock()
	defer s.operationMu.Unlock()

	s.mu.Lock()
	if s.starting {
		s.mu.Unlock()
		return contracts.SimulationSnapshot{}, apperrors.Conflict("simulation is already starting")
	}
	if s.engine != nil {
		status := s.engine.Snapshot().State.Status
		if status != contracts.SimulationStatusStopped && status != contracts.SimulationStatusFailed {
			s.mu.Unlock()
			return contracts.SimulationSnapshot{}, apperrors.Conflict("a simulation is already active")
		}
	}
	s.starting = true
	s.mu.Unlock()

	devices, err := s.devices.GetDevices(contracts.GetDevicesRequest{})
	if err != nil {
		s.finishStarting()
		return contracts.SimulationSnapshot{}, fmt.Errorf("load devices: %w", err)
	}
	gateways, err := s.gateways.GetGateways(contracts.GetGatewaysRequest{})
	if err != nil {
		s.finishStarting()
		return contracts.SimulationSnapshot{}, fmt.Errorf("load gateways: %w", err)
	}

	options := s.options
	options.EventSink = s.handleEvent
	if s.checkpointStore != nil {
		checkpoint, checkpointErr := s.checkpointStore.Load()
		if checkpointErr != nil {
			s.finishStarting()
			return contracts.SimulationSnapshot{}, checkpointErr
		}
		if checkpoint != nil && checkpoint.State.Status != contracts.SimulationStatusStopped && checkpoint.State.Status != contracts.SimulationStatusFailed {
			options.Checkpoint = checkpoint
		}
	}
	runtime, err := engine.New(config, devices.Devices, gateways.Gateways, options)
	if err != nil {
		s.finishStarting()
		return contracts.SimulationSnapshot{}, fmt.Errorf("create simulation engine: %w", err)
	}

	s.mu.Lock()
	s.engine = runtime
	s.starting = false
	s.mu.Unlock()

	if err := runtime.Start(ctx); err != nil {
		s.mu.Lock()
		if s.engine == runtime {
			s.engine = nil
		}
		s.mu.Unlock()
		return contracts.SimulationSnapshot{}, fmt.Errorf("start simulation engine: %w", err)
	}

	return runtime.Snapshot(), nil
}

func (s *SimulationService) Pause() (contracts.SimulationSnapshot, error) {
	s.operationMu.Lock()
	defer s.operationMu.Unlock()
	runtime, err := s.currentEngine()
	if err != nil {
		return contracts.SimulationSnapshot{}, err
	}
	if err := runtime.Pause(); err != nil {
		return contracts.SimulationSnapshot{}, err
	}
	return runtime.Snapshot(), nil
}

func (s *SimulationService) Resume() (contracts.SimulationSnapshot, error) {
	s.operationMu.Lock()
	defer s.operationMu.Unlock()
	runtime, err := s.currentEngine()
	if err != nil {
		return contracts.SimulationSnapshot{}, err
	}
	if err := runtime.Resume(); err != nil {
		return contracts.SimulationSnapshot{}, err
	}
	return runtime.Snapshot(), nil
}

func (s *SimulationService) Stop(ctx context.Context) (contracts.SimulationSnapshot, error) {
	s.operationMu.Lock()
	defer s.operationMu.Unlock()
	runtime, err := s.currentEngine()
	if err != nil {
		return contracts.SimulationSnapshot{}, err
	}
	if err := runtime.Stop(ctx); err != nil {
		return contracts.SimulationSnapshot{}, err
	}
	if s.checkpointStore != nil {
		if err := s.checkpointStore.Clear(); err != nil {
			return contracts.SimulationSnapshot{}, err
		}
	}
	return runtime.Snapshot(), nil
}

func (s *SimulationService) Snapshot() (contracts.SimulationSnapshot, error) {
	runtime, err := s.currentEngine()
	if err != nil {
		return contracts.SimulationSnapshot{}, err
	}
	return runtime.Snapshot(), nil
}

func (s *SimulationService) Subscribe() (SimulationSubscription, error) {
	if _, err := s.currentEngine(); err != nil {
		return SimulationSubscription{}, err
	}

	channel := make(chan SimulationUpdate, 128)
	s.mu.Lock()
	s.nextID++
	id := s.nextID
	s.subscribers[id] = channel
	s.mu.Unlock()

	return SimulationSubscription{
		Updates: channel,
		Close: func() {
			s.mu.Lock()
			if _, ok := s.subscribers[id]; ok {
				delete(s.subscribers, id)
			}
			s.mu.Unlock()
		},
	}, nil
}

func (s *SimulationService) currentEngine() (*engine.Engine, error) {
	s.mu.RLock()
	runtime := s.engine
	s.mu.RUnlock()
	if runtime == nil {
		return nil, apperrors.Conflict("%s", ErrSimulationNotStarted.Error())
	}
	return runtime, nil
}

func (s *SimulationService) finishStarting() {
	s.mu.Lock()
	s.starting = false
	s.mu.Unlock()
}

func (s *SimulationService) handleEvent(event contracts.SimulationEvent) {
	s.mu.RLock()
	runtime := s.engine
	updates := make([]chan SimulationUpdate, 0, len(s.subscribers))
	for _, channel := range s.subscribers {
		updates = append(updates, channel)
	}
	s.mu.RUnlock()
	if runtime == nil {
		return
	}

	update := SimulationUpdate{Snapshot: runtime.Snapshot(), Event: event}
	if s.checkpointStore != nil {
		if err := s.checkpointStore.Save(runtime.Checkpoint()); err != nil {
			log.Printf("simulation checkpoint save failed: %v", err)
		}
	}
	for _, channel := range updates {
		select {
		case channel <- update:
		default:
			// A slow websocket client must not pause the simulation loop.
		}
	}
}
