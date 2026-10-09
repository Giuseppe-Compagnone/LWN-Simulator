import { GatewayBridgeConfig } from "@lwn-simulator/contracts";
import { PropsWithChildren } from "react";

export interface ProfileConfigurationServiceContent {
  gatewayBridgeConfig: GatewayBridgeConfig | null;
  getGatewayBridge: () => Promise<GatewayBridgeConfig>;
  updateGatewayBridge: (
    config: GatewayBridgeConfig,
  ) => Promise<GatewayBridgeConfig>;
}

export interface ProfileConfigurationServiceProviderProps
  extends PropsWithChildren {
  baseUrl: string;
  profileID: string;
}
