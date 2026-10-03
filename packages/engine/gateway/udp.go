package gateway

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	contracts "github.com/Giuseppe-Compagnone/lwn-contracts/generated"
	"github.com/Giuseppe-Compagnone/lwn-engine/types"
)

const (
	protocolVersion  byte = 2
	pushDataType     byte = 0x00
	pushAckType      byte = 0x01
	pullDataType     byte = 0x02
	pullResponseType byte = 0x03
	pullAckType      byte = 0x04
	txAckType        byte = 0x05
)

var (
	ErrInvalidGatewayEndpoint = errors.New("invalid gateway UDP endpoint")
	ErrGatewayNotConnected    = errors.New("gateway UDP adapter is not connected")
	ErrGatewayHeartbeat       = errors.New("gateway heartbeat timeout")
)

type UDPOptions struct {
	LocalAddress      string
	HeartbeatInterval time.Duration
	Timeout           time.Duration
	ReconnectInitial  time.Duration
	ReconnectMax      time.Duration
	PacketBuffer      int
	EventBuffer       int
}

func (options UDPOptions) withDefaults() UDPOptions {
	if options.LocalAddress == "" {
		options.LocalAddress = ":1700"
	}
	if options.HeartbeatInterval <= 0 {
		options.HeartbeatInterval = 5 * time.Second
	}
	if options.Timeout <= 0 {
		options.Timeout = options.HeartbeatInterval * 3
	}
	if options.ReconnectInitial <= 0 {
		options.ReconnectInitial = time.Second
	}
	if options.ReconnectMax <= 0 {
		options.ReconnectMax = 30 * time.Second
	}
	if options.PacketBuffer <= 0 {
		options.PacketBuffer = 64
	}
	if options.EventBuffer <= 0 {
		options.EventBuffer = 64
	}
	return options
}

type UDPFactory struct {
	mu      sync.Mutex
	options UDPOptions
	runtime *udpRuntime
}

type udpRuntime struct {
	conn     *net.UDPConn
	cancel   context.CancelFunc
	done     chan struct{}
	adapters map[string]*udpAdapter
}

func NewUDPFactory(options UDPOptions) *UDPFactory {
	return &UDPFactory{options: options.withDefaults()}
}

func (factory *UDPFactory) LocalAddress() string {
	factory.mu.Lock()
	defer factory.mu.Unlock()
	return factory.options.LocalAddress
}

func (factory *UDPFactory) NewGatewayAdapter(gateway contracts.Gateway) (types.GatewayAdapter, error) {
	if gateway.Type != contracts.Real {
		return nil, fmt.Errorf("gateway %s is not real", gateway.ID)
	}
	if gateway.GatewayIPv4 == nil || net.ParseIP(*gateway.GatewayIPv4).To4() == nil {
		return nil, fmt.Errorf("%w: gateway IPv4 is invalid", ErrInvalidGatewayEndpoint)
	}
	if gateway.GatewayPort == nil || *gateway.GatewayPort < 1 || *gateway.GatewayPort > 65535 {
		return nil, fmt.Errorf("%w: gateway port is invalid", ErrInvalidGatewayEndpoint)
	}
	if len(gateway.GatewayEUI) != 16 {
		return nil, fmt.Errorf("%w: gateway EUI must contain 16 hexadecimal characters", ErrInvalidGatewayEndpoint)
	}
	if _, err := hex.DecodeString(gateway.GatewayEUI); err != nil {
		return nil, fmt.Errorf("%w: gateway EUI is invalid", ErrInvalidGatewayEndpoint)
	}

	remote, err := net.ResolveUDPAddr("udp", net.JoinHostPort(*gateway.GatewayIPv4, strconv.Itoa(int(*gateway.GatewayPort))))
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidGatewayEndpoint, err)
	}
	return &udpAdapter{
		factory:    factory,
		gatewayID:  gateway.ID,
		gatewayEUI: strings.ToUpper(gateway.GatewayEUI),
		remote:     remote,
		options:    factory.options,
		inbound:    make(chan udpDatagram, factory.options.PacketBuffer),
		packets:    make(chan types.GatewayPacket, factory.options.PacketBuffer),
		events:     make(chan types.GatewayAdapterEvent, factory.options.EventBuffer),
		failures:   make(chan error, 4),
		done:       make(chan struct{}),
	}, nil
}

func (factory *UDPFactory) register(adapter *udpAdapter) error {
	factory.mu.Lock()
	defer factory.mu.Unlock()

	if factory.runtime == nil {
		local, err := net.ResolveUDPAddr("udp", factory.options.LocalAddress)
		if err != nil {
			return fmt.Errorf("resolve UDP listen address: %w", err)
		}
		conn, err := net.ListenUDP("udp", local)
		if err != nil {
			return fmt.Errorf("listen for gateway UDP packets: %w", err)
		}
		ctx, cancel := context.WithCancel(context.Background())
		runtime := &udpRuntime{
			conn:     conn,
			cancel:   cancel,
			done:     make(chan struct{}),
			adapters: make(map[string]*udpAdapter),
		}
		factory.runtime = runtime
		go factory.readLoop(ctx, runtime)
	}
	if _, exists := factory.runtime.adapters[adapter.gatewayEUI]; exists {
		return fmt.Errorf("gateway EUI %s is already registered", adapter.gatewayEUI)
	}
	factory.runtime.adapters[adapter.gatewayEUI] = adapter
	return nil
}

func (factory *UDPFactory) unregister(adapter *udpAdapter) {
	factory.mu.Lock()
	defer factory.mu.Unlock()
	if factory.runtime == nil {
		return
	}
	delete(factory.runtime.adapters, adapter.gatewayEUI)
	if len(factory.runtime.adapters) != 0 {
		return
	}
	runtime := factory.runtime
	factory.runtime = nil
	runtime.cancel()
	_ = runtime.conn.Close()
}

func (factory *UDPFactory) reconnect(adapter *udpAdapter) error {
	factory.unregister(adapter)
	return factory.register(adapter)
}

func (factory *UDPFactory) send(adapter *udpAdapter, data []byte, destination *net.UDPAddr) error {
	factory.mu.Lock()
	defer factory.mu.Unlock()
	if factory.runtime == nil {
		return ErrGatewayNotConnected
	}
	if destination == nil {
		destination = adapter.remote
	}
	if _, err := factory.runtime.conn.WriteToUDP(data, destination); err != nil {
		return fmt.Errorf("send gateway UDP packet: %w", err)
	}
	return nil
}

func (factory *UDPFactory) Close() error {
	factory.mu.Lock()
	adapters := make([]*udpAdapter, 0)
	if factory.runtime != nil {
		for _, adapter := range factory.runtime.adapters {
			adapters = append(adapters, adapter)
		}
	}
	factory.mu.Unlock()
	for _, adapter := range adapters {
		_ = adapter.Close()
	}

	factory.mu.Lock()
	runtime := factory.runtime
	factory.runtime = nil
	if runtime != nil {
		runtime.cancel()
		_ = runtime.conn.Close()
	}
	factory.mu.Unlock()
	if runtime != nil {
		<-runtime.done
	}
	return nil
}

func (factory *UDPFactory) readLoop(ctx context.Context, runtime *udpRuntime) {
	defer close(runtime.done)
	buffer := make([]byte, 64*1024)
	for {
		if ctx.Err() != nil {
			return
		}
		if err := runtime.conn.SetReadDeadline(time.Now().Add(250 * time.Millisecond)); err != nil {
			factory.broadcastError(runtime, err)
			return
		}
		size, remote, err := runtime.conn.ReadFromUDP(buffer)
		if err != nil {
			var netError net.Error
			if errors.As(err, &netError) && netError.Timeout() {
				continue
			}
			factory.broadcastError(runtime, err)
			return
		}
		factory.routeDatagram(ctx, runtime, remote, append([]byte(nil), buffer[:size]...))
	}
}

func (factory *UDPFactory) routeDatagram(ctx context.Context, runtime *udpRuntime, remote *net.UDPAddr, data []byte) {
	if len(data) < 4 || data[0] != protocolVersion {
		factory.broadcastError(runtime, errors.New("invalid Semtech UDP packet header"))
		return
	}

	factory.mu.Lock()
	if factory.runtime != runtime {
		factory.mu.Unlock()
		return
	}
	var adapter *udpAdapter
	if len(data) >= 12 && (data[3] == pushDataType || data[3] == pullDataType) {
		gatewayEUI := strings.ToUpper(hex.EncodeToString(data[4:12]))
		adapter = runtime.adapters[gatewayEUI]
	} else {
		for _, candidate := range runtime.adapters {
			if candidate.remote.IP.Equal(remote.IP) && candidate.remote.Port == remote.Port {
				adapter = candidate
				break
			}
			if adapter == nil && candidate.remote.IP.Equal(remote.IP) {
				adapter = candidate
			}
		}
	}
	factory.mu.Unlock()
	if adapter == nil {
		return
	}

	switch data[3] {
	case pushDataType:
		_ = factory.writeAck(runtime.conn, remote, data[1:3], pushAckType)
	case pullDataType:
		_ = factory.writeAck(runtime.conn, remote, data[1:3], pullAckType)
	}
	select {
	case adapter.inbound <- udpDatagram{data: data, remote: remote}:
	case <-ctx.Done():
	}
}

func (factory *UDPFactory) writeAck(conn *net.UDPConn, remote *net.UDPAddr, token []byte, packetType byte) error {
	if len(token) != 2 {
		return errors.New("invalid Semtech UDP token")
	}
	_, err := conn.WriteToUDP([]byte{protocolVersion, token[0], token[1], packetType}, remote)
	return err
}

func (factory *UDPFactory) broadcastError(runtime *udpRuntime, err error) {
	factory.mu.Lock()
	adapters := make([]*udpAdapter, 0, len(runtime.adapters))
	for _, adapter := range runtime.adapters {
		adapters = append(adapters, adapter)
	}
	factory.mu.Unlock()
	for _, adapter := range adapters {
		select {
		case adapter.failures <- err:
		default:
		}
	}
}

type udpAdapter struct {
	factory    *UDPFactory
	gatewayID  string
	gatewayEUI string
	remote     *net.UDPAddr
	options    UDPOptions

	mu       sync.Mutex
	cancel   context.CancelFunc
	started  bool
	closed   bool
	lastPeer *net.UDPAddr

	inbound  chan udpDatagram
	packets  chan types.GatewayPacket
	events   chan types.GatewayAdapterEvent
	failures chan error
	done     chan struct{}
}

type udpDatagram struct {
	data   []byte
	remote *net.UDPAddr
}

func (adapter *udpAdapter) Start(ctx context.Context) error {
	if ctx == nil {
		return errors.New("gateway adapter context cannot be nil")
	}
	adapter.mu.Lock()
	if adapter.closed {
		adapter.mu.Unlock()
		return errors.New("gateway adapter is closed")
	}
	if adapter.started {
		adapter.mu.Unlock()
		return errors.New("gateway adapter is already started")
	}
	if err := adapter.factory.register(adapter); err != nil {
		adapter.mu.Unlock()
		return err
	}
	runCtx, cancel := context.WithCancel(ctx)
	adapter.cancel = cancel
	adapter.started = true
	adapter.mu.Unlock()
	go adapter.run(runCtx)
	return nil
}

func (adapter *udpAdapter) Send(ctx context.Context, packet types.GatewayPacket) error {
	if ctx == nil {
		return errors.New("gateway packet context cannot be nil")
	}
	adapter.mu.Lock()
	destination := adapter.lastPeer
	if destination == nil {
		destination = adapter.remote
	}
	adapter.mu.Unlock()

	dataRate := packet.DataRate
	if dataRate == "" {
		dataRate = "SF7BW125"
	}
	frequency := float64(packet.Frequency) / 1_000_000
	if frequency <= 0 {
		frequency = 868.1
	}
	txpk := txPacket{
		Immediate:  true,
		Frequency:  frequency,
		Modulation: "LORA",
		DataRate:   dataRate,
		CodingRate: "4/5",
		Power:      packet.Power,
		Size:       len(packet.Payload),
		Data:       base64.StdEncoding.EncodeToString(packet.Payload),
	}
	body, err := json.Marshal(struct {
		TXPK txPacket `json:"txpk"`
	}{TXPK: txpk})
	if err != nil {
		return fmt.Errorf("encode gateway downlink: %w", err)
	}
	token, err := randomToken()
	if err != nil {
		return fmt.Errorf("create gateway packet token: %w", err)
	}
	frame := append(packetHeader(token, pullResponseType), body...)
	if err := adapter.factory.send(adapter, frame, destination); err != nil {
		return err
	}
	return nil
}

func (adapter *udpAdapter) Packets() <-chan types.GatewayPacket { return adapter.packets }

func (adapter *udpAdapter) Events() <-chan types.GatewayAdapterEvent { return adapter.events }

func (adapter *udpAdapter) Close() error {
	adapter.mu.Lock()
	if adapter.closed {
		started := adapter.started
		adapter.mu.Unlock()
		if started {
			<-adapter.done
		}
		return nil
	}
	adapter.closed = true
	cancel := adapter.cancel
	started := adapter.started
	adapter.mu.Unlock()
	if cancel != nil {
		cancel()
	}
	adapter.factory.unregister(adapter)
	if started {
		<-adapter.done
		return nil
	}
	close(adapter.done)
	close(adapter.packets)
	close(adapter.events)
	return nil
}

func (adapter *udpAdapter) run(ctx context.Context) {
	defer func() {
		adapter.factory.unregister(adapter)
		close(adapter.done)
		close(adapter.packets)
		close(adapter.events)
	}()

	lastActivity := time.Now()
	connected := false
	backoff := adapter.options.ReconnectInitial
	reconnectAt := time.Time{}
	tickInterval := 250 * time.Millisecond
	if heartbeatTick := adapter.options.HeartbeatInterval / 2; heartbeatTick > 0 && heartbeatTick < tickInterval {
		tickInterval = heartbeatTick
	}
	ticker := time.NewTicker(tickInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case datagram, ok := <-adapter.inbound:
			if !ok {
				return
			}
			adapter.mu.Lock()
			adapter.lastPeer = datagram.remote
			adapter.mu.Unlock()
			lastActivity = time.Now()
			if !connected {
				if !adapter.emitState(ctx, contracts.Connected) {
					return
				}
				connected = true
				backoff = adapter.options.ReconnectInitial
				reconnectAt = time.Time{}
			}
			if err := adapter.handleDatagram(ctx, datagram); err != nil && !adapter.emitError(ctx, err, false) {
				return
			}
		case err := <-adapter.failures:
			if err == nil {
				continue
			}
			if !adapter.emitError(ctx, err, false) || !adapter.emitState(ctx, contracts.Disconnected) || !adapter.emitState(ctx, contracts.Reconnecting) {
				return
			}
			connected = false
			reconnectAt = time.Now().Add(backoff)
			backoff = minDuration(backoff*2, adapter.options.ReconnectMax)
		case <-ticker.C:
			now := time.Now()
			if now.Sub(lastActivity) >= adapter.options.Timeout {
				if !adapter.emitError(ctx, ErrGatewayHeartbeat, true) || !adapter.emitState(ctx, contracts.Disconnected) || !adapter.emitState(ctx, contracts.Reconnecting) {
					return
				}
				connected = false
				lastActivity = now
				reconnectAt = now.Add(backoff)
				backoff = minDuration(backoff*2, adapter.options.ReconnectMax)
			} else if !connected && !reconnectAt.IsZero() && !now.Before(reconnectAt) {
				if ctx.Err() != nil {
					return
				}
				if err := adapter.factory.reconnect(adapter); err != nil && !adapter.emitError(ctx, err, false) {
					return
				}
				reconnectAt = now.Add(backoff)
				backoff = minDuration(backoff*2, adapter.options.ReconnectMax)
			}
		}
	}
}

func (adapter *udpAdapter) handleDatagram(ctx context.Context, datagram udpDatagram) error {
	data := datagram.data
	if len(data) < 4 || data[0] != protocolVersion {
		return errors.New("invalid Semtech UDP packet header")
	}
	switch data[3] {
	case pullDataType:
		return adapter.emitHeartbeat(ctx)
	case pushDataType:
		if len(data) < 12 {
			return errors.New("invalid PUSH_DATA packet")
		}
		if err := adapter.validateGatewayEUI(data[4:12]); err != nil {
			return err
		}
		return adapter.handlePushData(ctx, data[12:])
	case txAckType:
		return adapter.handleTXAck(data[4:])
	default:
		return fmt.Errorf("unsupported Semtech UDP packet type 0x%02x", data[3])
	}
}

func (adapter *udpAdapter) handlePushData(ctx context.Context, body []byte) error {
	var payload pushDataPayload
	if len(body) == 0 {
		return adapter.emitHeartbeat(ctx)
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return fmt.Errorf("decode PUSH_DATA payload: %w", err)
	}
	if len(payload.RXPK) == 0 {
		return adapter.emitHeartbeat(ctx)
	}
	for _, received := range payload.RXPK {
		decoded, err := base64.StdEncoding.DecodeString(received.Data)
		if err != nil {
			return fmt.Errorf("decode gateway payload: %w", err)
		}
		packet := types.GatewayPacket{
			GatewayID:       adapter.gatewayID,
			Payload:         decoded,
			Frequency:       int64(math.Round(received.Frequency * 1_000_000)),
			Bandwidth:       parseBandwidth(received.DataRate),
			SpreadingFactor: parseSpreadingFactor(received.DataRate),
			DataRate:        received.DataRate,
			ReceivedAt:      time.Now(),
		}
		select {
		case adapter.packets <- packet:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return adapter.emitHeartbeat(ctx)
}

func parseSpreadingFactor(dataRate string) int {
	if !strings.HasPrefix(dataRate, "SF") {
		return 0
	}
	value, err := strconv.Atoi(strings.TrimPrefix(strings.Split(dataRate, "BW")[0], "SF"))
	if err != nil || value < 7 || value > 12 {
		return 0
	}
	return value
}

func parseBandwidth(dataRate string) int64 {
	parts := strings.Split(dataRate, "BW")
	if len(parts) != 2 {
		return 0
	}
	value, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil || value <= 0 {
		return 0
	}
	return value * 1000
}

func (adapter *udpAdapter) handleTXAck(body []byte) error {
	if len(body) == 0 {
		return nil
	}
	var ack txAckPayload
	if err := json.Unmarshal(body, &ack); err != nil {
		return fmt.Errorf("decode TX_ACK payload: %w", err)
	}
	if ack.TXPKAck.Error != "" && ack.TXPKAck.Error != "NONE" {
		return fmt.Errorf("gateway TX_ACK: %s", ack.TXPKAck.Error)
	}
	return nil
}

func (adapter *udpAdapter) validateGatewayEUI(value []byte) error {
	expected, err := hex.DecodeString(adapter.gatewayEUI)
	if err != nil || len(expected) != len(value) || !equalBytes(expected, value) {
		return errors.New("gateway EUI in PUSH_DATA does not match configured gateway")
	}
	return nil
}

func (adapter *udpAdapter) emitState(ctx context.Context, state contracts.GatewayConnectionState) bool {
	return adapter.emitEvent(ctx, types.GatewayAdapterEvent{GatewayID: adapter.gatewayID, State: state, At: time.Now()})
}

func (adapter *udpAdapter) emitHeartbeat(ctx context.Context) error {
	if !adapter.emitEvent(ctx, types.GatewayAdapterEvent{GatewayID: adapter.gatewayID, State: contracts.Connected, Heartbeat: true, At: time.Now()}) {
		return ctx.Err()
	}
	return nil
}

func (adapter *udpAdapter) emitError(ctx context.Context, err error, timeout bool) bool {
	return adapter.emitEvent(ctx, types.GatewayAdapterEvent{GatewayID: adapter.gatewayID, State: contracts.Error, Error: err.Error(), Timeout: timeout, At: time.Now()})
}

func (adapter *udpAdapter) emitEvent(ctx context.Context, event types.GatewayAdapterEvent) bool {
	select {
	case adapter.events <- event:
		return true
	case <-ctx.Done():
		return false
	}
}

type pushDataPayload struct {
	RXPK []rxPacket `json:"rxpk"`
}

type rxPacket struct {
	Data      string  `json:"data"`
	Frequency float64 `json:"freq"`
	DataRate  string  `json:"datr"`
}

type txPacket struct {
	Immediate  bool    `json:"imme"`
	Frequency  float64 `json:"freq"`
	Modulation string  `json:"modu"`
	DataRate   string  `json:"datr"`
	CodingRate string  `json:"codr"`
	Power      int     `json:"powe,omitempty"`
	Size       int     `json:"size"`
	Data       string  `json:"data"`
}

type txAckPayload struct {
	TXPKAck struct {
		Error string `json:"error"`
	} `json:"txpk_ack"`
}

func packetHeader(token [2]byte, packetType byte) []byte {
	return []byte{protocolVersion, token[0], token[1], packetType}
}

func randomToken() ([2]byte, error) {
	var token [2]byte
	_, err := rand.Read(token[:])
	return token, err
}

func minDuration(left, right time.Duration) time.Duration {
	if left < right {
		return left
	}
	return right
}

func equalBytes(left, right []byte) bool {
	return len(left) == len(right) && strings.EqualFold(hex.EncodeToString(left), hex.EncodeToString(right))
}
