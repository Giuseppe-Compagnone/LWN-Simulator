import { useCallback, useEffect, useMemo, useState } from "react";
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
import { NotificationHandler } from "@lwn-simulator/ui-components";
import { DeviceService } from "./DeviceService";

const DeviceServiceProvider = (props: DeviceServiceProviderProps) => {
  // States
  const [devices, setDevices] = useState<Array<Device> | null>(null);

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

  // Effects
  useEffect(() => {
    (async () => {
      try {
        const res = await getDevices();

        setDevices(res);
      } catch {
        NotificationHandler.instance.error("Failed to load devices");
      }
    })();
  }, [getDevices]);

  // Memos
  const value = useMemo(
    (): DeviceServiceContent => ({
      createDevice,
      getDevice,
      getDevices,
      updateDevice,
      deleteDevice,
      devices,
    }),
    [createDevice, getDevice, getDevices, updateDevice, deleteDevice, devices],
  );

  return (
    <DeviceServiceContext.Provider value={value}>
      {props.children}
    </DeviceServiceContext.Provider>
  );
};

export default DeviceServiceProvider;
