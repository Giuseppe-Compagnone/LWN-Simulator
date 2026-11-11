package services

import (
	"errors"
	"testing"

	contracts "github.com/Giuseppe-Compagnone/lwn-contracts/generated"
	"github.com/Giuseppe-Compagnone/lwn-engine/types"
	"lwn-simulator-backend/internal/apperrors"
)

func TestLiveDeviceUpdatesAllowOnlyStatusPayloadAndEncoding(t *testing.T) {
	for _, paused := range []bool{false, true} {
		t.Run(map[bool]string{false: "running", true: "paused"}[paused], func(t *testing.T) {
			device := validSimulationDevice(testDeviceID1)
			repo := &mockDeviceRepository{devices: []contracts.Device{device}}
			devices := NewDeviceService(repo)
			simulation := NewSimulationService(devices, simulationGatewaySource{}, types.Options{Clock: types.NewManualClock(0)})
			devices.SetRuntimeSynchronizer(simulation)
			if _, err := simulation.Start(t.Context(), contracts.SimulationConfig{Speed: 1}); err != nil {
				t.Fatal(err)
			}
			defer simulation.Stop(t.Context())
			if paused {
				if _, err := simulation.Pause(); err != nil {
					t.Fatal(err)
				}
			}
			updated := device
			updated.PayloadConfig.Payload = "AQID"
			updated.PayloadConfig.Base64Encoded = true
			if _, err := devices.UpdateDevice(contracts.UpdateDeviceRequest{ID: device.ID, Device: updated}); err != nil {
				t.Fatal(err)
			}
			if repo.lastUpdatedDevice.PayloadConfig.Payload != "AQID" {
				t.Fatal("payload was not persisted")
			}
			for _, mutation := range []func(*contracts.Device){
				func(d *contracts.Device) { d.Name = "Changed" },
				func(d *contracts.Device) { d.Class = contracts.ClassA },
				func(d *contracts.Device) { d.PayloadConfig.UplinkInterval = 5 },
				func(d *contracts.Device) { d.PayloadConfig.MType = contracts.ConfirmedDataUp },
				func(d *contracts.Device) { d.FrameConfig.FPort = 2 },
				func(d *contracts.Device) { d.RX1Config.Delay = serviceIntPtr(2000) },
				func(d *contracts.Device) { d.LocationConfig.Latitude = serviceFloat32Ptr(40) },
				func(d *contracts.Device) {
					d.PayloadConfig.PayloadCodec = &contracts.PayloadCodec{Enabled: true, Script: "function decodeUplink(input) { return {data:{}}; }"}
				},
			} {
				forbidden := updated
				mutation(&forbidden)
				writes := repo.updateCalls
				_, err := devices.UpdateDevice(contracts.UpdateDeviceRequest{ID: device.ID, Device: forbidden})
				if !errors.Is(err, apperrors.ErrInvalid) || repo.updateCalls != writes {
					t.Fatalf("forbidden live edit was persisted or returned wrong error: %v", err)
				}
			}
			updated.Active = false
			if _, err := devices.UpdateDevice(contracts.UpdateDeviceRequest{ID: device.ID, Device: updated}); err != nil {
				t.Fatal(err)
			}
			if _, err := simulation.Stop(t.Context()); err != nil {
				t.Fatal(err)
			}
			updated.Name = "After stop"
			if _, err := devices.UpdateDevice(contracts.UpdateDeviceRequest{ID: device.ID, Device: updated}); err != nil {
				t.Fatal(err)
			}
		})
	}
}
