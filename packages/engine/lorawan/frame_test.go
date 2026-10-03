package lorawan

import "testing"

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
