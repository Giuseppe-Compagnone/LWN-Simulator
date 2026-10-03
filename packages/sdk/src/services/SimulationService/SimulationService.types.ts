import {
  SimulationConfig,
  SimulationEvent,
  SimulationSnapshot,
} from "@lwn-simulator/contracts";
import { PropsWithChildren } from "react";

export type SimulationConnectionState =
  | "disconnected"
  | "connecting"
  | "connected";

export interface SimulationServiceContent {
  start: (config: SimulationConfig) => Promise<SimulationSnapshot>;
  pause: () => Promise<SimulationSnapshot>;
  resume: () => Promise<SimulationSnapshot>;
  stop: () => Promise<SimulationSnapshot>;
  getSnapshot: () => Promise<SimulationSnapshot>;
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
