import { createContext } from "react";
import { DeviceServiceContent } from "./DeviceService.types";

const DeviceServiceContext = createContext<DeviceServiceContent>({
  createDevice: () => new Promise(() => {}),
  getDevice: () => new Promise(() => {}),
  getDevices: () => new Promise(() => {}),
  updateDevice: () => new Promise(() => {}),
  deleteDevice: () => new Promise(() => {}),
});

export default DeviceServiceContext;
