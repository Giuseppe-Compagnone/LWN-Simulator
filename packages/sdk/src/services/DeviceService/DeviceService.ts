import {
  CreateDeviceRequest,
  CreateDeviceResponse,
  Device,
} from "@lwn-simulator/contracts";
import { BaseService } from "../../models";

export class DeviceService extends BaseService {
  private static _instance: DeviceService | null = null;

  public static get instance() {
    if (!this._instance) {
      this._instance = new DeviceService();
    }

    return this._instance;
  }

  private constructor() {
    super("device");
  }

  public createDevice = async (req: CreateDeviceRequest): Promise<Device> => {
    const device: CreateDeviceResponse = await this.apiCaller.post(
      "/create-device",
      req,
    );

    return device.device;
  };
}
