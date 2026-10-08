"use client";

import { useCallback, useEffect, useMemo, useState } from "react";
import { GatewayBridgeConfig } from "@lwn-simulator/contracts";
import { ProfileConfigurationService } from "./ProfileConfigurationService";
import ProfileConfigurationServiceContext from "./ProfileConfigurationServiceContext";
import { ProfileConfigurationServiceProviderProps } from "./ProfileConfigurationService.types";

const ProfileConfigurationServiceProvider = (
  props: ProfileConfigurationServiceProviderProps,
) => {
  const service = useMemo(
    () => new ProfileConfigurationService(props.baseUrl),
    [props.baseUrl],
  );
  const [gatewayBridgeConfig, setGatewayBridgeConfig] =
    useState<GatewayBridgeConfig | null>(null);

  const getGatewayBridge = useCallback(async () => {
    const config = await service.getGatewayBridge();
    setGatewayBridgeConfig(config);
    return config;
  }, [service]);

  const updateGatewayBridge = useCallback(
    async (config: GatewayBridgeConfig) => {
      const updated = await service.updateGatewayBridge(config);
      setGatewayBridgeConfig(updated);
      return updated;
    },
    [service],
  );

  useEffect(() => {
    void Promise.resolve()
      .then(getGatewayBridge)
      .catch(() => setGatewayBridgeConfig(null));
  }, [getGatewayBridge]);

  const value = useMemo(
    () => ({
      gatewayBridgeConfig,
      getGatewayBridge,
      updateGatewayBridge,
    }),
    [gatewayBridgeConfig, getGatewayBridge, updateGatewayBridge],
  );

  return (
    <ProfileConfigurationServiceContext.Provider value={value}>
      {props.children}
    </ProfileConfigurationServiceContext.Provider>
  );
};

export default ProfileConfigurationServiceProvider;
