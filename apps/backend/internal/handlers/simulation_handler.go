package handlers

import (
	"context"
	"net/http"
	"net/url"
	"strings"
	"time"

	contracts "github.com/Giuseppe-Compagnone/lwn-contracts/generated"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/gorilla/websocket"
	"lwn-simulator-backend/internal/services"
)

type SimulationService interface {
	Start(context.Context, contracts.SimulationConfig) (contracts.SimulationSnapshot, error)
	Pause() (contracts.SimulationSnapshot, error)
	Resume() (contracts.SimulationSnapshot, error)
	Stop(context.Context) (contracts.SimulationSnapshot, error)
	Snapshot() (contracts.SimulationSnapshot, error)
	Subscribe() (services.SimulationSubscription, error)
}

type SimulationHandler struct {
	service   SimulationService
	validator *validator.Validate
	upgrader  websocket.Upgrader
}

func NewSimulationHandler(service SimulationService, validator *validator.Validate) *SimulationHandler {
	return &SimulationHandler{
		service:   service,
		validator: validator,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  4096,
			WriteBufferSize: 8192,
			CheckOrigin: func(r *http.Request) bool {
				origin := r.Header.Get("Origin")
				if origin == "" {
					return true
				}
				parsed, parseErr := url.Parse(origin)
				if parseErr != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") {
					return false
				}
				return strings.EqualFold(parsed.Hostname(), "localhost") ||
					parsed.Hostname() == "127.0.0.1" || parsed.Hostname() == "::1"
			},
		},
	}
}

func (h *SimulationHandler) Start(c *gin.Context) {
	var req contracts.SimulationConfig
	if !bindJSONAndValidate(c, h.validator, &req) {
		return
	}
	res, err := h.service.Start(c.Request.Context(), req)
	if err != nil {
		writeServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, res)
}

func (h *SimulationHandler) Pause(c *gin.Context) { h.control(c, h.service.Pause) }

func (h *SimulationHandler) Resume(c *gin.Context) { h.control(c, h.service.Resume) }

func (h *SimulationHandler) Stop(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	res, err := h.service.Stop(ctx)
	if err != nil {
		writeServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, res)
}

func (h *SimulationHandler) Snapshot(c *gin.Context) {
	res, err := h.service.Snapshot()
	if err != nil {
		writeServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, res)
}

func (h *SimulationHandler) WebSocket(c *gin.Context) {
	connection, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer connection.Close()

	subscription, err := h.service.Subscribe()
	if err != nil {
		_ = writeWebSocketMessage(connection, contracts.SimulationWebSocketMessage{
			Type:                  contracts.SimulationErrorMessage,
			TimestampMilliseconds: time.Now().UnixMilli(),
			Error:                 stringPointer(err.Error()),
		})
		return
	}
	defer subscription.Close()

	if snapshot, snapshotErr := h.service.Snapshot(); snapshotErr == nil {
		if err := writeWebSocketMessage(connection, contracts.SimulationWebSocketMessage{
			Type:                  contracts.SimulationSnapshotMessage,
			TimestampMilliseconds: time.Now().UnixMilli(),
			Snapshot:              &snapshot,
		}); err != nil {
			return
		}
	}

	done := watchWebSocketConnection(connection)
	for {
		select {
		case <-done:
			return
		case update, ok := <-subscription.Updates:
			if !ok {
				return
			}
			event := update.Event
			if err := writeWebSocketMessage(connection, contracts.SimulationWebSocketMessage{
				Type:                  contracts.SimulationEventMessage,
				TimestampMilliseconds: time.Now().UnixMilli(),
				Event:                 &event,
				Snapshot:              &update.Snapshot,
			}); err != nil {
				return
			}
		}
	}
}

func (h *SimulationHandler) control(
	c *gin.Context,
	action func() (contracts.SimulationSnapshot, error),
) {
	res, err := action()
	if err != nil {
		writeServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, res)
}

func writeWebSocketMessage(connection *websocket.Conn, message contracts.SimulationWebSocketMessage) error {
	return connection.WriteJSON(message)
}

func watchWebSocketConnection(connection *websocket.Conn) <-chan struct{} {
	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			if _, _, err := connection.ReadMessage(); err != nil {
				return
			}
		}
	}()
	return done
}

func stringPointer(value string) *string { return &value }
