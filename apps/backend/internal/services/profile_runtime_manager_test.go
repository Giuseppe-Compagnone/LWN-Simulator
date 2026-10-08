package services

import (
	"context"
	"testing"

	contracts "github.com/Giuseppe-Compagnone/lwn-contracts/generated"
	engineTypes "github.com/Giuseppe-Compagnone/lwn-engine/types"
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
	if first.Device == second.Device || first.Gateway == second.Gateway || first.Simulation == second.Simulation {
		t.Fatal("profiles share runtime services")
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
	if _, err := second.Simulation.Start(ctx, contracts.SimulationConfig{Speed: 1}); err == nil {
		t.Fatal("second profile started while another simulation was active")
	}
	if _, err := first.Simulation.Stop(ctx); err != nil {
		t.Fatal(err)
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

	archive, err := manager.ExportArchive(defaultProfile.ID)
	if err != nil {
		t.Fatal(err)
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
}
