import { GatewayBridgeConfig } from "@lwn-simulator/contracts";

export const gatewayBridgeStorageKey = "lwn-simulator.gateway-bridge";

export const defaultGatewayBridgeConfig: GatewayBridgeConfig = {
  enabled: false,
  address: "0.0.0.0",
  port: 1700,
};

export const readGatewayBridgeConfig = (): GatewayBridgeConfig => {
  if (typeof window === "undefined") return defaultGatewayBridgeConfig;

  try {
    const stored = window.localStorage.getItem(gatewayBridgeStorageKey);
    if (!stored) return defaultGatewayBridgeConfig;

    const parsed = JSON.parse(stored) as Partial<GatewayBridgeConfig>;
    return {
      enabled: parsed.enabled === true,
      address:
        typeof parsed.address === "string" && parsed.address.trim()
          ? parsed.address.trim()
          : defaultGatewayBridgeConfig.address,
      port:
        typeof parsed.port === "number" && parsed.port > 0
          ? parsed.port
          : defaultGatewayBridgeConfig.port,
    };
  } catch {
    return defaultGatewayBridgeConfig;
  }
};

export const saveGatewayBridgeConfig = (
  config: GatewayBridgeConfig,
): GatewayBridgeConfig => {
  const normalized: GatewayBridgeConfig = {
    enabled: config.enabled,
    address: config.address?.trim() || defaultGatewayBridgeConfig.address,
    port: config.port || defaultGatewayBridgeConfig.port,
  };

  if (typeof window !== "undefined") {
    window.localStorage.setItem(
      gatewayBridgeStorageKey,
      JSON.stringify(normalized),
    );
  }

  return normalized;
};
