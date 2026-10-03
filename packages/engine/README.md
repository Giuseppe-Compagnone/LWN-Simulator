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
- `radio.go`: airtime, channel selection, link budget and radio outcomes;
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
- radio transmissions with deterministic airtime and regional channel selection;
- log-distance RSSI/SNR estimates, overlapping-transmission collision detection
  and seeded packet-loss outcomes;
- radio-level metrics and link metadata in the event stream and runtime snapshot;
- deterministic tests, including race detection.

At the beginning of a run, every active device gets an uplink scheduled at
simulation time zero. Subsequent uplinks follow the device's
`payloadConfig.uplinkInterval`, expressed in seconds. The packet first occupies
the selected regional channel for its calculated airtime. The current channel
plan is intentionally small and deterministic: it provides the primary
simulation channels for each supported region without attempting to model
every regulatory channel.

At transmission completion, a packet is delivered to each active virtual
gateway whose location is within the device's `advancedConfig.antennaRange`,
expressed in meters. Real gateways are not connected by this phase and are
intentionally excluded from virtual coverage. Transmissions overlap when they
share a gateway, frequency, bandwidth and spreading factor; overlapping
transmissions are marked as collisions and are lost. A packet with no covered
gateway is lost with reason `no-gateway`. Weak links can produce `low-snr` or
seeded `random-loss` outcomes.

The event log is authoritative and retains every generated event. The live
channel is a non-blocking view for consumers, so a slow consumer cannot stop
the simulation. Each packet event carries a `packetID`; radio events also
expose channel frequency, bandwidth, spreading factor, airtime, RSSI, SNR,
collision status and, for losses, a typed loss reason. Snapshots expose
`SimulationMetrics` with uplink totals, delivery rate, join/ACK counters,
retransmissions, frame-counter errors, radio transmissions/losses/collisions,
average link measurements and active runtime counts. Runtime devices retain
their last radio measurements for the frontend.

ABP devices are ready to transmit when the simulation starts. OTAA devices
first send a join request and become ready after a join accept in RX1 or RX2.
Confirmed uplinks keep the same frame counter and packet identity across
retransmissions, and stop after the device's configured retransmission limit.
RX1/RX2 delays use seconds, while RX2 duration uses milliseconds, matching the
device configuration model.

Real gateway transports and frontend streaming build on top of this foundation
in the following roadmap phases.
