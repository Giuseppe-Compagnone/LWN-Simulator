import { CreateDeviceRequest, Device } from "@lwn-simulator/contracts";
import { PropsWithChildren } from "react";

export interface DeviceServiceContent {
  createDevice: (req: CreateDeviceRequest) => Promise<Device>;
}

export interface DeviceServiceProviderProps extends PropsWithChildren {}
