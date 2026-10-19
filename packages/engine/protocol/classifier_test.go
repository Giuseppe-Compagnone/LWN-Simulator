package protocol

import (
	"bytes"
	"testing"

	contracts "github.com/Giuseppe-Compagnone/lwn-contracts/generated"
)

func TestClassifyLoRaWANFrameInLoRaWANMode(t *testing.T) {
	frame, err := (LoRaWANHandler{}).EncodeDataFrame(DataFrame{
		DevAddr:   0x26011bda,
		FCnt:      4,
		FPort:     bytePointer(1),
		Payload:   []byte("uplink"),
		Direction: 0,
		NwkSKey:   make([]byte, 16),
		AppSKey:   make([]byte, 16),
	})
	if err != nil {
		t.Fatalf("encode LoRaWAN frame: %v", err)
	}

	classified, err := Classify(frame, contracts.ReceiveLoRaWAN)
	if err != nil {
		t.Fatalf("classify LoRaWAN frame: %v", err)
	}
	if classified.Protocol != contracts.LoRaWAN || classified.Packet.DevAddr != 0x26011bda {
		t.Fatalf("unexpected classification: %+v", classified)
	}
}

func TestClassifyRawFrameInRawMode(t *testing.T) {
	payload := []byte{0x00, 0x20, 0xff, 0x01}
	classified, err := Classify(payload, contracts.ReceiveRawLoRa)
	if err != nil {
		t.Fatalf("classify raw frame: %v", err)
	}
	if classified.Protocol != contracts.RawLoRa || !bytes.Equal(classified.Raw, payload) {
		t.Fatalf("unexpected raw classification: %+v", classified)
	}
}

func TestClassifyAutoUsesLoRaWANWhenStructurallyValid(t *testing.T) {
	frame, err := (LoRaWANHandler{}).EncodeDataFrame(DataFrame{
		DevAddr:   0x01020304,
		FCnt:      1,
		FPort:     bytePointer(1),
		Payload:   []byte("data"),
		Direction: 0,
		NwkSKey:   make([]byte, 16),
		AppSKey:   make([]byte, 16),
	})
	if err != nil {
		t.Fatalf("encode LoRaWAN frame: %v", err)
	}

	classified, err := Classify(frame, contracts.ReceiveAuto)
	if err != nil {
		t.Fatalf("classify auto frame: %v", err)
	}
	if classified.Protocol != contracts.LoRaWAN {
		t.Fatalf("expected LoRaWAN classification, got %s", classified.Protocol)
	}
}

func TestClassifyAutoFallsBackToRaw(t *testing.T) {
	payload := []byte{0xde, 0xad, 0xbe, 0xef}
	classified, err := Classify(payload, contracts.ReceiveAuto)
	if err != nil {
		t.Fatalf("classify raw auto frame: %v", err)
	}
	if classified.Protocol != contracts.RawLoRa || !bytes.Equal(classified.Raw, payload) {
		t.Fatalf("expected raw classification, got %+v", classified)
	}
}

func TestClassifyRejectsEmptyFrameInRawMode(t *testing.T) {
	if _, err := Classify(nil, contracts.ReceiveRawLoRa); err == nil {
		t.Fatal("expected empty raw frame to be rejected")
	}
}

func TestClassifyEnforcesStrictReceiveModes(t *testing.T) {
	tests := []struct {
		name    string
		payload []byte
		mode    contracts.RadioReceiveMode
	}{
		{name: "raw frame in LoRaWAN mode", payload: []byte{0xde, 0xad}, mode: contracts.ReceiveLoRaWAN},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			if _, err := Classify(testCase.payload, testCase.mode); err == nil {
				t.Fatal("expected strict receive mode to reject incompatible frame")
			}
		})
	}
}

func bytePointer(value byte) *byte {
	return &value
}
