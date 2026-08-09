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
}
