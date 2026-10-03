package services

import (
	"path/filepath"
	"testing"

	contracts "github.com/Giuseppe-Compagnone/lwn-contracts/generated"
	"github.com/Giuseppe-Compagnone/lwn-engine/types"
)

func TestFileSimulationCheckpointStoreRoundTripAndAtomicClear(t *testing.T) {
	store := NewFileSimulationCheckpointStore(filepath.Join(t.TempDir(), "simulation-checkpoint.json"))
	seed := int64(42)
	want := types.EngineCheckpoint{
		Config:   contracts.SimulationConfig{Speed: 2, Seed: &seed},
		State:    contracts.SimulationState{Status: contracts.SimulationStatusRunning, Speed: 2, ElapsedMilliseconds: 100},
		Sessions: map[string]types.DeviceSession{"device": {Joined: true, FrameCounterUp: 3}},
	}
	if err := store.Save(want); err != nil {
		t.Fatalf("save checkpoint: %v", err)
	}
	got, err := store.Load()
	if err != nil {
		t.Fatalf("load checkpoint: %v", err)
	}
	if got == nil || got.Config.Speed != want.Config.Speed || got.State.ElapsedMilliseconds != want.State.ElapsedMilliseconds || got.Sessions["device"].FrameCounterUp != 3 {
		t.Fatalf("unexpected checkpoint after round trip: %+v", got)
	}
	if err := store.Clear(); err != nil {
		t.Fatalf("clear checkpoint: %v", err)
	}
	got, err = store.Load()
	if err != nil || got != nil {
		t.Fatalf("expected empty checkpoint after clear, got %v, %v", got, err)
	}
}
