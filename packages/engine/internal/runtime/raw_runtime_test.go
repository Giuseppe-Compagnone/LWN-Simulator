package runtime

import (
	"testing"
	"time"

	contracts "github.com/Giuseppe-Compagnone/lwn-contracts/generated"
	"github.com/Giuseppe-Compagnone/lwn-engine/types"
)

func TestRawDeviceTransmitsWithoutLoRaWANJoin(t *testing.T) {
	clock := types.NewManualClock(0)
	device := validDevice("a1c6e32b-4f0d-4b50-9fc5-000000000701")
	raw := contracts.RawLoRa
	device.RadioProtocol = &raw
	device.RadioReceiveMode = receiveModePointer(contracts.ReceiveRawLoRa)
	device.PayloadConfig.Payload = "raw payload"
	gateway := validGateway("6f0f1f30-7dc5-4bb3-a6f5-11b2ed9a7c71")

	engine := newTestEngine(t, clock, device, gateway)
	defer stopEngine(t, engine)

	started := waitForEventType(t, engine.Events(), contracts.RadioTransmissionStarted)
	if started.RadioFrame == nil || started.RadioFrame.Protocol != contracts.RawLoRa {
		t.Fatalf("expected raw radio frame, got %+v", started.RadioFrame)
	}
	if string(started.RadioFrame.Payload) != "raw payload" {
		t.Fatalf("unexpected raw payload: %q", started.RadioFrame.Payload)
	}

	if err := clock.Advance(time.Second); err != nil {
		t.Fatalf("advance clock: %v", err)
	}
	received := waitForEventType(t, engine.Events(), contracts.GatewayPacketReceived)
	if received.RadioFrame == nil || received.RadioFrame.Protocol != contracts.RawLoRa {
		t.Fatalf("expected raw protocol on gateway event, got %+v", received.RadioFrame)
	}
	if engine.Snapshot().Devices[0].Joined {
		t.Fatal("raw device must not require a LoRaWAN join")
	}
}

func TestRawDownlinkIsAcceptedOnlyByRawReceivingDevice(t *testing.T) {
	device := validDevice("a1c6e32b-4f0d-4b50-9fc5-000000000702")
	raw := contracts.RawLoRa
	device.RadioProtocol = &raw
	device.RadioReceiveMode = receiveModePointer(contracts.ReceiveRawLoRa)
	engine, err := New(contracts.SimulationConfig{Speed: 1}, []contracts.Device{device}, nil, types.Options{})
	if err != nil {
		t.Fatalf("create engine: %v", err)
	}

	engine.mu.Lock()
	events := engine.processRawDownlinkLocked(types.GatewayPacket{
		GatewayID: "gateway",
		DeviceID:  device.ID,
		Protocol:  contracts.RawLoRa,
		Kind:      types.GatewayPacketDownlink,
		Payload:   []byte{0x01, 0x02},
	})
	engine.mu.Unlock()
	if len(events) != 1 || events[0].Type != contracts.DeviceDownlinkTransmitted {
		t.Fatalf("raw downlink was not delivered: %+v", events)
	}
	if events[0].RadioFrame == nil || events[0].RadioFrame.Protocol != contracts.RawLoRa {
		t.Fatalf("raw downlink protocol was not exposed: %+v", events[0].RadioFrame)
	}

	lorawanMode := contracts.ReceiveLoRaWAN
	device.RadioReceiveMode = &lorawanMode
	registry, err := NewRegistry([]contracts.Device{device}, nil)
	if err != nil {
		t.Fatalf("rebuild registry: %v", err)
	}
	engine.registry = registry
	engine.mu.Lock()
	rejected := engine.processRawDownlinkLocked(types.GatewayPacket{
		GatewayID: "gateway",
		DeviceID:  device.ID,
		Protocol:  contracts.RawLoRa,
		Kind:      types.GatewayPacketDownlink,
		Payload:   []byte{0x01},
	})
	engine.mu.Unlock()
	if len(rejected) != 1 || rejected[0].Type != contracts.DeviceDownlinkDropped {
		t.Fatalf("raw downlink should be rejected by LoRaWAN-only device: %+v", rejected)
	}
}

func receiveModePointer(mode contracts.RadioReceiveMode) *contracts.RadioReceiveMode {
	return &mode
}
