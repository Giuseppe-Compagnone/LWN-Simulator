package services

import (
	"context"
	"testing"
	"time"

	contracts "github.com/Giuseppe-Compagnone/lwn-contracts/generated"
	"github.com/Giuseppe-Compagnone/lwn-engine/types"
)

type simulationDeviceSource struct{}

func (simulationDeviceSource) GetDevices(contracts.GetDevicesRequest) (contracts.GetDevicesResponse, error) {
	return contracts.GetDevicesResponse{Devices: []contracts.Device{}}, nil
}

type simulationGatewaySource struct{}

func (simulationGatewaySource) GetGateways(contracts.GetGatewaysRequest) (contracts.GetGatewaysResponse, error) {
	return contracts.GetGatewaysResponse{Gateways: []contracts.Gateway{}}, nil
}

func TestSimulationServiceLifecycleAndSingleSession(t *testing.T) {
	service := NewSimulationService(
		simulationDeviceSource{},
		simulationGatewaySource{},
		types.Options{EventBuffer: 32},
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	snapshot, err := service.Start(ctx, contracts.SimulationConfig{Speed: 1})
	if err != nil {
		t.Fatalf("start simulation: %v", err)
	}
	if snapshot.State.Status != contracts.SimulationStatusRunning {
		t.Fatalf("unexpected start status: %s", snapshot.State.Status)
	}

	if _, err := service.Start(ctx, contracts.SimulationConfig{Speed: 1}); err == nil {
		t.Fatal("starting a second simulation should fail")
	}

	subscription, err := service.Subscribe()
	if err != nil {
		t.Fatalf("subscribe to simulation: %v", err)
	}
	defer subscription.Close()

	if _, err := service.Pause(); err != nil {
		t.Fatalf("pause simulation: %v", err)
	}
	if update := receiveSimulationUpdate(t, subscription.Updates); update.Event.Type != contracts.SimulationPaused {
		t.Fatalf("unexpected pause event: %s", update.Event.Type)
	}

	if _, err := service.Resume(); err != nil {
		t.Fatalf("resume simulation: %v", err)
	}
	if update := receiveSimulationUpdate(t, subscription.Updates); update.Event.Type != contracts.SimulationResumed {
		t.Fatalf("unexpected resume event: %s", update.Event.Type)
	}

	if _, err := service.Stop(context.Background()); err != nil {
		t.Fatalf("stop simulation: %v", err)
	}
	if snapshot, err := service.Snapshot(); err != nil || snapshot.State.Status != contracts.SimulationStatusStopped {
		t.Fatalf("unexpected stopped snapshot: %+v, %v", snapshot.State.Status, err)
	}
}

func receiveSimulationUpdate(t *testing.T, updates <-chan SimulationUpdate) SimulationUpdate {
	t.Helper()
	select {
	case update := <-updates:
		return update
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for simulation update")
		return SimulationUpdate{}
	}
}
