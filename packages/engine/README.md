# Simulation engine

The engine is the deterministic, server-side core of the simulator. It owns
simulation time, lifecycle transitions, the runtime registry, the scheduler,
and the event log.

Shared data types are defined in the OpenAPI YAML contracts under
`packages/contracts/schemas/Simulation` and generated for Go and TypeScript.
Implementation-only types live under `types/` when they are needed by the
engine runtime and test clocks.

## Package layout

- `engine.go`, `errors.go`, `registry.go`, `scheduler.go`, `validation.go`:
  small public facades that preserve the stable `engine` import path;
- `internal/runtime/`: lifecycle and traffic orchestration implementation,
  including protocol sessions, receive windows, ACKs, retries, airtime,
  channel selection, link budget and radio outcomes;
- `gateway/`: real-gateway adapters, currently Semtech UDP;
- `types/`: runtime abstractions such as clocks, scheduler events and sinks;
- `scheduler/`: thread-safe priority queue for simulation events;
- `registry/`: validated device and gateway runtime registry;
- `validation/`: hardware and simulation configuration validation;
- `geometry/`: geographic distance and coverage calculations.

Runtime tests live next to the implementation in `internal/runtime/`, while
package-specific tests remain next to their domain package. This keeps the
module root focused on its public API and makes the dependency direction
explicit: the facade delegates to the internal runtime, and the runtime uses
the reusable domain packages.

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
- LoRaWAN data-frame construction/parsing with AES-128 payload protection and MIC validation at the gateway boundary;
- RX1/RX2 receive windows for confirmed traffic;
- ACK handling, frame counters and bounded retransmissions;
- Class A and Class C application downlink queues, in addition to Class B ping-slot downlinks;
- radio transmissions with deterministic airtime and regional channel selection;
- log-distance RSSI/SNR estimates, overlapping-transmission collision detection
  and seeded packet-loss outcomes;
- radio-level metrics and link metadata in the event stream and runtime snapshot;
- real-gateway transport hooks with a shared Semtech UDP listener, heartbeat
  monitoring, packet ingress/egress, timeout detection and reconnect backoff;
- real-gateway uplink routing for ABP/OTAA data frames, confirmed-uplink ACKs
  and TX_ACK-correlated retry retention;
- Class B beacon synchronization for virtual gateways, deterministic ping-slot
  scheduling, queued downlinks, synchronization loss detection and runtime
  metrics/events;
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
expressed in meters. Real gateways are intentionally excluded from virtual
coverage and participate through their transport adapter. Real gateway
adapters are provided through `types.Options.GatewayAdapterFactory`; the concrete
`gateway.UDPFactory` implements the Semtech UDP packet-forwarder protocol used
by ChirpStack Gateway Bridge. It listens on the configurable local address
(`:1700` by default), routes `PUSH_DATA` and `PULL_DATA` by Gateway EUI,
responds with the corresponding acknowledgements and sends downlinks as
`PULL_RESP`. Transmissions overlap when they share a gateway, frequency,
bandwidth and spreading factor; overlapping
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
their last radio measurements for the frontend. Runtime gateways expose the
connection state, heartbeat timestamp, attempts, network error and ingress/
egress counters.

ABP devices are ready to transmit when the simulation starts. OTAA devices
first send a join request and become ready after a join accept in RX1 or RX2.
Confirmed uplinks keep the same frame counter and packet identity across
retransmissions, and stop after the device's configured retransmission limit.
RX1/RX2 delays use seconds, while RX2 duration uses milliseconds, matching the
device configuration model.

Class B devices synchronize with the beacon emitted by the nearest covered
active virtual gateway. Beacons follow the LoRaWAN 128-second cadence. The
engine uses a one-second ping-slot cadence as its deterministic simulation
default, so a synchronized device has a continuously scheduled next ping slot
and can receive one queued downlink per slot. A downlink queued through
`Engine.QueueClassBDownlink` remains durable while the device is waiting for a
beacon or recovering synchronization; it is removed only after transmission.
The queue validates the device class, FPort, data rate, UUID and regional
payload limit. The runtime snapshot exposes synchronization state, last beacon,
next ping slot and missed-beacon count. The event log records beacon schedule
and transmission, synchronization, missed beacons, ping-slot schedule/open,
downlink queue and downlink transmission events. If the next beacon is not
received within two beacon periods, the device loses synchronization and its
pending downlinks remain queued for the next successful synchronization.

Real gateways participate once their adapter reports `connected`. The engine
then emits the same beacon cadence and sends Class B beacon and downlink
packets through the adapter with an absolute UTC transmission time. Gateway
uplinks are parsed at the same boundary: valid ABP/OTAA frames are checked
with their MIC, OTAA Join-Accepts are encrypted with the AppKey, ABP data
frames update the device session and confirmed frames produce an authenticated
ACK. Pending
timed packets are included in the checkpoint and are retried when a gateway
reports a transport or TX_ACK failure. The
Semtech UDP adapter encodes these as timed `PULL_RESP` packets (`imme: false`,
`time: ...`) instead of immediate transmissions, allowing ChirpStack Gateway
Bridge and compatible packet-forwarders to schedule them on the gateway. The
beacon payload follows the LoRaWAN Class B 17-byte format and currently uses
the EU868 beacon radio defaults, matching the simulator's primary region.
Gateway-specific regional beacon plans can be added when the gateway contract
exposes its region.

Frontend streaming builds on top of this foundation in the following roadmap
phase.
