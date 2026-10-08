package services

import (
	"os"
	"path/filepath"
	"testing"

	contracts "github.com/Giuseppe-Compagnone/lwn-contracts/generated"
)

func TestProfileConfigurationServiceReturnsDefaultsWhenConfigurationIsMissing(t *testing.T) {
	service := NewProfileConfigurationService(filepath.Join(t.TempDir(), "gateway-bridge.json"))

	config, err := service.GetGatewayBridge()
	if err != nil {
		t.Fatalf("get missing gateway bridge configuration: %v", err)
	}
	if config != defaultGatewayBridgeConfig {
		t.Fatalf("unexpected default gateway bridge configuration: %+v", config)
	}
}

func TestProfileConfigurationServicePersistsConfigurationAtomically(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "gateway-bridge.json")
	service := NewProfileConfigurationService(path)
	wanted := contracts.GatewayBridgeConfig{
		Enabled: true,
		Address: "bridge.example.test",
		Port:    1701,
	}

	updated, err := service.UpdateGatewayBridge(wanted)
	if err != nil {
		t.Fatalf("update gateway bridge configuration: %v", err)
	}
	if updated != wanted {
		t.Fatalf("unexpected updated configuration: %+v", updated)
	}

	loaded, err := service.GetGatewayBridge()
	if err != nil {
		t.Fatalf("reload gateway bridge configuration: %v", err)
	}
	if loaded != wanted {
		t.Fatalf("configuration was not persisted: %+v", loaded)
	}

	if _, err := os.Stat(path); err != nil {
		t.Fatalf("configuration file was not created: %v", err)
	}
}

func TestProfileConfigurationServiceNormalizesUnsafeValues(t *testing.T) {
	service := NewProfileConfigurationService(filepath.Join(t.TempDir(), "gateway-bridge.json"))

	updated, err := service.UpdateGatewayBridge(contracts.GatewayBridgeConfig{
		Enabled: true,
		Port:    0,
	})
	if err != nil {
		t.Fatalf("update invalid gateway bridge configuration: %v", err)
	}
	if updated.Address != defaultGatewayBridgeConfig.Address || updated.Port != defaultGatewayBridgeConfig.Port {
		t.Fatalf("configuration was not normalized: %+v", updated)
	}
}

func TestProfileConfigurationServiceRejectsMalformedConfiguration(t *testing.T) {
	path := filepath.Join(t.TempDir(), "gateway-bridge.json")
	if err := os.WriteFile(path, []byte("{"), 0600); err != nil {
		t.Fatal(err)
	}

	_, err := NewProfileConfigurationService(path).GetGatewayBridge()
	if err == nil {
		t.Fatal("expected malformed configuration to fail")
	}
}
