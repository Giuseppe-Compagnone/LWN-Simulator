"use client";

import { useCallback, useEffect, useMemo, useState } from "react";
import {
  GatewayBridgeConfig,
  RealtimeWebSocketMessage,
  RealtimeWebSocketMessageType,
} from "@lwn-simulator/contracts";
import { useWebSocket } from "../../websocket";
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
  const { subscribe } = useWebSocket();

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

  const handleRealtimeMessage = useCallback(
    (rawMessage: string) => {
      try {
        const message = JSON.parse(rawMessage) as RealtimeWebSocketMessage;
        if (message.profileID && message.profileID !== props.profileID) return;
        if (
          message.type ===
            RealtimeWebSocketMessageType.ProfileGatewayBridgeUpdated &&
          message.gatewayBridge
        ) {
          setGatewayBridgeConfig(message.gatewayBridge);
        }
      } catch {
        // Ignore messages owned by another realtime consumer.
      }
    },
    [props.profileID],
  );

  useEffect(
    () => subscribe(handleRealtimeMessage),
    [handleRealtimeMessage, subscribe],
  );

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
