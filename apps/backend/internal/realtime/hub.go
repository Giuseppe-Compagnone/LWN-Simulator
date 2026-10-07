package realtime

import (
	"sync"

	contracts "github.com/Giuseppe-Compagnone/lwn-contracts/generated"
)

const subscriberBufferSize = 4096

// Subscription is a best-effort realtime stream. Consumers can recover a
// missed update through the corresponding HTTP endpoint after reconnecting.
type Subscription struct {
	Updates <-chan contracts.RealtimeWebSocketMessage
	Close   func()
}

// Hub fans application updates out to every connected frontend. A single hub
// is shared by hardware, simulation and log producers so multiple browser
// tabs observe the same state transitions.
type Hub struct {
	mu          sync.Mutex
	nextID      uint64
	subscribers map[uint64]chan contracts.RealtimeWebSocketMessage
}

func NewHub() *Hub {
	return &Hub{subscribers: make(map[uint64]chan contracts.RealtimeWebSocketMessage)}
}

func (h *Hub) Subscribe() Subscription {
	h.mu.Lock()
	h.nextID++
	id := h.nextID
	updates := make(chan contracts.RealtimeWebSocketMessage, subscriberBufferSize)
	h.subscribers[id] = updates
	h.mu.Unlock()

	return Subscription{
		Updates: updates,
		Close: func() {
			h.mu.Lock()
			if channel, ok := h.subscribers[id]; ok {
				delete(h.subscribers, id)
				close(channel)
			}
			h.mu.Unlock()
		},
	}
}

func (h *Hub) Publish(message contracts.RealtimeWebSocketMessage) {
	h.mu.Lock()
	defer h.mu.Unlock()

	for _, updates := range h.subscribers {
		select {
		case updates <- message:
		default:
			// The websocket transport will reconnect and consumers can recover
			// durable state through HTTP. Never block a simulation or CRUD call
			// on a slow browser tab.
		}
	}
}
