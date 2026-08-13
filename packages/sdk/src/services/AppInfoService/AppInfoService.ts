import { AppInfoResponse, StatusResponse } from "@lwn-simulator/contracts";
import { BaseService } from "../../models";

/**
 * Service responsible for retrieving application status and information
 * from the backend.
 *
 * Implements the singleton pattern and uses the application information
 * API namespace configured by the base service.
 */
export class AppInfoService extends BaseService {
  private static _instance: AppInfoService | null = null;

  /**
   * Returns the singleton instance of the application information service.
   */
  public static get instance() {
    if (!this._instance) {
      this._instance = new AppInfoService();
    }

    return this._instance;
  }

  private constructor() {
    super("app-info");
  }

  /**
   * Retrieves the current application status from the backend.
   *
   * @returns A promise resolving to the application's status information.
   */
  public status = async (): Promise<StatusResponse> => {
    const res: StatusResponse = await this.apiCaller.get("/status");

    return res;
  };

  /**
   * Retrieves information about the application from the backend.
   *
   * @returns A promise resolving to the application's information.
   */
  public appInfo = async (): Promise<AppInfoResponse> => {
    const res: AppInfoResponse = await this.apiCaller.get("/info");

    return res;
  };
}
