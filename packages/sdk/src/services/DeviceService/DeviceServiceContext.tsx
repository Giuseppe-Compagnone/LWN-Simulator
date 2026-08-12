import { createContext } from "react";
import { DeviceServiceContent } from "./DeviceService.types";

const DeviceServiceContext = createContext<DeviceServiceContent>({});

export default DeviceServiceContext;
