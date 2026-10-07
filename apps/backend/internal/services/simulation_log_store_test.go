package services

import (
	"os"
	"path/filepath"
	"testing"

	contracts "github.com/Giuseppe-Compagnone/lwn-contracts/generated"
)

func TestFileSimulationLogStorePersistsRunsAndFiltersByEventAtReadTime(t *testing.T) {
	store := NewFileSimulationLogStore(filepath.Join(t.TempDir(), "logs.json"))
	run := contracts.SimulationRun{Summary: contracts.SimulationRunSummary{Id: "550e8400-e29b-41d4-a716-446655440000", StartedAtMilliseconds: 10, Status: contracts.SimulationRunStatusRunning, Speed: 1}, Objects: []contracts.SimulationLogObject{{Id: "550e8400-e29b-41d4-a716-446655440001", Name: "Weather station", Kind: contracts.SimulationLogObjectKindDevice, Identifier: "70B3D57ED0001001"}}}
	if err := store.Create(run); err != nil {
		t.Fatal(err)
	}
	deviceID := "550e8400-e29b-41d4-a716-446655440001"
	if err := store.Append(run.Summary.Id, contracts.SimulationEvent{ID: "550e8400-e29b-41d4-a716-446655440002", Sequence: 1, Type: contracts.DeviceUplinkTransmitted, DeviceID: &deviceID, Message: "uplink"}); err != nil {
		t.Fatal(err)
	}
	loaded, events, err := store.Get(run.Summary.Id)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Summary.EventCount != 1 || len(events) != 1 || events[0].DeviceID == nil || *events[0].DeviceID != deviceID {
		t.Fatalf("unexpected persisted run: %+v %+v", loaded, events)
	}
	updatedSummary := loaded.Summary
	updatedSummary.PacketSuccessRate = 0.75
	updatedSummary.DurationMilliseconds = 1500
	if err := store.UpdateSummary(run.Summary.Id, updatedSummary); err != nil {
		t.Fatal(err)
	}
	updated, _, err := store.Get(run.Summary.Id)
	if err != nil {
		t.Fatal(err)
	}
	if updated.Summary.PacketSuccessRate != 0.75 || updated.Summary.EventCount != 1 {
		t.Fatalf("live summary was not persisted: %+v", updated.Summary)
	}
	list, err := store.List()
	if err != nil {
		t.Fatal(err)
	}
	if list.Total != 1 || len(list.Runs) != 1 {
		t.Fatalf("unexpected list response: %+v", list)
	}
}

func TestFileSimulationLogStorePersistsEventBatchesInSidecar(t *testing.T) {
	path := filepath.Join(t.TempDir(), "logs.json")
	store := NewFileSimulationLogStore(path)
	runID := "550e8400-e29b-41d4-a716-446655440010"
	run := contracts.SimulationRun{Summary: contracts.SimulationRunSummary{Id: runID, StartedAtMilliseconds: 10, Status: contracts.SimulationRunStatusRunning, Speed: 1}}
	if err := store.Create(run); err != nil {
		t.Fatal(err)
	}
	events := []contracts.SimulationEvent{
		{ID: "550e8400-e29b-41d4-a716-446655440011", Sequence: 1, Type: contracts.DeviceRegistered, Message: "registered"},
		{ID: "550e8400-e29b-41d4-a716-446655440012", Sequence: 2, Type: contracts.DeviceUplinkTransmitted, Message: "uplink"},
	}
	summary := run.Summary
	summary.PacketSuccessRate = 1
	if err := store.AppendBatch(runID, events, &summary); err != nil {
		t.Fatal(err)
	}

	// Re-open the store to prove that events are durable and not served only
	// by the in-memory cache.
	reopened := NewFileSimulationLogStore(path)
	loaded, persisted, err := reopened.Get(runID)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Summary.EventCount != 2 || loaded.Summary.PacketSuccessRate != 1 || len(persisted) != 2 {
		t.Fatalf("unexpected persisted batch: run=%+v events=%+v", loaded.Summary, persisted)
	}

	page, err := reopened.Events(runID, SimulationLogEventQuery{AfterSequence: 1, Limit: 1, Types: map[contracts.SimulationEventType]struct{}{contracts.DeviceUplinkTransmitted: {}}})
	if err != nil {
		t.Fatal(err)
	}
	if len(page.Events) != 1 || page.Events[0].Sequence != 2 || page.LastSequence != 2 {
		t.Fatalf("unexpected event page: %+v", page)
	}

	indexPath, err := reopened.eventIndexPath(runID)
	if err != nil {
		t.Fatal(err)
	}
	if info, err := os.Stat(indexPath); err != nil || info.Size() == 0 {
		t.Fatalf("event index was not persisted: info=%v err=%v", info, err)
	}
}
