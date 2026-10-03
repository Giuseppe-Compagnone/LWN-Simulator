package runtime

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	contracts "github.com/Giuseppe-Compagnone/lwn-contracts/generated"
	"github.com/Giuseppe-Compagnone/lwn-engine/types"
	"github.com/google/uuid"
)

func (e *Engine) startGatewayAdapters(ctx context.Context) {
	if e.gatewayAdapterFactory == nil {
		return
	}

	for _, gateway := range e.registry.ActiveGateways() {
		if gateway.Type != contracts.Real {
			continue
		}

		adapter, err := e.gatewayAdapterFactory.NewGatewayAdapter(gateway)
		if err != nil {
			e.publishGatewayError(gateway.ID, fmt.Errorf("create gateway adapter: %w", err), false)
			continue
		}

		e.mu.Lock()
		e.gatewayAdapters[gateway.ID] = adapter
		connecting := e.updateGatewayRuntimeLocked(gateway.ID, contracts.Connecting, "", true)
		e.mu.Unlock()
		e.publish(connecting)

		go e.consumeGatewayAdapter(ctx, gateway.ID, adapter)
		if err := adapter.Start(ctx); err != nil {
			e.publishGatewayError(gateway.ID, fmt.Errorf("start gateway adapter: %w", err), false)
			_ = adapter.Close()
		}
	}
}

func (e *Engine) consumeGatewayAdapter(ctx context.Context, gatewayID string, adapter types.GatewayAdapter) {
	packets := adapter.Packets()
	adapterEvents := adapter.Events()

	for packets != nil || adapterEvents != nil {
		select {
		case <-ctx.Done():
			return
		case event, ok := <-adapterEvents:
			if !ok {
				adapterEvents = nil
				continue
			}
			if event.GatewayID == "" {
				event.GatewayID = gatewayID
			}
			e.handleGatewayAdapterEvent(event)
		case packet, ok := <-packets:
			if !ok {
				packets = nil
				continue
			}
			if packet.GatewayID == "" {
				packet.GatewayID = gatewayID
			}
			e.handleGatewayPacket(packet)
		}
	}
}

func (e *Engine) handleGatewayAdapterEvent(adapterEvent types.GatewayAdapterEvent) {
	e.mu.Lock()

	runtime, ok := e.gatewayRuntime[adapterEvent.GatewayID]
	if !ok {
		e.mu.Unlock()
		return
	}

	if adapterEvent.Heartbeat {
		runtime.State = contracts.Connected
		runtime.LastHeartbeat = e.clock.Now()
		e.metrics.GatewayHeartbeats++
		event := e.newGatewayEventLocked(
			contracts.GatewayHeartbeat,
			"gateway heartbeat received",
			adapterEvent.GatewayID,
			contracts.Connected,
			nil,
			"",
			"",
		)
		e.mu.Unlock()
		e.publish(event)
		return
	}

	eventType := contracts.GatewayNetworkError
	message := "gateway network error"
	switch adapterEvent.State {
	case contracts.Connected:
		runtime.State = contracts.Connected
		runtime.LastNetworkError = ""
		e.metrics.GatewayConnections++
		eventType = contracts.GatewayConnected
		message = "gateway connected"
	case contracts.Disconnected:
		runtime.State = contracts.Disconnected
		eventType = contracts.GatewayDisconnected
		message = "gateway disconnected"
	case contracts.Reconnecting:
		runtime.State = contracts.Reconnecting
		runtime.ConnectionAttempts++
		e.metrics.GatewayReconnects++
		eventType = contracts.GatewayReconnecting
		message = "gateway reconnecting"
	case contracts.Error:
		runtime.State = contracts.Error
		runtime.LastNetworkError = adapterEvent.Error
		e.metrics.GatewayNetworkErrors++
		if adapterEvent.Timeout {
			e.metrics.GatewayTimeouts++
		}
	default:
		e.mu.Unlock()
		return
	}

	event := e.newGatewayEventLocked(
		eventType,
		message,
		adapterEvent.GatewayID,
		adapterEvent.State,
		nil,
		"",
		adapterEvent.Error,
	)
	e.mu.Unlock()
	e.publish(event)
}

func (e *Engine) handleGatewayPacket(packet types.GatewayPacket) {
	e.mu.Lock()
	runtime, ok := e.gatewayRuntime[packet.GatewayID]
	if !ok {
		e.mu.Unlock()
		return
	}
	runtime.IngressPackets++
	e.metrics.GatewayIngressPackets++
	payloadSize := int64(len(packet.Payload))
	event := e.newGatewayEventLocked(
		contracts.GatewayPacketIngress,
		"packet received from real gateway",
		packet.GatewayID,
		contracts.Connected,
		&payloadSize,
		"",
		"",
	)
	if packet.Frequency > 0 {
		event.ChannelFrequency = &packet.Frequency
	}
	if packet.Bandwidth > 0 {
		event.Bandwidth = &packet.Bandwidth
	}
	if packet.SpreadingFactor > 0 {
		spreadingFactor := int32(packet.SpreadingFactor)
		event.SpreadingFactor = &spreadingFactor
	}
	if dataRate := parseGatewayDataRate(packet.DataRate); dataRate >= 0 {
		event.DataRate = &dataRate
	}
	e.mu.Unlock()
	e.publish(event)
}

func (e *Engine) processVirtualGatewayHeartbeatLocked(scheduled types.ScheduledEvent) []contracts.SimulationEvent {
	gateway, ok := e.registry.Gateway(scheduled.GatewayID)
	if !ok || !gateway.Active || gateway.Type != contracts.Virtual || gateway.KeepAlive == nil {
		return nil
	}
	if runtime := e.gatewayRuntime[gateway.ID]; runtime != nil {
		runtime.State = contracts.Connected
		runtime.LastHeartbeat = scheduled.At
	}
	e.metrics.GatewayHeartbeats++
	event := e.newGatewayEventLocked(
		contracts.GatewayHeartbeat,
		"virtual gateway heartbeat received",
		gateway.ID,
		contracts.Connected,
		nil,
		"",
		"",
	)
	next := types.ScheduledEvent{
		ID: uuid.NewString(), At: scheduled.At + time.Duration(*gateway.KeepAlive)*time.Second,
		Type: contracts.GatewayHeartbeat, Message: "virtual gateway heartbeat scheduled",
		GatewayID: gateway.ID, Kind: types.ScheduledEventGatewayHeartbeat,
	}
	if err := e.scheduler.Schedule(next); err != nil {
		return []contracts.SimulationEvent{event, e.newEventLocked(contracts.PacketDropped, "virtual gateway heartbeat could not be scheduled", "", gateway.ID, next.ID)}
	}
	return []contracts.SimulationEvent{event}
}

func parseGatewayDataRate(value string) int {
	if !strings.HasPrefix(value, "SF") {
		return -1
	}
	parts := strings.SplitN(strings.TrimPrefix(value, "SF"), "BW", 2)
	if len(parts) != 2 {
		return -1
	}
	sf, err := strconv.Atoi(parts[0])
	if err != nil || sf < 7 || sf > 12 {
		return -1
	}
	return 12 - sf
}

func (e *Engine) publishGatewayError(gatewayID string, err error, timeout bool) {
	e.handleGatewayAdapterEvent(types.GatewayAdapterEvent{
		GatewayID: gatewayID,
		State:     contracts.Error,
		Error:     err.Error(),
		Timeout:   timeout,
	})
}

func (e *Engine) updateGatewayRuntimeLocked(
	gatewayID string,
	state contracts.GatewayConnectionState,
	errorMessage string,
	countAttempt bool,
) contracts.SimulationEvent {
	runtime := e.gatewayRuntime[gatewayID]
	if runtime == nil {
		runtime = &types.GatewayRuntime{}
		e.gatewayRuntime[gatewayID] = runtime
	}
	runtime.State = state
	runtime.LastNetworkError = errorMessage
	if countAttempt {
		runtime.ConnectionAttempts++
	}
	return e.newGatewayEventLocked(
		contracts.GatewayConnecting,
		"gateway connection starting",
		gatewayID,
		state,
		nil,
		"",
		errorMessage,
	)
}

func (e *Engine) newGatewayEventLocked(
	eventType contracts.SimulationEventType,
	message string,
	gatewayID string,
	state contracts.GatewayConnectionState,
	payloadSize *int64,
	packetID string,
	errorMessage string,
) contracts.SimulationEvent {
	event := e.newEventLocked(eventType, message, "", gatewayID, "")
	if packetID != "" {
		event.PacketID = &packetID
	}
	event.GatewayState = &state
	if payloadSize != nil {
		event.PayloadSize = payloadSize
	}
	if errorMessage != "" {
		event.Error = &errorMessage
	}
	e.eventLog[len(e.eventLog)-1] = event
	return event
}

func (e *Engine) stopGatewayAdapters() {
	e.mu.Lock()
	adapters := make([]types.GatewayAdapter, 0, len(e.gatewayAdapters))
	for gatewayID, adapter := range e.gatewayAdapters {
		adapters = append(adapters, adapter)
		delete(e.gatewayAdapters, gatewayID)
	}
	e.mu.Unlock()

	for _, adapter := range adapters {
		_ = adapter.Close()
	}
	if closer, ok := e.gatewayAdapterFactory.(types.GatewayAdapterFactoryCloser); ok {
		_ = closer.Close()
	}
}

func (e *Engine) SendGatewayPacket(ctx context.Context, gatewayID string, payload []byte) error {
	if ctx == nil {
		return fmt.Errorf("send gateway packet context cannot be nil")
	}

	e.mu.RLock()
	adapter, ok := e.gatewayAdapters[gatewayID]
	e.mu.RUnlock()
	if !ok || adapter == nil {
		return fmt.Errorf("%w: %s", ErrGatewayTransportUnavailable, gatewayID)
	}

	packet := types.GatewayPacket{GatewayID: gatewayID, Payload: append([]byte(nil), payload...)}
	if err := adapter.Send(ctx, packet); err != nil {
		e.publishGatewayError(gatewayID, err, false)
		return err
	}

	e.mu.Lock()
	runtime := e.gatewayRuntime[gatewayID]
	if runtime != nil {
		runtime.EgressPackets++
	}
	e.metrics.GatewayEgressPackets++
	payloadSize := int64(len(payload))
	event := e.newGatewayEventLocked(
		contracts.GatewayPacketEgress,
		"packet sent to real gateway",
		gatewayID,
		contracts.Connected,
		&payloadSize,
		"",
		"",
	)
	e.mu.Unlock()
	e.publish(event)
	return nil
}
