package engine

import (
	"strings"
	"testing"

	contracts "github.com/Giuseppe-Compagnone/lwn-contracts/generated"
)

func TestValidateSimulationConfigRejectsInvalidSpeed(t *testing.T) {
	err := ValidateSimulationConfig(contracts.SimulationConfig{Speed: 0})
	if err == nil || !strings.Contains(err.Error(), "speed") {
		t.Fatalf("expected speed validation error, got %v", err)
	}
}

func TestValidateHardwareRejectsInconsistentGateway(t *testing.T) {
	gateway := validGateway("6f0f1f30-7dc5-4bb3-a6f5-11b2ed9a7c01")
	gateway.Type = contracts.Real
	gateway.KeepAlive = int32Ptr(30)
	gateway.GatewayIPv4 = nil
	gateway.GatewayPort = nil

	err := ValidateHardware(nil, []contracts.Gateway{gateway})
	if err == nil {
		t.Fatal("expected real gateway validation to fail")
	}
	if !strings.Contains(err.Error(), "gatewayIPv4") || !strings.Contains(err.Error(), "keepAlive") {
		t.Fatalf("expected endpoint validation errors, got %v", err)
	}
}

func TestValidateHardwareRejectsDuplicateIdentity(t *testing.T) {
	device := validDevice("a1c6e32b-4f0d-4b50-9fc5-000000000001")
	duplicate := validDevice("a1c6e32b-4f0d-4b50-9fc5-000000000002")
	duplicate.DevEUI = device.DevEUI

	err := ValidateHardware([]contracts.Device{device, duplicate}, nil)
	if err == nil || !strings.Contains(err.Error(), "DevEUI") {
		t.Fatalf("expected duplicate DevEUI error, got %v", err)
	}
}

func intPtr(value int) *int { return &value }

func int64Ptr(value int64) *int64 { return &value }

func int32Ptr(value int32) *int32 { return &value }

func float32Ptr(value float32) *float32 { return &value }

func optionalCounter(value int) **int {
	inner := value
	outer := &inner
	return &outer
}

func validDevice(id string) contracts.Device {
	return contracts.Device{
		ID:         id,
		Active:     true,
		Name:       "Test Device",
		DevEUI:     "70B3D57ED0000001",
		Class:      contracts.ClassA,
		Activation: contracts.OOTA,
		OOTAConfig: &contracts.OOTAConfig{
			JoinEUI: "70B3D57ED0000001",
			AppKey:  "70B3D57ED00000000000000000000001",
		},
		LocationConfig: contracts.LocationConfig{
			Latitude:  float32Ptr(37.5),
			Longitude: float32Ptr(15.1),
			Altitude:  float32Ptr(100),
			Region:    contracts.EU868,
		},
		RX1Config: contracts.RX1Config{
			Delay:          intPtr(1),
			Duration:       intPtr(1000),
			DataRateOffset: intPtr(0),
		},
		RX2Config: contracts.RX2Config{
			Delay:            intPtr(1),
			Duration:         intPtr(1000),
			ChannelFrequency: 869525000,
			DataRate:         intPtr(0),
			ACKTimeout:       2000,
		},
		FrameConfig: contracts.FrameConfig{
			FPort:                         1,
			Retransmission:                intPtr(0),
			FCntUp:                        optionalCounter(0),
			FCntDown:                      optionalCounter(0),
			DisableFrameCounterValidation: false,
		},
		PayloadConfig: contracts.PayloadConfig{
			UplinkInterval:           10,
			OversizedPayloadBehavior: contracts.Truncate,
			MType:                    contracts.UnconfirmedDataUp,
			Payload:                  "test",
			Base64Encoded:            false,
		},
		AdvancedConfig: contracts.AdvancedConfig{
			AntennaRange: 1000,
			ADREnabled:   true,
		},
	}
}

func validGateway(id string) contracts.Gateway {
	return contracts.Gateway{
		ID:         id,
		Active:     true,
		Name:       "Test Gateway",
		Type:       contracts.Virtual,
		MacAddress: "02:00:00:10:00:01",
		GatewayEUI: "A840410001000101",
		KeepAlive:  int32Ptr(30),
		Latitude:   float32Ptr(37.5),
		Longitude:  float32Ptr(15.1),
		Altitude:   float32Ptr(100),
	}
}
