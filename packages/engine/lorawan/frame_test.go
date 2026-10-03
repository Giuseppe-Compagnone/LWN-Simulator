package lorawan

import (
	"strings"
	"testing"
)

func TestBuildAndParseDataFrame(t *testing.T) {
	nwkSKey := []byte{0x00, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff}
	appSKey := []byte{0xff, 0xee, 0xdd, 0xcc, 0xbb, 0xaa, 0x99, 0x88, 0x77, 0x66, 0x55, 0x44, 0x33, 0x22, 0x11, 0x00}
	fPort := byte(10)
	frame, err := BuildDataFrame(DataFrameOptions{
		DevAddr:   0x26011bda,
		FCnt:      7,
		FPort:     &fPort,
		Payload:   []byte("payload"),
		Confirmed: true,
		NwkSKey:   nwkSKey,
		AppSKey:   appSKey,
	})
	if err != nil {
		t.Fatalf("build data frame: %v", err)
	}
	packet, err := Parse(frame)
	if err != nil {
		t.Fatalf("parse data frame: %v", err)
	}
	if packet.MType != MTypeConfirmedDataUp || packet.DevAddr != 0x26011bda || packet.FCnt != 7 || packet.FPort == nil || *packet.FPort != fPort {
		var port byte
		if packet.FPort != nil {
			port = *packet.FPort
		}
		t.Fatalf("unexpected parsed frame: type=%d devaddr=%x fcnt=%d port=%d want=%d packet=%+v", packet.MType, packet.DevAddr, packet.FCnt, port, fPort, packet)
	}
	if !VerifyDataMIC(packet, nwkSKey, 0) {
		t.Fatal("frame MIC did not validate")
	}
}

func TestDeriveSessionKeysIsDeterministic(t *testing.T) {
	key := make([]byte, 16)
	firstNwk, firstApp, err := DeriveSessionKeys(key, 1, 2, 3)
	if err != nil {
		t.Fatalf("derive first session keys: %v", err)
	}
	secondNwk, secondApp, err := DeriveSessionKeys(key, 1, 2, 3)
	if err != nil {
		t.Fatalf("derive second session keys: %v", err)
	}
	for index := range firstNwk {
		if firstNwk[index] != secondNwk[index] || firstApp[index] != secondApp[index] {
			t.Fatal("session key derivation is not deterministic")
		}
	}
}

func TestBuildJoinAcceptCreatesEncryptedPayload(t *testing.T) {
	frame, err := BuildJoinAccept(JoinAcceptOptions{
		AppNonce:          1,
		NetID:             2,
		DevAddr:           0x26011bda,
		RX1DataRate:       5,
		RX1DataRateOffset: 1,
		RXDelay:           1,
		AppKey:            make([]byte, 16),
	})
	if err != nil {
		t.Fatalf("build join accept: %v", err)
	}
	if len(frame) != 17 || frame[0] != MTypeJoinAccept<<5 {
		t.Fatalf("unexpected join accept frame: len=%d mhdr=%x", len(frame), frame[0])
	}
	if string(frame[1:]) == string(make([]byte, 16)) {
		t.Fatal("join accept payload was not encrypted")
	}
}

func TestBuildJoinRequestCreatesVerifiableFrame(t *testing.T) {
	key := make([]byte, 16)
	frame, err := BuildJoinRequest(JoinRequestOptions{JoinEUI: 1, DevEUI: 2, DevNonce: 3, AppKey: key})
	if err != nil {
		t.Fatalf("build join request: %v", err)
	}
	packet, err := Parse(frame)
	if err != nil || packet.JoinEUI != 1 || packet.DevEUI != 2 || packet.DevNonce != 3 || !VerifyJoinRequestMIC(packet, key) {
		t.Fatalf("unexpected join request: packet=%+v error=%v", packet, err)
	}
}

func TestLoRaWANCodecRejectsMalformedInputs(t *testing.T) {
	if _, err := Parse([]byte{0x40}); err == nil {
		t.Fatal("expected short frame to be rejected")
	}
	if _, err := Parse([]byte{0x00, 1, 2, 3, 4}); err == nil {
		t.Fatal("expected malformed join request to be rejected")
	}
	if _, err := Parse([]byte{0xe0, 1, 2, 3, 4}); err != nil {
		t.Fatalf("unknown frame type should remain parseable: %v", err)
	}
	if _, err := BuildDataFrame(DataFrameOptions{NwkSKey: []byte{1}, AppSKey: make([]byte, 16)}); err == nil {
		t.Fatal("expected invalid key length to be rejected")
	}
	key := make([]byte, 16)
	fPort := byte(0)
	if _, err := BuildDataFrame(DataFrameOptions{FPort: &fPort, NwkSKey: key, AppSKey: key}); err == nil {
		t.Fatal("expected invalid FPort to be rejected")
	}
	if _, err := BuildDataFrame(DataFrameOptions{Payload: []byte{1}, NwkSKey: key, AppSKey: key}); err == nil {
		t.Fatal("expected payload without FPort to be rejected")
	}
	if _, err := BuildDataFrame(DataFrameOptions{Direction: 2, NwkSKey: key, AppSKey: key}); err == nil {
		t.Fatal("expected invalid direction to be rejected")
	}
	if _, err := BuildJoinAccept(JoinAcceptOptions{AppKey: key, RX1DataRate: 16}); err == nil {
		t.Fatal("expected invalid Join-Accept settings to be rejected")
	}
	if _, err := DecodeHex("not-hex"); err == nil || !strings.Contains(err.Error(), "decode") {
		t.Fatalf("expected invalid hex error, got %v", err)
	}
	if _, _, err := DeriveSessionKeys(key[:1], 0, 0, 0); err == nil {
		t.Fatal("expected invalid session-key input to be rejected")
	}
}

func TestLoRaWANJoinRequestMICValidation(t *testing.T) {
	key := make([]byte, 16)
	raw := make([]byte, 23)
	raw[0] = MTypeJoinRequest << 5
	raw[1] = 1
	raw[9] = 2
	raw[17] = 3
	mac, err := cmac(key, raw[:19])
	if err != nil {
		t.Fatalf("compute join MIC: %v", err)
	}
	copy(raw[19:], mac[:4])
	packet, err := Parse(raw)
	if err != nil || !VerifyJoinRequestMIC(packet, key) {
		t.Fatalf("valid join request was rejected: %v", err)
	}
	raw[19] ^= 0xff
	packet, _ = Parse(raw)
	if VerifyJoinRequestMIC(packet, key) {
		t.Fatal("tampered join request MIC was accepted")
	}
	if VerifyJoinRequestMIC(Packet{}, key[:1]) {
		t.Fatal("invalid join request inputs were accepted")
	}
}
