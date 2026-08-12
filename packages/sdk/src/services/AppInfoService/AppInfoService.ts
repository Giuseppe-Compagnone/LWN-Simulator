import { AppInfoResponse, StatusResponse } from "@lwn-simulator/contracts";
import { BaseService } from "../../models";

export class AppInfoService extends BaseService {
  private static _instance: AppInfoService | null = null;

  public static get instance() {
    if (!this._instance) {
      this._instance = new AppInfoService();
    }

    return this._instance;
  }

  private constructor() {
    super("app-info");
  }

  public status = async (): Promise<StatusResponse> => {
    const res: StatusResponse = await this.apiCaller.get("/status");

    return res;
  };

  public appInfo = async (): Promise<AppInfoResponse> => {
    const res: AppInfoResponse = await this.apiCaller.get("/info");

    return res;
  };
}
