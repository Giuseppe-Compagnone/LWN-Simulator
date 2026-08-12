import { useCallback, useMemo } from "react";
import {
  DeviceServiceContent,
  DeviceServiceProviderProps,
} from "./DeviceService.types";
import DeviceServiceContext from "./DeviceServiceContext";
import { CreateDeviceRequest, Device } from "@lwn-simulator/contracts";
import { DeviceService } from "./DeviceService";

const DeviceServiceProvider = (props: DeviceServiceProviderProps) => {
  // Callbacks
  const createDevice = useCallback(
    async (req: CreateDeviceRequest): Promise<Device> => {
      return DeviceService.instance.createDevice(req);
    },
    [],
  );

  // Memos
  const value = useMemo(
    (): DeviceServiceContent => ({ createDevice }),
    [createDevice],
  );

  return (
    <DeviceServiceContext.Provider value={value}>
      {props.children}
    </DeviceServiceContext.Provider>
  );
};

export default DeviceServiceProvider;
