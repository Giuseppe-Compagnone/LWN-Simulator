import {
  CreateGatewayRequest,
  DeleteGatewayRequest,
  Gateway,
  GetGatewayRequest,
  UpdateGatewayRequest,
} from "@lwn-simulator/contracts";
import { PropsWithChildren } from "react";

export interface GatewayServiceContent {
  createGateway: (req: CreateGatewayRequest) => Promise<Gateway>;
  getGateway: (req: GetGatewayRequest) => Promise<Gateway>;
  getGateways: () => Promise<Array<Gateway>>;
  updateGateway: (req: UpdateGatewayRequest) => Promise<Gateway>;
  deleteGateway: (req: DeleteGatewayRequest) => Promise<void>;
  gateways: Array<Gateway> | null;
  error: Error | null;
}

export interface GatewayServiceProviderProps extends PropsWithChildren {
  baseUrl: string;
  profileID?: string;
}
