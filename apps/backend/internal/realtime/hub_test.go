package realtime

import (
	"testing"
	"time"

	contracts "github.com/Giuseppe-Compagnone/lwn-contracts/generated"
)

func TestHubPublishesToEverySubscriber(t *testing.T) {
	hub := NewHub()
	first := hub.Subscribe()
	second := hub.Subscribe()
	defer first.Close()
	defer second.Close()

	want := contracts.RealtimeWebSocketMessage{
		Type:                  contracts.RealtimeDeviceCreatedMessage,
		TimestampMilliseconds: 1,
		ResourceID:            stringPointer("device-1"),
	}
	hub.Publish(want)

	for name, subscription := range map[string]Subscription{"first": first, "second": second} {
		t.Run(name, func(t *testing.T) {
			select {
			case got := <-subscription.Updates:
				if got.Type != want.Type || got.ResourceID == nil || *got.ResourceID != "device-1" {
					t.Fatalf("unexpected realtime message: %+v", got)
				}
			case <-time.After(time.Second):
				t.Fatal("subscriber did not receive realtime message")
			}
		})
	}
}

func TestHubCloseIsIdempotentAndClosesStream(t *testing.T) {
	hub := NewHub()
	subscription := hub.Subscribe()
	subscription.Close()
	subscription.Close()

	select {
	case _, ok := <-subscription.Updates:
		if ok {
			t.Fatal("closed subscription still returned a message")
		}
	case <-time.After(time.Second):
		t.Fatal("subscription was not closed")
	}
}

func stringPointer(value string) *string { return &value }
