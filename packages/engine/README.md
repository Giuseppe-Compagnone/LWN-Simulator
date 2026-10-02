# Simulation engine

The engine is the deterministic, server-side core of the simulator. It owns
simulation time, lifecycle transitions, the runtime registry, the scheduler,
and the event log.

Shared data types are defined in the OpenAPI YAML contracts under
`packages/contracts/schemas/Simulation` and generated for Go and TypeScript.
Implementation-only types live under `types/` when they are needed by the
engine runtime and test clocks.

## Current scope

The foundation currently provides:

- lifecycle control: start, pause, resume, and stop;
- real and deterministic manual clocks;
- a priority-based event scheduler;
- validated device and gateway runtime registration;
- immutable runtime snapshots;
- ordered event records and a live event channel;
- deterministic tests, including race detection.

Traffic generation, LoRaWAN behavior, radio propagation, real gateway
transports, and frontend streaming build on top of this foundation in the
following roadmap phases.
