# Simulation engine

The engine is the deterministic, server-side core of the simulator. It owns
simulation time, lifecycle transitions, the runtime registry, the scheduler,
and the event log.

Shared data types are defined in the OpenAPI YAML contracts under
`packages/contracts/schemas/Simulation` and generated for Go and TypeScript.
Implementation-only types live under `types/` when they are needed by the
engine runtime and test clocks.

## Package layout

- `engine.go`: public lifecycle and traffic orchestration;
- `protocol.go`: OTAA/ABP sessions, receive windows, ACKs and retries;
- `types/`: runtime abstractions such as clocks, scheduler events and sinks;
- `scheduler/`: thread-safe priority queue for simulation events;
- `registry/`: validated device and gateway runtime registry;
- `validation/`: hardware and simulation configuration validation;
- `geometry/`: geographic distance and coverage calculations.

The small wrapper files in the root package preserve the public `engine` API
while keeping each implementation concern isolated in its own package.

## Current scope

The foundation currently provides:

- lifecycle control: start, pause, resume, and stop;
- real and deterministic manual clocks;
- a priority-based event scheduler;
- validated device and gateway runtime registration;
- immutable runtime snapshots;
- ordered event records and a live event channel;
- periodic uplinks generated from active devices;
- deterministic geometric coverage for active virtual gateways;
- packet delivery/drop events and base traffic metrics;
- OTAA join sessions and immediate ABP sessions;
- RX1/RX2 receive windows for confirmed traffic;
- ACK handling, frame counters and bounded retransmissions;
- deterministic tests, including race detection.

At the beginning of a run, every active device gets an uplink scheduled at
simulation time zero. Subsequent uplinks follow the device's
`payloadConfig.uplinkInterval`, expressed in seconds. A packet is received by
each active virtual gateway whose location is within the device's
`advancedConfig.antennaRange`, expressed in meters. Real gateways are not
connected by this phase and are intentionally excluded from virtual coverage.

The event log is authoritative and retains every generated event. The live
channel is a non-blocking view for consumers, so a slow consumer cannot stop
the simulation. Each packet event carries a `packetID`; snapshots expose
`SimulationMetrics` with uplink totals, delivery rate, join/ACK counters,
retransmissions, frame-counter errors and active runtime counts.

ABP devices are ready to transmit when the simulation starts. OTAA devices
first send a join request and become ready after a join accept in RX1 or RX2.
Confirmed uplinks keep the same frame counter and packet identity across
retransmissions, and stop after the device's configured retransmission limit.
RX1/RX2 delays use seconds, while RX2 duration uses milliseconds, matching the
device configuration model.

LoRaWAN protocol behavior, radio propagation, real gateway transports, and
frontend streaming build on top of this foundation in the following roadmap
phases.
