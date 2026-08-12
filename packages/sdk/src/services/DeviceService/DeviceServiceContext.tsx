import { createContext } from "react";
import { DeviceServiceContent } from "./DeviceService.types";

const DeviceServiceContext = createContext<DeviceServiceContent>({
  createDevice: () => new Promise(() => {}),
});

export default DeviceServiceContext;
