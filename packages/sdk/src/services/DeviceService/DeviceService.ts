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
}
