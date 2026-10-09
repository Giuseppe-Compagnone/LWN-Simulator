package services

import (
	"context"
	"errors"
	"testing"

	contracts "github.com/Giuseppe-Compagnone/lwn-contracts/generated"
	engineTypes "github.com/Giuseppe-Compagnone/lwn-engine/types"
	"lwn-simulator-backend/internal/apperrors"
	"lwn-simulator-backend/internal/database"
	"lwn-simulator-backend/internal/repositories"
)

func TestProfileRuntimeManagerIsolatesProfilesAndAllowsOneSimulation(t *testing.T) {
	dataDir := t.TempDir()
	profiles := NewProfileService(repositories.NewProfileRepository(dataDir), dataDir)
	defaultProfile, err := profiles.EnsureDefaultProfile()
	if err != nil {
		t.Fatal(err)
	}
	secondProfile, err := profiles.Create("Second network")
	if err != nil {
		t.Fatal(err)
	}
	manager := NewProfileRuntimeManager(dataDir, profiles, engineTypes.Options{}, nil)
	first, err := manager.Runtime(defaultProfile.ID)
	if err != nil {
		t.Fatal(err)
	}
	second, err := manager.Runtime(secondProfile.ID)
	if err != nil {
		t.Fatal(err)
	}
	if first.Device == second.Device || first.Gateway == second.Gateway || first.Configuration == second.Configuration || first.Simulation == second.Simulation {
		t.Fatal("profiles share runtime services")
	}
	if _, err := first.Configuration.UpdateGatewayBridge(contracts.GatewayBridgeConfig{
		Enabled: true,
		Address: "127.0.0.1",
		Port:    1701,
	}); err != nil {
		t.Fatal(err)
	}
	secondBridge, err := second.Configuration.GetGatewayBridge()
	if err != nil {
		t.Fatal(err)
	}
	if secondBridge != defaultGatewayBridgeConfig {
		t.Fatalf("profiles share gateway bridge configuration: %+v", secondBridge)
	}
	firstStorage, err := database.NewProfileStorage(dataDir, defaultProfile.ID)
	if err != nil {
		t.Fatal(err)
	}
	secondStorage, err := database.NewProfileStorage(dataDir, secondProfile.ID)
	if err != nil {
		t.Fatal(err)
	}
	if firstStorage.Directory() == secondStorage.Directory() {
		t.Fatal("profiles share storage")
	}

	ctx := context.Background()
	if _, err := first.Simulation.Start(ctx, contracts.SimulationConfig{Speed: 1}); err != nil {
		t.Fatal(err)
	}
	activity, err := manager.ActiveSimulation()
	if err != nil || !activity.Active || activity.ProfileID == nil || *activity.ProfileID != defaultProfile.ID {
		t.Fatalf("active simulation activity = %+v, err = %v", activity, err)
	}
	if _, err := second.Simulation.Start(ctx, contracts.SimulationConfig{Speed: 1}); err == nil {
		t.Fatal("second profile started while another simulation was active")
	} else if !errors.Is(err, apperrors.ErrConflict) {
		t.Fatalf("starting a simulation on another profile returned %v, want a conflict", err)
	}
	if _, err := first.Simulation.Stop(ctx); err != nil {
		t.Fatal(err)
	}
	activity, err = manager.ActiveSimulation()
	if err != nil || activity.Active {
		t.Fatalf("simulation remained active after stop: %+v, err = %v", activity, err)
	}
	if _, err := second.Simulation.Start(ctx, contracts.SimulationConfig{Speed: 1}); err != nil {
		t.Fatalf("second profile could not start after the first stopped: %v", err)
	}
	if _, err := second.Simulation.Stop(ctx); err != nil {
		t.Fatal(err)
	}
}

func TestProfileRuntimeManagerExportsAndImportsAProfileAsANewProfile(t *testing.T) {
	dataDir := t.TempDir()
	profiles := NewProfileService(repositories.NewProfileRepository(dataDir), dataDir)
	defaultProfile, err := profiles.EnsureDefaultProfile()
	if err != nil {
		t.Fatal(err)
	}
	manager := NewProfileRuntimeManager(dataDir, profiles, engineTypes.Options{}, nil)
	if _, err := manager.Runtime(defaultProfile.ID); err != nil {
		t.Fatal(err)
	}
	runtime, err := manager.Runtime(defaultProfile.ID)
	if err != nil {
		t.Fatal(err)
	}
	run := contracts.SimulationRun{
		Summary: contracts.SimulationRunSummary{
			Id:                    "550e8400-e29b-41d4-a716-446655440100",
			StartedAtMilliseconds: 100,
			EndedAtMilliseconds:   int64Pointer(200),
			Status:                contracts.SimulationRunStatusStopped,
			EventCount:            1,
			PacketSuccessRate:     100,
		},
		Objects: []contracts.SimulationLogObject{},
	}
	if err := runtime.Simulation.inner.logStore.Import(run, []contracts.SimulationEvent{{
		ID:       "550e8400-e29b-41d4-a716-446655440101",
		Sequence: 1,
		Type:     contracts.SimulationStarted,
	}}); err != nil {
		t.Fatal(err)
	}
	if _, err := runtime.Configuration.UpdateGatewayBridge(contracts.GatewayBridgeConfig{
		Enabled: true,
		Address: "127.0.0.1",
		Port:    1701,
	}); err != nil {
		t.Fatal(err)
	}

	archive, err := manager.ExportArchive(defaultProfile.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(archive.Logs) != 1 || len(archive.Logs[0].Events) != 1 {
		t.Fatalf("simulation history was not exported: %+v", archive.Logs)
	}
	if archive.GatewayBridge == nil || archive.GatewayBridge.Address != "127.0.0.1" {
		t.Fatalf("gateway bridge configuration was not exported: %+v", archive.GatewayBridge)
	}
	imported, err := manager.ImportArchive(archive)
	if err != nil {
		t.Fatal(err)
	}
	if imported.ID == defaultProfile.ID || imported.Name == defaultProfile.Name {
		t.Fatalf("imported profile was not recreated: %+v", imported)
	}
	importedArchive, err := manager.ExportArchive(imported.ID)
	if err != nil {
		t.Fatal(err)
	}
	if importedArchive.Profile.ID != imported.ID {
		t.Fatalf("archive kept the source profile id: %+v", importedArchive.Profile)
	}
	if importedArchive.GatewayBridge == nil || importedArchive.GatewayBridge.Port != 1701 {
		t.Fatalf("gateway bridge configuration was not imported: %+v", importedArchive.GatewayBridge)
	}
}

func TestProfileRuntimeManagerRejectsInvalidArchiveBeforeCreatingProfile(t *testing.T) {
	dataDir := t.TempDir()
	profiles := NewProfileService(repositories.NewProfileRepository(dataDir), dataDir)
	defaultProfile, err := profiles.EnsureDefaultProfile()
	if err != nil {
		t.Fatal(err)
	}
	manager := NewProfileRuntimeManager(dataDir, profiles, engineTypes.Options{}, nil)

	_, err = manager.ImportArchive(contracts.ProfileArchive{
		Format:  profileArchiveFormat,
		Version: profileArchiveVersion,
		Profile: contracts.Profile{ID: "550e8400-e29b-41d4-a716-446655440000", Name: "Invalid import"},
		Devices: []contracts.Device{{ID: "not-a-uuid"}},
	})
	if err == nil {
		t.Fatal("invalid archive was accepted")
	}

	stored, err := profiles.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(stored) != 1 || stored[0].ID != defaultProfile.ID {
		t.Fatalf("invalid import created a profile: %+v", stored)
	}
}
