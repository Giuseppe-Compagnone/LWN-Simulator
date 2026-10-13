# Backend

The backend is the Go service that provides the simulator API. It manages
profiles, devices, gateways and gateway-bridge settings, starts and controls
simulations, persists run history and exposes live state to connected clients.

It serves HTTP endpoints for ordinary operations and WebSocket endpoints for
shared real-time updates. A backend instance supports one active simulation at
a time, even when several frontend clients or profiles are connected.

## Responsibilities

- validate and persist profile, device and gateway configuration;
- coordinate the simulation engine and its lifecycle;
- expose snapshots, metrics, events and historical logs;
- publish profile and simulation changes to all connected clients;
- manage virtual gateways and real gateway/ChirpStack transport settings;
- expose application status and Prometheus-compatible telemetry.

## Structure

- `cmd/server/`: executable entry point and command-line options;
- `internal/server/`: HTTP server, middleware and route registration;
- `internal/handlers/`: request decoding, validation and response mapping;
- `internal/services/`: profile, hardware, simulation and persistence
  orchestration;
- `internal/repositories/`: storage access for profiles, devices and gateways;
- `internal/database/`: database and development-storage integration;
- `internal/realtime/` and `internal/websocket/`: fan-out of live updates;
- `internal/telemetry/`: metrics instrumentation;
- `dev-db/`: development data used by local workflows.

## Run locally

```bash
cd apps/backend
go run ./cmd/server -port 8080
```

The server listens on port `8080` by default. From the repository root,
`yarn dev` starts the backend together with the web frontend.

## Tests

```bash
go test ./...
go test -race ./...
```

The backend tests cover handlers, validation, repositories, profile isolation,
simulation lifecycle, WebSocket delivery and checkpoint/log persistence.
