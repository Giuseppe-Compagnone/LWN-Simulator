package gateway

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"net"
	"testing"
	"time"

	contracts "github.com/Giuseppe-Compagnone/lwn-contracts/generated"
	"github.com/Giuseppe-Compagnone/lwn-engine/types"
)

func TestUDPAdapterImplementsSemtechIngressAndEgress(t *testing.T) {
	gatewaySocket, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 0})
	if err != nil {
		t.Fatalf("listen fake gateway: %v", err)
	}
	defer gatewaySocket.Close()

	gateway := testRealGateway(gatewaySocket.LocalAddr().(*net.UDPAddr).Port)
	factory := NewUDPFactory(UDPOptions{LocalAddress: "127.0.0.1:0", Timeout: time.Second})
	adapter, err := factory.NewGatewayAdapter(gateway)
	if err != nil {
		t.Fatalf("create adapter: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	defer adapter.Close()

	if err := adapter.Start(ctx); err != nil {
		t.Fatalf("start adapter: %v", err)
	}
	factory.mu.Lock()
	serverAddress := factory.runtime.conn.LocalAddr().(*net.UDPAddr)
	factory.mu.Unlock()

	token := [2]byte{0x12, 0x34}
	eui, _ := hexBytes(gateway.GatewayEUI)
	pullData := append(packetHeader(token, pullDataType), eui...)
	if _, err := gatewaySocket.WriteToUDP(pullData, serverAddress); err != nil {
		t.Fatalf("send PULL_DATA: %v", err)
	}
	readSemtechPacket(t, gatewaySocket, pullAckType)
	waitAdapterEvent(t, adapter.Events(), contracts.Connected)
	waitHeartbeat(t, adapter.Events())

	body, err := json.Marshal(pushDataPayload{RXPK: []rxPacket{{
		Data: base64.StdEncoding.EncodeToString([]byte("uplink")), Frequency: 868.3, DataRate: "SF7BW125",
		Modulation: "LORA", CodingRate: "4/5", RSSI: -72, SNR: 5.5, Size: 6,
	}}})
	if err != nil {
		t.Fatalf("encode PUSH_DATA: %v", err)
	}
	pushData := append(packetHeader([2]byte{0x22, 0x33}, pushDataType), eui...)
	pushData = append(pushData, body...)
	if _, err := gatewaySocket.WriteToUDP(pushData, serverAddress); err != nil {
		t.Fatalf("send PUSH_DATA: %v", err)
	}
	readSemtechPacket(t, gatewaySocket, pushAckType)
	select {
	case packet := <-adapter.Packets():
		if string(packet.Payload) != "uplink" || packet.Frequency != 868300000 || packet.RSSI != -72 || packet.SNR != 5.5 {
			t.Fatalf("unexpected ingress packet: %+v", packet)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for ingress packet")
	}

	if err := adapter.Send(ctx, types.GatewayPacket{Payload: []byte("downlink")}); err != nil {
		t.Fatalf("send PULL_RESP: %v", err)
	}
	response := readSemtechPacket(t, gatewaySocket, pullResponseType)
	var responseBody struct {
		TXPK txPacket `json:"txpk"`
	}
	if err := json.Unmarshal(response[4:], &responseBody); err != nil {
		t.Fatalf("decode PULL_RESP: %v", err)
	}
	decoded, err := base64.StdEncoding.DecodeString(responseBody.TXPK.Data)
	if err != nil || string(decoded) != "downlink" {
		t.Fatalf("unexpected egress payload: %+v", responseBody.TXPK)
	}

	target := time.Now().Add(time.Second).UTC().Truncate(time.Millisecond)
	if err := adapter.Send(ctx, types.GatewayPacket{Payload: []byte("timed-downlink"), TransmitAt: target}); err != nil {
		t.Fatalf("send timed PULL_RESP: %v", err)
	}
	timedResponse := readSemtechPacket(t, gatewaySocket, pullResponseType)
	var timedBody struct {
		TXPK txPacket `json:"txpk"`
	}
	if err := json.Unmarshal(timedResponse[4:], &timedBody); err != nil {
		t.Fatalf("decode timed PULL_RESP: %v", err)
	}
	if timedBody.TXPK.Immediate || timedBody.TXPK.Time == "" {
		t.Fatalf("expected timed packet-forwarder transmission: %+v", timedBody.TXPK)
	}
	if got, err := time.Parse(time.RFC3339Nano, timedBody.TXPK.Time); err != nil || !got.Equal(target) {
		t.Fatalf("unexpected timed transmission timestamp: got=%q target=%s err=%v", timedBody.TXPK.Time, target, err)
	}
	txAckBody := []byte(`{"txpk_ack":{"error":"TOO_LATE"}}`)
	// Semtech TX_ACK includes the gateway EUI before its optional JSON body.
	txAck := append([]byte{protocolVersion, timedResponse[1], timedResponse[2], txAckType}, eui...)
	txAck = append(txAck, txAckBody...)
	if _, err := gatewaySocket.WriteToUDP(txAck, serverAddress); err != nil {
		t.Fatalf("send TX_ACK: %v", err)
	}
	deadline := time.After(time.Second)
	for {
		select {
		case event := <-adapter.Events():
			if event.State != contracts.Error {
				continue
			}
			if event.Packet == nil || string(event.Packet.Payload) != "timed-downlink" {
				t.Fatalf("expected correlated TX_ACK error, got %+v", event)
			}
			return
		case <-deadline:
			t.Fatal("timed out waiting for correlated TX_ACK error")
		}
	}
}

func TestUDPAdapterAcceptsSemtechGatewayAcknowledgements(t *testing.T) {
	gatewaySocket, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 0})
	if err != nil {
		t.Fatalf("listen fake gateway: %v", err)
	}
	defer gatewaySocket.Close()

	gateway := testRealGateway(gatewaySocket.LocalAddr().(*net.UDPAddr).Port)
	factory := NewUDPFactory(UDPOptions{LocalAddress: "127.0.0.1:0", Timeout: time.Second})
	adapter, err := factory.NewGatewayAdapter(gateway)
	if err != nil {
		t.Fatalf("create adapter: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	defer adapter.Close()
	if err := adapter.Start(ctx); err != nil {
		t.Fatalf("start adapter: %v", err)
	}
	rawAdapter := adapter.(*udpAdapter)
	factory.mu.Lock()
	serverAddress := factory.runtime.conn.LocalAddr().(*net.UDPAddr)
	factory.mu.Unlock()
	eui, _ := hexBytes(gateway.GatewayEUI)

	pullToken := [2]byte{0x20, 0x21}
	if _, err := gatewaySocket.WriteToUDP(append(packetHeader(pullToken, pullAckType), nil...), serverAddress); err != nil {
		t.Fatalf("send PULL_ACK: %v", err)
	}
	waitHeartbeat(t, adapter.Events())

	pushToken := [2]byte{0x22, 0x23}
	pushAck := append(packetHeader(pushToken, pushAckType), nil...)
	if _, err := gatewaySocket.WriteToUDP(pushAck, serverAddress); err != nil {
		t.Fatalf("send PUSH_ACK: %v", err)
	}
	waitHeartbeat(t, adapter.Events())

	emptyTXAck := append(packetHeader([2]byte{0x24, 0x25}, txAckType), eui...)
	if err := rawAdapter.handleDatagram(ctx, udpDatagram{data: emptyTXAck, remote: gatewaySocket.LocalAddr().(*net.UDPAddr)}); err != nil {
		t.Fatalf("empty TX_ACK should be accepted: %v", err)
	}
}

func TestUDPAdapterRejectsUnsupportedOrInconsistentRXPK(t *testing.T) {
	factory := NewUDPFactory(UDPOptions{})
	adapter, err := factory.NewGatewayAdapter(testRealGateway(1700))
	if err != nil {
		t.Fatalf("create adapter: %v", err)
	}
	defer adapter.Close()
	rawAdapter := adapter.(*udpAdapter)
	ctx := context.Background()

	unsupported, err := json.Marshal(pushDataPayload{RXPK: []rxPacket{{
		Data: base64.StdEncoding.EncodeToString([]byte("uplink")), Frequency: 868.1,
		DataRate: "SF7BW125", Modulation: "FSK", Size: 6,
	}}})
	if err != nil {
		t.Fatalf("encode unsupported RXPK: %v", err)
	}
	if err := rawAdapter.handlePushData(ctx, unsupported); err == nil {
		t.Fatal("unsupported modulation should be rejected")
	}

	mismatch, err := json.Marshal(pushDataPayload{RXPK: []rxPacket{{
		Data: base64.StdEncoding.EncodeToString([]byte("uplink")), Frequency: 868.1,
		DataRate: "SF7BW125", Modulation: "LORA", Size: 99,
	}}})
	if err != nil {
		t.Fatalf("encode inconsistent RXPK: %v", err)
	}
	if err := rawAdapter.handlePushData(ctx, mismatch); err == nil {
		t.Fatal("inconsistent RXPK size should be rejected")
	}
}

func TestUDPAdapterRejectsTXAckWithMissingGatewayEUI(t *testing.T) {
	gatewaySocket, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 0})
	if err != nil {
		t.Fatalf("listen fake gateway: %v", err)
	}
	defer gatewaySocket.Close()

	gateway := testRealGateway(gatewaySocket.LocalAddr().(*net.UDPAddr).Port)
	factory := NewUDPFactory(UDPOptions{LocalAddress: "127.0.0.1:0", Timeout: time.Second})
	adapter, err := factory.NewGatewayAdapter(gateway)
	if err != nil {
		t.Fatalf("create adapter: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	defer adapter.Close()
	if err := adapter.Start(ctx); err != nil {
		t.Fatalf("start adapter: %v", err)
	}
	udpAdapter := adapter.(*udpAdapter)

	malformed := append(packetHeader([2]byte{0x30, 0x31}, txAckType), []byte(`{"txpk_ack":{"error":"NONE"}}`)...)
	if err := udpAdapter.handleDatagram(ctx, udpDatagram{data: malformed, remote: gatewaySocket.LocalAddr().(*net.UDPAddr)}); err == nil {
		t.Fatal("TX_ACK without the gateway EUI should be rejected")
	}
}

func TestUDPAdapterReportsTimeoutAndReconnectState(t *testing.T) {
	gatewaySocket, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 0})
	if err != nil {
		t.Fatalf("listen fake gateway: %v", err)
	}
	defer gatewaySocket.Close()

	gateway := testRealGateway(gatewaySocket.LocalAddr().(*net.UDPAddr).Port)
	factory := NewUDPFactory(UDPOptions{
		LocalAddress:     "127.0.0.1:0",
		Timeout:          50 * time.Millisecond,
		ReconnectInitial: 10 * time.Millisecond,
		ReconnectMax:     20 * time.Millisecond,
	})
	adapter, err := factory.NewGatewayAdapter(gateway)
	if err != nil {
		t.Fatalf("create adapter: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	defer adapter.Close()
	if err := adapter.Start(ctx); err != nil {
		t.Fatalf("start adapter: %v", err)
	}
	factory.mu.Lock()
	serverAddress := factory.runtime.conn.LocalAddr().(*net.UDPAddr)
	factory.mu.Unlock()
	eui, _ := hexBytes(gateway.GatewayEUI)
	if _, err := gatewaySocket.WriteToUDP(append(packetHeader([2]byte{1, 2}, pullDataType), eui...), serverAddress); err != nil {
		t.Fatalf("send PULL_DATA: %v", err)
	}
	readSemtechPacket(t, gatewaySocket, pullAckType)
	waitAdapterEvent(t, adapter.Events(), contracts.Connected)
	waitHeartbeat(t, adapter.Events())
	waitAdapterEvent(t, adapter.Events(), contracts.Error)
	waitAdapterEvent(t, adapter.Events(), contracts.Disconnected)
	waitAdapterEvent(t, adapter.Events(), contracts.Reconnecting)
}

func TestSemtechDataRateParsing(t *testing.T) {
	if got := parseSpreadingFactor("SF7BW125"); got != 7 {
		t.Fatalf("expected SF7, got %d", got)
	}
	if got := parseBandwidth("SF7BW125"); got != 125000 {
		t.Fatalf("expected 125kHz bandwidth, got %d", got)
	}
	if got := parseSpreadingFactor("FSK"); got != 0 {
		t.Fatalf("expected invalid spreading factor to return zero, got %d", got)
	}
}

func TestUDPFactoryConfiguresGatewayBridgeBeforeStart(t *testing.T) {
	factory := NewUDPFactory(UDPOptions{LocalAddress: "127.0.0.1:1700"})

	if err := factory.ConfigureGatewayBridge(contracts.GatewayBridgeConfig{
		Enabled: true,
		Address: "127.0.0.1",
		Port:    1700,
	}); err != nil {
		t.Fatalf("configure gateway bridge: %v", err)
	}
	if got := factory.LocalAddress(); got != "127.0.0.1:1700" {
		t.Fatalf("bridge configuration changed the local listen address: %q", got)
	}
	if !factory.SupportsVirtualGateways() {
		t.Fatal("expected virtual gateway transport to be enabled")
	}

	if err := factory.ConfigureGatewayBridge(contracts.GatewayBridgeConfig{
		Enabled: true,
		Address: "not/an/address",
		Port:    1700,
	}); err == nil {
		t.Fatal("invalid bridge address should be rejected")
	}
	if err := factory.ConfigureGatewayBridge(contracts.GatewayBridgeConfig{Enabled: false}); err != nil {
		t.Fatalf("disable gateway bridge: %v", err)
	}
	if factory.SupportsVirtualGateways() {
		t.Fatal("disabling the bridge should disable virtual gateway transport")
	}
}

func TestUDPVirtualGatewaySendsKeepaliveAndUplinkToBridge(t *testing.T) {
	bridgeSocket, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 0})
	if err != nil {
		t.Fatalf("listen fake gateway bridge: %v", err)
	}
	defer bridgeSocket.Close()

	factory := NewUDPFactory(UDPOptions{
		HeartbeatInterval: 50 * time.Millisecond,
		Timeout:           500 * time.Millisecond,
	})
	bridgeAddress := bridgeSocket.LocalAddr().(*net.UDPAddr)
	if err := factory.ConfigureGatewayBridge(contracts.GatewayBridgeConfig{
		Enabled: true, Address: "127.0.0.1", Port: int32(bridgeAddress.Port),
	}); err != nil {
		t.Fatalf("configure gateway bridge: %v", err)
	}
	virtualGateway := contracts.Gateway{
		ID: "virtual-gateway-test", Active: true, Name: "Virtual Test Gateway",
		Type: contracts.Virtual, GatewayEUI: "A840410001000102",
	}
	adapter, err := factory.NewGatewayAdapter(virtualGateway)
	if err != nil {
		t.Fatalf("create virtual adapter: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	defer adapter.Close()
	if err := adapter.Start(ctx); err != nil {
		t.Fatalf("start virtual adapter: %v", err)
	}

	keepalive := readSemtechPacket(t, bridgeSocket, pullDataType)
	if len(keepalive) < 12 {
		t.Fatalf("PULL_DATA did not contain gateway EUI: %x", keepalive)
	}
	if _, err := bridgeSocket.WriteToUDP(append(packetHeader([2]byte{keepalive[1], keepalive[2]}, pullAckType), nil...), mustUDPAddr(adapter)); err != nil {
		t.Fatalf("send PULL_ACK: %v", err)
	}
	waitAdapterEvent(t, adapter.Events(), contracts.Connected)

	uplink := types.GatewayPacket{
		GatewayID: "virtual-gateway-test", Payload: []byte("uplink"),
		Frequency: 868300000, Bandwidth: 125000, SpreadingFactor: 7,
	}
	if err := adapter.(types.GatewayUplinkAdapter).SendUplink(ctx, uplink); err != nil {
		t.Fatalf("send virtual uplink: %v", err)
	}
	packet := readSemtechPacket(t, bridgeSocket, pushDataType)
	if len(packet) < 12 {
		t.Fatalf("PUSH_DATA did not contain gateway EUI: %x", packet)
	}
	var body pushDataPayload
	if err := json.Unmarshal(packet[12:], &body); err != nil {
		t.Fatalf("decode PUSH_DATA: %v", err)
	}
	if len(body.RXPK) != 1 || body.RXPK[0].DataRate != "SF7BW125" || body.RXPK[0].Size != len(uplink.Payload) {
		t.Fatalf("unexpected virtual uplink metadata: %+v", body.RXPK)
	}
	decoded, err := base64.StdEncoding.DecodeString(body.RXPK[0].Data)
	if err != nil || string(decoded) != "uplink" {
		t.Fatalf("unexpected virtual uplink payload: %q (%v)", decoded, err)
	}
}

func mustUDPAddr(adapter types.GatewayAdapter) *net.UDPAddr {
	return adapter.(*udpAdapter).conn.LocalAddr().(*net.UDPAddr)
}

func TestUDPFactoryIgnoresDisabledGatewayBridge(t *testing.T) {
	factory := NewUDPFactory(UDPOptions{LocalAddress: "127.0.0.1:1700"})
	if err := factory.ConfigureGatewayBridge(contracts.GatewayBridgeConfig{
		Enabled: false,
		Address: "not-an-address",
		Port:    0,
	}); err != nil {
		t.Fatalf("disabled bridge configuration should be ignored: %v", err)
	}
	if got := factory.LocalAddress(); got != "127.0.0.1:1700" {
		t.Fatalf("disabled configuration changed listen address to %q", got)
	}
}

func TestUDPFactoryRejectsBridgeReconfigurationWhileRunning(t *testing.T) {
	gatewaySocket, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 0})
	if err != nil {
		t.Fatalf("listen fake gateway: %v", err)
	}
	defer gatewaySocket.Close()

	factory := NewUDPFactory(UDPOptions{LocalAddress: "127.0.0.1:0"})
	adapter, err := factory.NewGatewayAdapter(testRealGateway(gatewaySocket.LocalAddr().(*net.UDPAddr).Port))
	if err != nil {
		t.Fatalf("create adapter: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	defer adapter.Close()
	if err := adapter.Start(ctx); err != nil {
		t.Fatalf("start adapter: %v", err)
	}

	if err := factory.ConfigureGatewayBridge(contracts.GatewayBridgeConfig{
		Enabled: true,
		Address: "127.0.0.1",
		Port:    1701,
	}); err == nil {
		t.Fatal("running factory should reject bridge reconfiguration")
	}
}

func testRealGateway(port int) contracts.Gateway {
	return contracts.Gateway{
		ID: "6f0f1f30-7dc5-4bb3-a6f5-11b2ed9a7c01", Active: true, Name: "Real Test Gateway",
		Type: contracts.Real, MacAddress: "02:00:00:10:00:01", GatewayEUI: "A840410001000101",
		GatewayIPv4: gatewayStringPtr("127.0.0.1"), GatewayPort: gatewayInt32Ptr(int32(port)),
		Latitude: gatewayFloat32Ptr(37.5), Longitude: gatewayFloat32Ptr(15.1), Altitude: gatewayFloat32Ptr(100),
	}
}

func gatewayStringPtr(value string) *string { return &value }

func gatewayInt32Ptr(value int32) *int32 { return &value }

func gatewayFloat32Ptr(value float32) *float32 { return &value }

func hexBytes(value string) ([]byte, error) {
	return hex.DecodeString(value)
}

func readSemtechPacket(t *testing.T, socket *net.UDPConn, expectedType byte) []byte {
	t.Helper()
	if err := socket.SetReadDeadline(time.Now().Add(time.Second)); err != nil {
		t.Fatalf("set fake gateway deadline: %v", err)
	}
	buffer := make([]byte, 4096)
	size, _, err := socket.ReadFromUDP(buffer)
	if err != nil {
		t.Fatalf("read Semtech packet: %v", err)
	}
	packet := buffer[:size]
	if len(packet) < 4 || packet[0] != protocolVersion || packet[3] != expectedType {
		t.Fatalf("unexpected Semtech packet: %x", packet)
	}
	return packet
}

func waitAdapterEvent(t *testing.T, events <-chan types.GatewayAdapterEvent, state contracts.GatewayConnectionState) {
	t.Helper()
	deadline := time.After(time.Second)
	for {
		select {
		case event := <-events:
			if event.State == state {
				return
			}
		case <-deadline:
			t.Fatalf("timed out waiting for gateway state %q", state)
		}
	}
}

func waitHeartbeat(t *testing.T, events <-chan types.GatewayAdapterEvent) {
	t.Helper()
	deadline := time.After(time.Second)
	for {
		select {
		case event := <-events:
			if event.Heartbeat {
				return
			}
		case <-deadline:
			t.Fatal("timed out waiting for gateway heartbeat")
		}
	}
}
