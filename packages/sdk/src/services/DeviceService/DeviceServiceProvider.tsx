import { useCallback, useMemo } from "react";
import {
  DeviceServiceContent,
  DeviceServiceProviderProps,
} from "./DeviceService.types";
import DeviceServiceContext from "./DeviceServiceContext";
import {
  CreateDeviceRequest,
  DeleteDeviceRequest,
  Device,
  GetDeviceRequest,
  UpdateDeviceRequest,
} from "@lwn-simulator/contracts";
import { DeviceService } from "./DeviceService";

const DeviceServiceProvider = (props: DeviceServiceProviderProps) => {
  // Callbacks
  const createDevice = useCallback(
    async (req: CreateDeviceRequest): Promise<Device> => {
      return DeviceService.instance.createDevice(req);
    },
    [],
  );

  const getDevice = useCallback(
    async (req: GetDeviceRequest): Promise<Device> => {
      return DeviceService.instance.getDevice(req);
    },
    [],
  );

  const getDevices = useCallback(async (): Promise<Array<Device>> => {
    return DeviceService.instance.getDevices();
  }, []);

  const updateDevice = useCallback(
    async (req: UpdateDeviceRequest): Promise<Device> => {
      return DeviceService.instance.updateDevice(req);
    },
    [],
  );

  const deleteDevice = useCallback(
    async (req: DeleteDeviceRequest): Promise<void> => {
      return DeviceService.instance.deleteDevice(req);
    },
    [],
  );

  // Memos
  const value = useMemo(
    (): DeviceServiceContent => ({
      createDevice,
      getDevice,
      getDevices,
      updateDevice,
      deleteDevice,
    }),
    [createDevice, getDevice, getDevices, updateDevice, deleteDevice],
  );

  return (
    <DeviceServiceContext.Provider value={value}>
      {props.children}
    </DeviceServiceContext.Provider>
  );
};

export default DeviceServiceProvider;
