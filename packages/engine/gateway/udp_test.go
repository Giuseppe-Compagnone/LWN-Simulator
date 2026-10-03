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
		if string(packet.Payload) != "uplink" || packet.Frequency != 868300000 {
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
