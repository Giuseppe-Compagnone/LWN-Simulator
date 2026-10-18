package payloadcodec

import (
	"strings"
	"testing"
	"time"
)

func TestDecodeReturnsDataObject(t *testing.T) {
	decoded, err := Decode(`function decodeUplink(input) {
  return { data: { temperature: input.bytes[0], port: input.fPort } };
}`, Input{Bytes: []byte{21}, FPort: 10})
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if decoded["temperature"] != float64(21) || decoded["port"] != float64(10) {
		t.Fatalf("unexpected decoded data: %#v", decoded)
	}
}

func TestDecodeAcceptsDirectObjectResult(t *testing.T) {
	decoded, err := Decode("function decodeUplink() { return { value: 'ok' }; }", Input{})
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if decoded["value"] != "ok" {
		t.Fatalf("unexpected decoded data: %#v", decoded)
	}
}

func TestDecodeReportsRuntimeErrors(t *testing.T) {
	_, err := Decode("function decodeUplink() { throw new Error('bad payload'); }", Input{})
	if err == nil || !strings.Contains(err.Error(), "decodeUplink failed") {
		t.Fatalf("expected decoder error, got %v", err)
	}
}

func TestDecodeInterruptsLongRunningCode(t *testing.T) {
	_, err := DecodeWithTimeout("function decodeUplink() { while (true) {} }", Input{}, 5*time.Millisecond)
	if err == nil || !strings.Contains(err.Error(), "timed out") {
		t.Fatalf("expected timeout error, got %v", err)
	}
}

func TestDecodeRejectsNonObjectData(t *testing.T) {
	_, err := Decode("function decodeUplink() { return { data: 42 }; }", Input{})
	if err == nil || !strings.Contains(err.Error(), "data must be a JSON object") {
		t.Fatalf("expected object validation error, got %v", err)
	}
}
