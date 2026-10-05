import { FormLogic, FormValue } from "@lwn-simulator/ui-components";

export interface GatewayBridgeFormValues {
  enabled: boolean;
  address: string;
  port: number;
}

export interface GatewayBridgeFormState {
  logic: FormLogic | null;
  values: GatewayBridgeFormValues;
  lastSaved: GatewayBridgeFormValues | null;
}

export type GatewayBridgeFormValueMap = Record<string, FormValue>;
