import {
  SimulationActionResponse,
  SimulationConfig,
  SimulationDownlinkRequest,
  SimulationEvent,
  SimulationEventsResponse,
  SimulationMACCommandRequest,
  SimulationSnapshot,
  SimulationUplinkRequest,
} from "@lwn-simulator/contracts";
import { PropsWithChildren } from "react";

export type SimulationConnectionState =
  | "disconnected"
  | "connecting"
  | "connected";

export interface GetSimulationEventsOptions {
  afterSequence?: number;
  limit?: number;
}

export interface SimulationServiceContent {
  start: (config: SimulationConfig) => Promise<SimulationSnapshot>;
  pause: () => Promise<SimulationSnapshot>;
  resume: () => Promise<SimulationSnapshot>;
  stop: () => Promise<SimulationSnapshot>;
  getSnapshot: () => Promise<SimulationSnapshot>;
  queueUplink: (
    req: SimulationUplinkRequest,
  ) => Promise<SimulationActionResponse>;
  queueDownlink: (
    req: SimulationDownlinkRequest,
  ) => Promise<SimulationActionResponse>;
  queueMACCommand: (
    req: SimulationMACCommandRequest,
  ) => Promise<SimulationActionResponse>;
  getEvents: (
    options?: GetSimulationEventsOptions,
  ) => Promise<SimulationEventsResponse>;
  getMetrics: () => Promise<string>;
  connect: () => void;
  disconnect: () => void;
  snapshot: SimulationSnapshot | null;
  events: Array<SimulationEvent>;
  connectionState: SimulationConnectionState;
  error: Error | null;
}

export interface SimulationServiceProviderProps extends PropsWithChildren {
  baseUrl: string;
}
