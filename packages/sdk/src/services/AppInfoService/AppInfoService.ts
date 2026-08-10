import { StatusResponse } from "@lwn-simulator/contracts";
import { BaseService } from "../../models";

export class AppInfoService extends BaseService {
  private static _instance: AppInfoService | null = null;

  public static get instance() {
    if (!this._instance) {
      this._instance = new AppInfoService();
    }

    return this._instance;
  }

  constructor() {
    super("app-info");
  }

  public status = async (): Promise<StatusResponse> => {
    const res: StatusResponse = await this.apiCaller.get("/status");

    return res;
  };
}
