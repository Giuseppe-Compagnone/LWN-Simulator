package engine

import (
	"context"
	"testing"
	"time"

	contracts "github.com/Giuseppe-Compagnone/lwn-contracts/generated"
	"github.com/Giuseppe-Compagnone/lwn-engine/types"
)

func TestPhaseFiveEngineTracksRealGatewayLifecycleAndPackets(t *testing.T) {
	adapter := newFakeGatewayAdapter()
	factory := &fakeGatewayFactory{adapter: adapter}
	gateway := validGateway("6f0f1f30-7dc5-4bb3-a6f5-11b2ed9a7c01")
	gateway.Type = contracts.Real
	gateway.KeepAlive = nil
	gateway.GatewayEUI = "A840410001000101"
	gateway.MacAddress = "02:00:00:10:00:02"
	gateway.GatewayIPv4 = engineStringPtr("127.0.0.1")
	gateway.GatewayPort = engineInt32Ptr(1700)

	engine, err := New(
		contracts.SimulationConfig{Speed: 1},
		nil,
		[]contracts.Gateway{gateway},
		types.Options{EventBuffer: 256, GatewayAdapterFactory: factory},
	)
	if err != nil {
		t.Fatalf("create engine: %v", err)
	}
	startEngine(t, engine)
	defer stopEngine(t, engine)

	waitForEventType(t, engine.Events(), contracts.GatewayConnecting)
	adapter.emit(types.GatewayAdapterEvent{GatewayID: gateway.ID, State: contracts.Connected})
	waitForEventType(t, engine.Events(), contracts.GatewayConnected)
	adapter.emit(types.GatewayAdapterEvent{GatewayID: gateway.ID, State: contracts.Connected, Heartbeat: true})
	waitForEventType(t, engine.Events(), contracts.GatewayHeartbeat)

	payload := []byte("downlink")
	if err := engine.SendGatewayPacket(context.Background(), gateway.ID, payload); err != nil {
		t.Fatalf("send gateway packet: %v", err)
	}
	select {
	case packet := <-adapter.sends:
		if string(packet.Payload) != string(payload) {
			t.Fatalf("unexpected egress payload: %q", packet.Payload)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for fake gateway egress")
	}
	waitForEventType(t, engine.Events(), contracts.GatewayPacketEgress)

	adapter.packets <- types.GatewayPacket{GatewayID: gateway.ID, Payload: []byte("uplink")}
	waitForEventType(t, engine.Events(), contracts.GatewayPacketIngress)
	adapter.emit(types.GatewayAdapterEvent{
		GatewayID: gateway.ID,
		State:     contracts.Error,
		Error:     "heartbeat timeout",
		Timeout:   true,
	})
	waitForEventType(t, engine.Events(), contracts.GatewayNetworkError)

	snapshot := engine.Snapshot()
	if snapshot.Gateways[0].GatewayState != contracts.Error || snapshot.Gateways[0].LastNetworkError != "heartbeat timeout" {
		t.Fatalf("unexpected real gateway runtime: %+v", snapshot.Gateways[0])
	}
	if snapshot.Gateways[0].IngressPackets != 1 || snapshot.Gateways[0].EgressPackets != 1 {
		t.Fatalf("unexpected gateway packet counters: %+v", snapshot.Gateways[0])
	}
	if snapshot.Metrics.GatewayConnections != 1 || snapshot.Metrics.GatewayHeartbeats != 1 ||
		snapshot.Metrics.GatewayIngressPackets != 1 || snapshot.Metrics.GatewayEgressPackets != 1 || snapshot.Metrics.GatewayTimeouts != 1 {
		t.Fatalf("unexpected gateway metrics: %+v", snapshot.Metrics)
	}
}

type fakeGatewayFactory struct {
	adapter *fakeGatewayAdapter
}

func (factory *fakeGatewayFactory) NewGatewayAdapter(contracts.Gateway) (types.GatewayAdapter, error) {
	return factory.adapter, nil
}

type fakeGatewayAdapter struct {
	events  chan types.GatewayAdapterEvent
	packets chan types.GatewayPacket
	sends   chan types.GatewayPacket
	closed  chan struct{}
}

func newFakeGatewayAdapter() *fakeGatewayAdapter {
	return &fakeGatewayAdapter{
		events:  make(chan types.GatewayAdapterEvent, 32),
		packets: make(chan types.GatewayPacket, 32),
		sends:   make(chan types.GatewayPacket, 32),
		closed:  make(chan struct{}),
	}
}

func (adapter *fakeGatewayAdapter) Start(context.Context) error { return nil }

func (adapter *fakeGatewayAdapter) Send(_ context.Context, packet types.GatewayPacket) error {
	adapter.sends <- packet
	return nil
}

func (adapter *fakeGatewayAdapter) Packets() <-chan types.GatewayPacket {
	return adapter.packets
}

func (adapter *fakeGatewayAdapter) Events() <-chan types.GatewayAdapterEvent {
	return adapter.events
}

func (adapter *fakeGatewayAdapter) Close() error {
	select {
	case <-adapter.closed:
	default:
		close(adapter.closed)
		close(adapter.events)
		close(adapter.packets)
	}
	return nil
}

func (adapter *fakeGatewayAdapter) emit(event types.GatewayAdapterEvent) {
	adapter.events <- event
}

func engineStringPtr(value string) *string { return &value }

func engineInt32Ptr(value int32) *int32 { return &value }
