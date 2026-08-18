import {
  CreateDeviceRequest,
  DeleteDeviceRequest,
  Device,
  GetDeviceRequest,
  UpdateDeviceRequest,
} from "@lwn-simulator/contracts";
import { PropsWithChildren } from "react";

/**
 * Content exposed by the device service.
 *
 * Provides operations for creating, retrieving, updating, and deleting devices.
 */
export interface DeviceServiceContent {
  /**
   * Creates a new device.
   *
   * @param req Request containing the device data.
   * @returns The newly created device.
   */
  createDevice: (req: CreateDeviceRequest) => Promise<Device>;

  /**
   * Retrieves a device by its identifier.
   *
   * @param req Request containing the device identifier.
   * @returns The requested device.
   */
  getDevice: (req: GetDeviceRequest) => Promise<Device>;

  /**
   * Retrieves all available devices.
   *
   * @returns A collection containing all devices.
   */
  getDevices: () => Promise<Array<Device>>;

  /**
   * Updates an existing device.
   *
   * @param req Request containing the device identifier and updated data.
   * @returns The updated device.
   */
  updateDevice: (req: UpdateDeviceRequest) => Promise<Device>;

  /**
   * Deletes an existing device.
   *
   * @param req Request containing the device identifier.
   */
  deleteDevice: (req: DeleteDeviceRequest) => Promise<void>;
  /**
   * Currently loaded devices.
   *
   * A `null` value indicates that the device collection has not been loaded
   * yet.
   */
  devices: Array<Device> | null;
}

/**
 * Properties for configuring the device service provider.
 *
 * The provider makes device service operations available to its descendant
 * components.
 */
export interface DeviceServiceProviderProps extends PropsWithChildren {}
