import { StatusResponse } from "@lwn-simulator/contracts";
import { PropsWithChildren } from "react";

/**
 * Represents the content exposed by the application information service.
 */
export interface AppInfoServiceContent {
  status: () => Promise<StatusResponse>;
}

/**
 * Properties for configuring the application information service provider.
 *
 * The provider makes application information service functionality available
 * to its descendant components.
 */
export interface AppInfoServiceProviderProps extends PropsWithChildren {
  /**
   * Base URL used to communicate with the application information service.
   */
  baseUrl: string;
}
