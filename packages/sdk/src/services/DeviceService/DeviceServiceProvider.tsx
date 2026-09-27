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
import { DeviceService } from "./DeviceService";

const DeviceServiceProvider = (props: DeviceServiceProviderProps) => {
  // States
  const [devices, setDevices] = useState<Array<Device> | null>(null);
  const [error, setError] = useState<Error | null>(null);

  // Memos
  const service = useMemo(
    () => new DeviceService(props.baseUrl),
    [props.baseUrl],
  );

  // Callbacks
  const createDevice = useCallback(
    async (req: CreateDeviceRequest): Promise<Device> => {
      const res = await service.createDevice(req);

      setError(null);

      setDevices((prev) => {
        if (!prev) return [res];
        return [...prev, res];
      });

      return res;
    },
    [service],
  );

  const getDevice = useCallback(
    async (req: GetDeviceRequest): Promise<Device> => {
      return service.getDevice(req);
    },
    [service],
  );

  const getDevices = useCallback(async (): Promise<Array<Device>> => {
    const res = await service.getDevices();

    setDevices(res);
    setError(null);

    return res;
  }, [service]);

  const updateDevice = useCallback(
    async (req: UpdateDeviceRequest): Promise<Device> => {
      const res = await service.updateDevice(req);

      setDevices((prev) =>
        prev?.map((device) => (device.id === res.id ? res : device)) ?? null,
      );
      setError(null);

      return res;
    },
    [service],
  );

  const deleteDevice = useCallback(
    async (req: DeleteDeviceRequest): Promise<void> => {
      await service.deleteDevice(req);

      setDevices((prev) =>
        prev?.filter((device) => device.id !== req.id) ?? null,
      );
      setError(null);
    },
    [service],
  );

  // Effects
  useEffect(() => {
    (async () => {
      try {
        const res = await getDevices();

        setDevices(res);
      } catch (err) {
        setDevices([]);
        setError(
          err instanceof Error ? err : new Error("Failed to load devices"),
        );
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
      error,
    }),
    [
      createDevice,
      getDevice,
      getDevices,
      updateDevice,
      deleteDevice,
      devices,
      error,
    ],
  );

  return (
    <DeviceServiceContext.Provider value={value}>
      {props.children}
    </DeviceServiceContext.Provider>
  );
};

export default DeviceServiceProvider;
