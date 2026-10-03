package validation

import (
	"math"
	"strings"
	"testing"

	contracts "github.com/Giuseppe-Compagnone/lwn-contracts/generated"
)

func TestValidateSimulationConfigAcceptsFinitePositiveSpeed(t *testing.T) {
	for _, speed := range []float64{0.01, 1, 100} {
		if err := ValidateSimulationConfig(contracts.SimulationConfig{Speed: speed}); err != nil {
			t.Fatalf("speed %v should be valid: %v", speed, err)
		}
	}
}

func TestValidateSimulationConfigRejectsNonFiniteOrNonPositiveSpeed(t *testing.T) {
	for _, speed := range []float64{0, -1, math.NaN(), math.Inf(1), math.Inf(-1)} {
		if err := ValidateSimulationConfig(contracts.SimulationConfig{Speed: speed}); err == nil {
			t.Fatalf("speed %v should be rejected", speed)
		}
	}
}

func TestValidateHardwareAcceptsValidFleet(t *testing.T) {
	if err := ValidateHardware([]contracts.Device{validationDevice()}, []contracts.Gateway{validationGateway()}); err != nil {
		t.Fatalf("valid fleet was rejected: %v", err)
	}
}

func TestValidateHardwareCoversDeviceAndGatewayRules(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*contracts.Device, *contracts.Gateway)
		field  string
	}{
		{"device identity", func(device *contracts.Device, _ *contracts.Gateway) {
			device.ID = "bad"
			device.Name = " "
			device.DevEUI = "zz"
		}, "devices[0].id"},
		{"device enums and location", func(device *contracts.Device, _ *contracts.Gateway) {
			device.Class = contracts.DeviceClass("invalid")
			device.Activation = contracts.DeviceActivation("invalid")
			device.LocationConfig.Latitude = float32Pointer(91)
			device.LocationConfig.Longitude = float32Pointer(-181)
			device.LocationConfig.Altitude = nil
			device.LocationConfig.Region = contracts.DeviceRegion("invalid")
		}, "devices[0].class"},
		{"device radio configuration", func(device *contracts.Device, _ *contracts.Gateway) {
			device.RX1Config.Delay = nil
			device.RX1Config.Duration = intPointer(0)
			device.RX1Config.DataRateOffset = intPointer(6)
			device.RX2Config.ChannelFrequency = 0
			device.RX2Config.DataRate = intPointer(6)
			device.RX2Config.ACKTimeout = 999
		}, "devices[0].RX1Config.delay"},
		{"device frame and payload", func(device *contracts.Device, _ *contracts.Gateway) {
			device.FrameConfig.FPort = 0
			device.FrameConfig.Retransmission = intPointer(-1)
			device.PayloadConfig.UplinkInterval = 0
			device.PayloadConfig.OversizedPayloadBehavior = contracts.OversizedPayloadBehavior("invalid")
			device.PayloadConfig.MType = contracts.DeviceMType("invalid")
			device.PayloadConfig.Base64Encoded = true
			device.PayloadConfig.Payload = "!"
		}, "devices[0].frameConfig.fPort"},
		{"device advanced settings", func(device *contracts.Device, _ *contracts.Gateway) {
			device.AdvancedConfig.AntennaRange = float32(math.Inf(1))
			rate := 6
			device.AdvancedConfig.UplinkDataRate = &rate
		}, "devices[0].advancedConfig.antennaRange"},
		{"missing OTAA config", func(device *contracts.Device, _ *contracts.Gateway) { device.OOTAConfig = nil }, "devices[0].OOTAConfig"},
		{"incompatible ABP config", func(device *contracts.Device, _ *contracts.Gateway) { device.ABPConfig = &contracts.ABPConfig{} }, "devices[0].ABPConfig"},
		{"gateway identity and location", func(_ *contracts.Device, gateway *contracts.Gateway) {
			gateway.ID = "bad"
			gateway.Name = ""
			gateway.MacAddress = "bad"
			gateway.GatewayEUI = "bad"
			gateway.Latitude = float32Pointer(91)
			gateway.Longitude = float32Pointer(-181)
			gateway.Altitude = nil
		}, "gateways[0].id"},
		{"virtual gateway fields", func(_ *contracts.Device, gateway *contracts.Gateway) {
			endpoint := "127.0.0.1"
			gateway.GatewayIPv4 = &endpoint
			gateway.GatewayPort = int32Pointer(1700)
			gateway.KeepAlive = int32Pointer(0)
		}, "gateways[0].keepAlive"},
		{"real gateway fields", func(_ *contracts.Device, gateway *contracts.Gateway) {
			gateway.Type = contracts.Real
			gateway.KeepAlive = int32Pointer(30)
			gateway.GatewayIPv4 = stringPointer("not-ip")
			gateway.GatewayPort = int32Pointer(65536)
		}, "gateways[0].keepAlive"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			device, gateway := validationDevice(), validationGateway()
			test.mutate(&device, &gateway)
			err := ValidateHardware([]contracts.Device{device}, []contracts.Gateway{gateway})
			if err == nil || !strings.Contains(err.Error(), test.field) {
				t.Fatalf("expected validation issue for %s, got %v", test.field, err)
			}
		})
	}
}

func TestValidateHardwareRejectsDuplicateIdentityAcrossEntities(t *testing.T) {
	device := validationDevice()
	duplicateDevice := validationDevice()
	duplicateDevice.ID = "8b75c9a2-9e1b-4da2-9226-000000000002"
	duplicateDevice.DevEUI = device.DevEUI
	gateway := validationGateway()
	duplicateGateway := validationGateway()
	duplicateGateway.ID = "8b75c9a2-9e1b-4da2-9226-000000000005"
	duplicateGateway.GatewayEUI = gateway.GatewayEUI
	duplicateGateway.MacAddress = gateway.MacAddress

	err := ValidateHardware([]contracts.Device{device, duplicateDevice}, []contracts.Gateway{gateway, duplicateGateway})
	if err == nil || !strings.Contains(err.Error(), "must be unique") {
		t.Fatalf("expected duplicate identity errors, got %v", err)
	}
}

func validationDevice() contracts.Device {
	return contracts.Device{
		ID: "8b75c9a2-9e1b-4da2-9226-000000000001", Active: true, Name: "Validation Device", DevEUI: "70B3D57ED0000001", Class: contracts.ClassA, Activation: contracts.OOTA,
		OOTAConfig:     &contracts.OOTAConfig{JoinEUI: "70B3D57ED0000002", AppKey: "70B3D57ED00000000000000000000001"},
		LocationConfig: contracts.LocationConfig{Latitude: float32Pointer(37.5), Longitude: float32Pointer(15.1), Altitude: float32Pointer(100), Region: contracts.EU868},
		RX1Config:      contracts.RX1Config{Delay: intPointer(1), Duration: intPointer(1000), DataRateOffset: intPointer(0)},
		RX2Config:      contracts.RX2Config{Delay: intPointer(1), Duration: intPointer(1000), ChannelFrequency: 869525000, DataRate: intPointer(5), ACKTimeout: 2000},
		FrameConfig:    contracts.FrameConfig{FPort: 1, Retransmission: intPointer(1)},
		PayloadConfig:  contracts.PayloadConfig{UplinkInterval: 10, OversizedPayloadBehavior: contracts.Truncate, MType: contracts.UnconfirmedDataUp, Payload: "test"},
		AdvancedConfig: contracts.AdvancedConfig{AntennaRange: 1000},
	}
}

func validationGateway() contracts.Gateway {
	return contracts.Gateway{
		ID: "8b75c9a2-9e1b-4da2-9226-000000000003", Active: true, Name: "Validation Gateway", Type: contracts.Virtual, MacAddress: "02:00:00:10:00:01", GatewayEUI: "A840410001000101", KeepAlive: int32Pointer(30),
		Latitude: float32Pointer(37.5), Longitude: float32Pointer(15.1), Altitude: float32Pointer(100),
	}
}

func intPointer(value int) *int             { return &value }
func int32Pointer(value int32) *int32       { return &value }
func float32Pointer(value float32) *float32 { return &value }
func stringPointer(value string) *string    { return &value }
