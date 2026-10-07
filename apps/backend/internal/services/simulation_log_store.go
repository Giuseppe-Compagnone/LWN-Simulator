package services

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	contracts "github.com/Giuseppe-Compagnone/lwn-contracts/generated"
)

type SimulationLogStore interface {
	List() (contracts.SimulationRunsResponse, error)
	GetRun(string) (contracts.SimulationRun, error)
	Get(string) (contracts.SimulationRun, []contracts.SimulationEvent, error)
	Events(string, SimulationLogEventQuery) (contracts.SimulationEventsResponse, error)
	Create(contracts.SimulationRun) error
	Append(string, contracts.SimulationEvent) error
	AppendBatch(string, []contracts.SimulationEvent, *contracts.SimulationRunSummary) error
	UpdateSummary(string, contracts.SimulationRunSummary) error
	Finalize(string, contracts.SimulationRunSummary) error
}

type SimulationLogEventQuery struct {
	AfterSequence int64
	Limit         int
	DeviceID      string
	GatewayID     string
	Types         map[contracts.SimulationEventType]struct{}
}

type persistedSimulationRun struct {
	Run    contracts.SimulationRun     `json:"run"`
	Events []contracts.SimulationEvent `json:"events"`
}

type simulationEventOffset struct {
	Sequence int64 `json:"sequence"`
	Offset   int64 `json:"offset"`
}

type FileSimulationLogStore struct {
	path   string
	mu     sync.Mutex
	loaded bool
	runs   []persistedSimulationRun
}

func NewFileSimulationLogStore(path string) *FileSimulationLogStore {
	return &FileSimulationLogStore{path: path}
}

func (store *FileSimulationLogStore) List() (contracts.SimulationRunsResponse, error) {
	store.mu.Lock()
	defer store.mu.Unlock()
	runs, err := store.loadLocked()
	if err != nil {
		return contracts.SimulationRunsResponse{}, err
	}
	result := make([]contracts.SimulationRun, len(runs))
	for index, run := range runs {
		result[index] = run.Run
	}
	sort.Slice(result, func(left, right int) bool {
		return result[left].Summary.StartedAtMilliseconds > result[right].Summary.StartedAtMilliseconds
	})
	return contracts.SimulationRunsResponse{Runs: result, Total: int32(len(result))}, nil
}

func (store *FileSimulationLogStore) Get(id string) (contracts.SimulationRun, []contracts.SimulationEvent, error) {
	store.mu.Lock()
	defer store.mu.Unlock()
	runs, err := store.loadLocked()
	if err != nil {
		return contracts.SimulationRun{}, nil, err
	}
	for _, run := range runs {
		if run.Run.Summary.Id == id {
			events := append([]contracts.SimulationEvent(nil), run.Events...)
			sidecar, sidecarErr := store.loadEventSidecarLocked(id)
			if sidecarErr != nil {
				return contracts.SimulationRun{}, nil, sidecarErr
			}
			events = append(events, sidecar...)
			return run.Run, events, nil
		}
	}
	return contracts.SimulationRun{}, nil, os.ErrNotExist
}

func (store *FileSimulationLogStore) GetRun(id string) (contracts.SimulationRun, error) {
	store.mu.Lock()
	defer store.mu.Unlock()
	runs, err := store.loadLocked()
	if err != nil {
		return contracts.SimulationRun{}, err
	}
	for _, run := range runs {
		if run.Run.Summary.Id == id {
			return run.Run, nil
		}
	}
	return contracts.SimulationRun{}, os.ErrNotExist
}

func (store *FileSimulationLogStore) Events(id string, query SimulationLogEventQuery) (contracts.SimulationEventsResponse, error) {
	store.mu.Lock()
	defer store.mu.Unlock()
	runs, err := store.loadLocked()
	if err != nil {
		return contracts.SimulationEventsResponse{}, err
	}
	for _, run := range runs {
		if run.Run.Summary.Id != id {
			continue
		}
		result := make([]contracts.SimulationEvent, 0, max(0, query.Limit))
		appendEvent := func(event contracts.SimulationEvent) bool {
			if !simulationLogEventMatches(event, query) {
				return false
			}
			result = append(result, event)
			return query.Limit > 0 && len(result) >= query.Limit
		}
		for _, event := range run.Events {
			if appendEvent(event) {
				return contracts.SimulationEventsResponse{Events: result, LastSequence: int64(run.Run.Summary.EventCount)}, nil
			}
		}
		if err := store.scanEventSidecarLocked(id, query.AfterSequence, appendEvent); err != nil {
			return contracts.SimulationEventsResponse{}, err
		}
		return contracts.SimulationEventsResponse{Events: result, LastSequence: int64(run.Run.Summary.EventCount)}, nil
	}
	return contracts.SimulationEventsResponse{}, os.ErrNotExist
}

func simulationLogEventMatches(event contracts.SimulationEvent, query SimulationLogEventQuery) bool {
	if event.Sequence <= query.AfterSequence {
		return false
	}
	if query.DeviceID != "" && (event.DeviceID == nil || *event.DeviceID != query.DeviceID) {
		return false
	}
	if query.GatewayID != "" && (event.GatewayID == nil || *event.GatewayID != query.GatewayID) {
		return false
	}
	if len(query.Types) > 0 {
		if _, ok := query.Types[event.Type]; !ok {
			return false
		}
	}
	return true
}

func (store *FileSimulationLogStore) Create(run contracts.SimulationRun) error {
	store.mu.Lock()
	defer store.mu.Unlock()
	runs, err := store.loadLocked()
	if err != nil {
		return err
	}
	runs = append(runs, persistedSimulationRun{Run: run, Events: []contracts.SimulationEvent{}})
	return store.saveLocked(runs)
}

func (store *FileSimulationLogStore) Append(id string, event contracts.SimulationEvent) error {
	return store.AppendBatch(id, []contracts.SimulationEvent{event}, nil)
}

// AppendBatch persists a group of events and its latest summary with one
// atomic file replacement. Keeping the decoded runs in memory avoids reading
// and decoding an ever-growing history for every engine event.
func (store *FileSimulationLogStore) AppendBatch(id string, events []contracts.SimulationEvent, summary *contracts.SimulationRunSummary) error {
	store.mu.Lock()
	defer store.mu.Unlock()
	runs, err := store.loadLocked()
	if err != nil {
		return err
	}
	for index := range runs {
		if runs[index].Run.Summary.Id != id {
			continue
		}
		if err := store.appendEventSidecarLocked(id, events); err != nil {
			return err
		}
		eventCount := runs[index].Run.Summary.EventCount + int32(len(events))
		if summary != nil {
			runs[index].Run.Summary = *summary
		}
		runs[index].Run.Summary.EventCount = eventCount
		return store.saveLocked(runs)
	}
	return os.ErrNotExist
}

func (store *FileSimulationLogStore) Finalize(id string, summary contracts.SimulationRunSummary) error {
	return store.UpdateSummary(id, summary)
}

func (store *FileSimulationLogStore) UpdateSummary(id string, summary contracts.SimulationRunSummary) error {
	store.mu.Lock()
	defer store.mu.Unlock()
	runs, err := store.loadLocked()
	if err != nil {
		return err
	}
	for index := range runs {
		if runs[index].Run.Summary.Id == id {
			eventCount := runs[index].Run.Summary.EventCount
			runs[index].Run.Summary = summary
			runs[index].Run.Summary.EventCount = eventCount
			return store.saveLocked(runs)
		}
	}
	return os.ErrNotExist
}

func (store *FileSimulationLogStore) eventDirectory() string {
	extension := filepath.Ext(store.path)
	return strings.TrimSuffix(store.path, extension) + ".events"
}

func (store *FileSimulationLogStore) eventPath(id string) (string, error) {
	if id == "" || filepath.Base(id) != id {
		return "", fmt.Errorf("invalid simulation run id %q", id)
	}
	return filepath.Join(store.eventDirectory(), id+".jsonl"), nil
}

func (store *FileSimulationLogStore) eventIndexPath(id string) (string, error) {
	if id == "" || filepath.Base(id) != id {
		return "", fmt.Errorf("invalid simulation run id %q", id)
	}
	return filepath.Join(store.eventDirectory(), id+".idx"), nil
}

func (store *FileSimulationLogStore) appendEventSidecarLocked(id string, events []contracts.SimulationEvent) error {
	if len(events) == 0 {
		return nil
	}
	path, err := store.eventPath(id)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return fmt.Errorf("create simulation event directory: %w", err)
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return fmt.Errorf("open simulation event log: %w", err)
	}
	offset, err := file.Seek(0, io.SeekEnd)
	if err != nil {
		_ = file.Close()
		return fmt.Errorf("seek simulation event log: %w", err)
	}
	encoder := json.NewEncoder(file)
	for _, event := range events {
		if err := encoder.Encode(event); err != nil {
			_ = file.Close()
			return fmt.Errorf("encode simulation event: %w", err)
		}
	}
	if err := file.Sync(); err != nil {
		_ = file.Close()
		return fmt.Errorf("sync simulation event log: %w", err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("close simulation event log: %w", err)
	}
	if err := store.appendEventOffsetLocked(id, simulationEventOffset{Sequence: events[0].Sequence, Offset: offset}); err != nil {
		// The event stream is the source of truth. A missing index only makes a
		// historical query slower and must never cause the durable events to be
		// retried and duplicated.
		log.Printf("append simulation event index: %v", err)
	}
	return nil
}

func (store *FileSimulationLogStore) appendEventOffsetLocked(id string, entry simulationEventOffset) error {
	path, err := store.eventIndexPath(id)
	if err != nil {
		return err
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return fmt.Errorf("open simulation event index: %w", err)
	}
	if err := json.NewEncoder(file).Encode(entry); err != nil {
		_ = file.Close()
		return fmt.Errorf("encode simulation event index: %w", err)
	}
	if err := file.Sync(); err != nil {
		_ = file.Close()
		return fmt.Errorf("sync simulation event index: %w", err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("close simulation event index: %w", err)
	}
	return nil
}

func (store *FileSimulationLogStore) loadEventSidecarLocked(id string) ([]contracts.SimulationEvent, error) {
	events := make([]contracts.SimulationEvent, 0)
	err := store.scanEventSidecarLocked(id, 0, func(event contracts.SimulationEvent) bool {
		events = append(events, event)
		return false
	})
	return events, err
}

func (store *FileSimulationLogStore) scanEventSidecarLocked(id string, afterSequence int64, visit func(contracts.SimulationEvent) bool) error {
	path, err := store.eventPath(id)
	if err != nil {
		return err
	}
	file, err := os.Open(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("open simulation event log: %w", err)
	}
	defer file.Close()
	if afterSequence > 0 {
		offset, offsetErr := store.eventOffsetLocked(id, afterSequence)
		if offsetErr != nil {
			return offsetErr
		}
		if offset > 0 {
			if _, err := file.Seek(offset, io.SeekStart); err != nil {
				return fmt.Errorf("seek simulation event log: %w", err)
			}
		}
	}

	scanner := bufio.NewScanner(file)
	buffer := make([]byte, 64*1024)
	scanner.Buffer(buffer, 4*1024*1024)
	for scanner.Scan() {
		var event contracts.SimulationEvent
		if err := json.Unmarshal(scanner.Bytes(), &event); err != nil {
			return fmt.Errorf("decode simulation event: %w", err)
		}
		if visit(event) {
			return nil
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("read simulation event log: %w", err)
	}
	return nil
}

func (store *FileSimulationLogStore) eventOffsetLocked(id string, afterSequence int64) (int64, error) {
	path, err := store.eventIndexPath(id)
	if err != nil {
		return 0, err
	}
	file, err := os.Open(path)
	if errors.Is(err, os.ErrNotExist) {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("open simulation event index: %w", err)
	}
	defer file.Close()

	offset := int64(0)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var entry simulationEventOffset
		if err := json.Unmarshal(scanner.Bytes(), &entry); err != nil {
			return 0, fmt.Errorf("decode simulation event index: %w", err)
		}
		if entry.Sequence > afterSequence {
			break
		}
		offset = entry.Offset
	}
	if err := scanner.Err(); err != nil {
		return 0, fmt.Errorf("read simulation event index: %w", err)
	}
	return offset, nil
}

func (store *FileSimulationLogStore) loadLocked() ([]persistedSimulationRun, error) {
	if store.loaded {
		return store.runs, nil
	}
	data, err := os.ReadFile(store.path)
	if errors.Is(err, os.ErrNotExist) {
		store.runs = []persistedSimulationRun{}
		store.loaded = true
		return store.runs, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read simulation logs: %w", err)
	}
	var runs []persistedSimulationRun
	if err := json.Unmarshal(data, &runs); err != nil {
		return nil, fmt.Errorf("decode simulation logs: %w", err)
	}
	store.runs = runs
	store.loaded = true
	return store.runs, nil
}

func (store *FileSimulationLogStore) saveLocked(runs []persistedSimulationRun) error {
	data, err := json.Marshal(runs)
	if err != nil {
		return fmt.Errorf("encode simulation logs: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(store.path), 0700); err != nil {
		return fmt.Errorf("create simulation log directory: %w", err)
	}
	temporary, err := os.CreateTemp(filepath.Dir(store.path), ".simulation-logs-*.tmp")
	if err != nil {
		return fmt.Errorf("create temporary simulation logs: %w", err)
	}
	name := temporary.Name()
	defer os.Remove(name)
	if err := temporary.Chmod(0600); err != nil {
		_ = temporary.Close()
		return err
	}
	if _, err := temporary.Write(data); err != nil {
		_ = temporary.Close()
		return err
	}
	if err := temporary.Sync(); err != nil {
		_ = temporary.Close()
		return err
	}
	if err := temporary.Close(); err != nil {
		return err
	}
	if err := os.Rename(name, store.path); err != nil {
		return err
	}
	store.runs = runs
	store.loaded = true
	return nil
}
