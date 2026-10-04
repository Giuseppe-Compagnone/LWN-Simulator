package runtime

import (
	"context"
	"strconv"
	"testing"
	"time"

	contracts "github.com/Giuseppe-Compagnone/lwn-contracts/generated"
	"github.com/Giuseppe-Compagnone/lwn-engine/lorawan"
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

func TestPhaseFiveRealUplinkUpdatesDeviceAndSchedulesConfirmedACK(t *testing.T) {
	adapter := newFakeGatewayAdapter()
	device := validDevice("a1c6e32b-4f0d-4b50-9fc5-000000000041")
	configureABP(&device)
	device.PayloadConfig.MType = contracts.ConfirmedDataUp
	gateway := validGateway("6f0f1f30-7dc5-4bb3-a6f5-11b2ed9a7c41")
	gateway.Type = contracts.Real
	gateway.KeepAlive = nil
	gateway.GatewayEUI = "A840410001000141"
	gateway.MacAddress = "02:00:00:10:00:41"
	gateway.GatewayIPv4 = engineStringPtr("127.0.0.1")
	gateway.GatewayPort = engineInt32Ptr(1741)

	engine, err := New(
		contracts.SimulationConfig{Speed: 1},
		[]contracts.Device{device},
		[]contracts.Gateway{gateway},
		types.Options{EventBuffer: 256, GatewayAdapterFactory: &fakeGatewayFactory{adapter: adapter}},
	)
	if err != nil {
		t.Fatalf("create engine: %v", err)
	}
	startEngine(t, engine)
	defer stopEngine(t, engine)
	waitForEventType(t, engine.Events(), contracts.GatewayConnecting)
	adapter.emit(types.GatewayAdapterEvent{GatewayID: gateway.ID, State: contracts.Connected})
	waitForEventType(t, engine.Events(), contracts.GatewayConnected)

	nwkSKey, _ := lorawan.DecodeHex(device.ABPConfig.NwkSKey)
	appSKey, _ := lorawan.DecodeHex(device.ABPConfig.AppSKey)
	fPort := byte(device.FrameConfig.FPort)
	payload, err := lorawan.BuildDataFrame(lorawan.DataFrameOptions{
		DevAddr:   0x26011bda,
		FCnt:      1,
		FPort:     &fPort,
		Payload:   []byte("uplink"),
		Confirmed: true,
		NwkSKey:   nwkSKey,
		AppSKey:   appSKey,
	})
	if err != nil {
		t.Fatalf("build real uplink: %v", err)
	}
	adapter.packets <- types.GatewayPacket{
		GatewayID: gateway.ID,
		Payload:   payload,
		Frequency: 868100000,
		Bandwidth: 125000,
		DataRate:  "SF7BW125",
	}
	waitForEventTypeWithLog(t, engine, contracts.UplinkACKReceived)
	if snapshot := engine.Snapshot(); snapshot.Devices[0].FrameCounterUp < 1 || snapshot.Metrics.SuccessfulUplinks != 1 {
		t.Fatalf("real uplink did not update runtime state: %+v", snapshot)
	}
	select {
	case ack := <-adapter.sends:
		decoded, parseErr := lorawan.Parse(ack.Payload)
		if parseErr != nil || decoded.MType != lorawan.MTypeUnconfirmedDataDown {
			t.Fatalf("unexpected real ACK frame: %+v", ack)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for real gateway ACK")
	}
}

func TestRealClassADownlinkCarriesACKFPendingAndMACCommands(t *testing.T) {
	adapter := newFakeGatewayAdapter()
	device := validDevice("a1c6e32b-4f0d-4b50-9fc5-000000000043")
	configureABP(&device)
	gateway := validGateway("6f0f1f30-7dc5-4bb3-a6f5-11b2ed9a7c43")
	gateway.Type = contracts.Real
	gateway.KeepAlive = nil
	gateway.GatewayEUI = "A840410001000143"
	gateway.MacAddress = "02:00:00:10:00:43"
	gateway.GatewayIPv4 = engineStringPtr("127.0.0.1")
	gateway.GatewayPort = engineInt32Ptr(1743)
	engine, err := New(
		contracts.SimulationConfig{Speed: 1},
		[]contracts.Device{device},
		[]contracts.Gateway{gateway},
		types.Options{EventBuffer: 256, GatewayAdapterFactory: &fakeGatewayFactory{adapter: adapter}},
	)
	if err != nil {
		t.Fatalf("create engine: %v", err)
	}
	startEngine(t, engine)
	defer stopEngine(t, engine)
	waitForEventType(t, engine.Events(), contracts.GatewayConnecting)
	adapter.emit(types.GatewayAdapterEvent{GatewayID: gateway.ID, State: contracts.Connected})
	waitForEventType(t, engine.Events(), contracts.GatewayConnected)

	dataRate, repetitions := 2, 3
	if _, err := engine.QueueDownlink(types.Downlink{
		DeviceID: device.ID, Payload: []byte("command"), Confirmed: true, DataRate: -1,
		MACCommands: []types.MACCommand{{Type: types.MACLinkADRReq, DataRate: &dataRate, NbTrans: &repetitions}},
	}); err != nil {
		t.Fatalf("queue first downlink: %v", err)
	}
	if _, err := engine.QueueDownlink(types.Downlink{DeviceID: device.ID, Payload: []byte("second"), DataRate: -1}); err != nil {
		t.Fatalf("queue second downlink: %v", err)
	}

	nwkSKey, _ := lorawan.DecodeHex(device.ABPConfig.NwkSKey)
	appSKey, _ := lorawan.DecodeHex(device.ABPConfig.AppSKey)
	fPort := byte(device.FrameConfig.FPort)
	uplink, err := lorawan.BuildDataFrame(lorawan.DataFrameOptions{
		DevAddr: 0x26011bda, FCnt: 1, FPort: &fPort, Payload: []byte("uplink"),
		Confirmed: true, NwkSKey: nwkSKey, AppSKey: appSKey,
	})
	if err != nil {
		t.Fatalf("build uplink: %v", err)
	}
	adapter.packets <- types.GatewayPacket{
		GatewayID: gateway.ID, Payload: uplink, Frequency: 868_100_000,
		Bandwidth: 125_000, SpreadingFactor: 7, DataRate: "SF7BW125",
	}
	waitForEventTypeWithLog(t, engine, contracts.DeviceDownlinkTransmitted)
	select {
	case sent := <-adapter.sends:
		packet, parseErr := lorawan.Parse(sent.Payload)
		if parseErr != nil {
			t.Fatalf("parse downlink: %v", parseErr)
		}
		if !packet.ACK || !packet.Confirmed || !packet.FPending || len(packet.FOpts) != 5 || packet.FOpts[0] != 0x03 {
			t.Fatalf("unexpected Class A downlink controls: %+v FOpts=%x", packet, packet.FOpts)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for real Class A downlink")
	}
	engine.mu.RLock()
	session := *engine.sessions[device.ID]
	engine.mu.RUnlock()
	if session.CurrentDataRate != dataRate || session.UnconfirmedRepetitions != repetitions {
		t.Fatalf("MAC command was not applied: %+v", session)
	}
}

func TestPhaseFiveRealOTAAJoinQueuesEncryptedJoinAccept(t *testing.T) {
	adapter := newFakeGatewayAdapter()
	device := validDevice("a1c6e32b-4f0d-4b50-9fc5-000000000042")
	gateway := validGateway("6f0f1f30-7dc5-4bb3-a6f5-11b2ed9a7c42")
	gateway.Type = contracts.Real
	gateway.KeepAlive = nil
	gateway.GatewayEUI = "A840410001000142"
	gateway.MacAddress = "02:00:00:10:00:42"
	gateway.GatewayIPv4 = engineStringPtr("127.0.0.1")
	gateway.GatewayPort = engineInt32Ptr(1742)

	engine, err := New(
		contracts.SimulationConfig{Speed: 1},
		[]contracts.Device{device},
		[]contracts.Gateway{gateway},
		types.Options{EventBuffer: 256, GatewayAdapterFactory: &fakeGatewayFactory{adapter: adapter}},
	)
	if err != nil {
		t.Fatalf("create engine: %v", err)
	}
	startEngine(t, engine)
	defer stopEngine(t, engine)
	waitForEventType(t, engine.Events(), contracts.GatewayConnecting)
	adapter.emit(types.GatewayAdapterEvent{GatewayID: gateway.ID, State: contracts.Connected})
	waitForEventType(t, engine.Events(), contracts.GatewayConnected)

	joinEUI, _ := strconv.ParseUint(device.OOTAConfig.JoinEUI, 16, 64)
	devEUI, _ := strconv.ParseUint(device.DevEUI, 16, 64)
	appKey, _ := lorawan.DecodeHex(device.OOTAConfig.AppKey)
	joinRequest, err := lorawan.BuildJoinRequest(lorawan.JoinRequestOptions{JoinEUI: joinEUI, DevEUI: devEUI, DevNonce: 7, AppKey: appKey})
	if err != nil {
		t.Fatalf("build real join request: %v", err)
	}
	adapter.packets <- types.GatewayPacket{GatewayID: gateway.ID, Payload: joinRequest, Frequency: 868100000, Bandwidth: 125000, DataRate: "SF7BW125"}
	waitForEventType(t, engine.Events(), contracts.JoinAcceptReceived)

	select {
	case packet := <-adapter.sends:
		if packet.Kind != types.GatewayPacketDownlink || len(packet.Payload) != 17 || packet.Payload[0] != lorawan.MTypeJoinAccept<<5 {
			t.Fatalf("unexpected real Join-Accept packet: %+v", packet)
		}
		if packet.TransmitAt.IsZero() {
			t.Fatal("Join-Accept packet is not scheduled for RX1")
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for real Join-Accept")
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
