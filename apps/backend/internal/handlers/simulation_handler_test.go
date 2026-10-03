package handlers

import (
	"context"
	"net/http/httptest"
	"testing"
	"time"

	contracts "github.com/Giuseppe-Compagnone/lwn-contracts/generated"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/gorilla/websocket"
	"lwn-simulator-backend/internal/services"
)

type fakeSimulationService struct {
	snapshot contracts.SimulationSnapshot
	updates  chan services.SimulationUpdate
}

func (service *fakeSimulationService) Start(context.Context, contracts.SimulationConfig) (contracts.SimulationSnapshot, error) {
	return service.snapshot, nil
}

func (service *fakeSimulationService) Pause() (contracts.SimulationSnapshot, error) {
	return service.snapshot, nil
}

func (service *fakeSimulationService) Resume() (contracts.SimulationSnapshot, error) {
	return service.snapshot, nil
}

func (service *fakeSimulationService) Stop(context.Context) (contracts.SimulationSnapshot, error) {
	return service.snapshot, nil
}

func (service *fakeSimulationService) Snapshot() (contracts.SimulationSnapshot, error) {
	return service.snapshot, nil
}

func (service *fakeSimulationService) Subscribe() (services.SimulationSubscription, error) {
	return services.SimulationSubscription{Updates: service.updates, Close: func() {}}, nil
}

func TestSimulationHandlerWebSocketStreamsSnapshotsAndEvents(t *testing.T) {
	gin.SetMode(gin.TestMode)
	service := &fakeSimulationService{
		snapshot: contracts.SimulationSnapshot{
			State: contracts.SimulationState{Status: contracts.SimulationStatusRunning, Speed: 1},
		},
		updates: make(chan services.SimulationUpdate, 1),
	}
	handler := NewSimulationHandler(service, validator.New())
	router := gin.New()
	router.GET("/ws", handler.WebSocket)
	server := httptest.NewServer(router)
	defer server.Close()

	connection, _, err := websocket.DefaultDialer.Dial("ws"+server.URL[len("http"):]+"/ws", nil)
	if err != nil {
		t.Fatalf("dial simulation websocket: %v", err)
	}
	defer connection.Close()

	var initial contracts.SimulationWebSocketMessage
	if err := connection.ReadJSON(&initial); err != nil {
		t.Fatalf("read initial snapshot: %v", err)
	}
	if initial.Type != contracts.SimulationSnapshotMessage || initial.Snapshot == nil {
		t.Fatalf("unexpected initial websocket message: %+v", initial)
	}

	service.updates <- services.SimulationUpdate{
		Snapshot: service.snapshot,
		Event: contracts.SimulationEvent{
			ID: "event-1", Sequence: 1, Type: contracts.SimulationPaused, Message: "paused",
		},
	}
	_ = connection.SetReadDeadline(time.Now().Add(time.Second))
	var event contracts.SimulationWebSocketMessage
	if err := connection.ReadJSON(&event); err != nil {
		t.Fatalf("read event update: %v", err)
	}
	if event.Type != contracts.SimulationEventMessage || event.Event == nil || event.Event.Type != contracts.SimulationPaused {
		t.Fatalf("unexpected event websocket message: %+v", event)
	}
}
