import {
  Device,
  SimulationActionResponse,
  SimulationDownlinkRequest,
  SimulationMACCommand,
  SimulationMACCommandRequest,
  SimulationMACCommandType,
  SimulationUplinkRequest,
} from "@lwn-simulator/contracts";

export enum SimulationCommandMode {
  Uplink = "uplink",
  Downlink = "downlink",
  MACCommand = "mac-command",
}

export type SimulationMACCommandFieldName = Exclude<
  keyof SimulationMACCommand,
  "type"
>;

export interface SimulationMACCommandFieldDescriptor {
  name: SimulationMACCommandFieldName;
  label: string;
  input: "number" | "checkbox";
  required?: boolean;
  minimum?: number;
  maximum?: number;
  step?: number;
  placeholder?: string;
}

export type SimulationMACCommandFields = Record<
  SimulationMACCommandType,
  Array<SimulationMACCommandFieldDescriptor>
>;

export interface SimulationCommandPanelProps {
  devices: Array<Device>;
  enabled: boolean;
  onQueueUplink: (
    request: SimulationUplinkRequest,
  ) => Promise<SimulationActionResponse>;
  onQueueDownlink: (
    request: SimulationDownlinkRequest,
  ) => Promise<SimulationActionResponse>;
  onQueueMACCommand: (
    request: SimulationMACCommandRequest,
  ) => Promise<SimulationActionResponse>;
}

