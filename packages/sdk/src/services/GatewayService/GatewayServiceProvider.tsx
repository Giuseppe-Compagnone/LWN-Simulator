import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import {
  CreateGatewayRequest,
  DeleteGatewayRequest,
  Gateway,
  GetGatewayRequest,
  RealtimeWebSocketMessage,
  RealtimeWebSocketMessageType,
  UpdateGatewayRequest,
} from "@lwn-simulator/contracts";
import GatewayServiceContext from "./GatewayServiceContext";
import { GatewayService } from "./GatewayService";
import {
  GatewayServiceContent,
  GatewayServiceProviderProps,
} from "./GatewayService.types";
import { useWebSocket } from "../../websocket";

const GatewayServiceProvider = (props: GatewayServiceProviderProps) => {
  const [gateways, setGateways] = useState<Array<Gateway> | null>(null);
  const [error, setError] = useState<Error | null>(null);
  const service = useMemo(() => new GatewayService(props.baseUrl), [props.baseUrl]);
  const { connectionState, subscribe } = useWebSocket();
  const hasConnectedRef = useRef(false);

  const createGateway = useCallback(async (req: CreateGatewayRequest) => {
    const res = await service.createGateway(req);
    setGateways((prev) => (prev ? [...prev, res] : [res]));
    setError(null);
    return res;
  }, [service]);

  const getGateway = useCallback((req: GetGatewayRequest) => service.getGateway(req), [service]);

  const getGateways = useCallback(async () => {
    const res = await service.getGateways();
    setGateways(res);
    setError(null);
    return res;
  }, [service]);

  const updateGateway = useCallback(async (req: UpdateGatewayRequest) => {
    const res = await service.updateGateway(req);
    setGateways((prev) => prev?.map((gateway) => gateway.id === res.id ? res : gateway) ?? null);
    setError(null);
    return res;
  }, [service]);

  const deleteGateway = useCallback(async (req: DeleteGatewayRequest) => {
    await service.deleteGateway(req);
    setGateways((prev) => prev?.filter((gateway) => gateway.id !== req.id) ?? null);
    setError(null);
  }, [service]);

  const handleRealtimeMessage = useCallback((rawMessage: string) => {
    try {
      const message = JSON.parse(rawMessage) as RealtimeWebSocketMessage;
      if (message.profileID && message.profileID !== props.profileID) return;
      if (
        (message.type === RealtimeWebSocketMessageType.GatewayCreated ||
          message.type === RealtimeWebSocketMessageType.GatewayUpdated) &&
        message.gateway
      ) {
        setGateways((previous) => {
          if (!previous) return [message.gateway as Gateway];
          const exists = previous.some((gateway) => gateway.id === message.gateway?.id);
          return exists
            ? previous.map((gateway) =>
                gateway.id === message.gateway?.id ? message.gateway as Gateway : gateway,
              )
            : [...previous, message.gateway as Gateway];
        });
      } else if (
        message.type === RealtimeWebSocketMessageType.GatewayDeleted &&
        message.resourceID
      ) {
        setGateways((previous) =>
          previous?.filter((gateway) => gateway.id !== message.resourceID) ?? null,
        );
      }
    } catch {
      // Ignore messages owned by another realtime consumer or malformed data.
    }
  }, [props.profileID]);

  useEffect(
    () => subscribe(handleRealtimeMessage),
    [handleRealtimeMessage, props.profileID, subscribe],
  );

  useEffect(() => {
    if (connectionState !== "connected") return;
    if (!hasConnectedRef.current) {
      hasConnectedRef.current = true;
      return;
    }

    void getGateways().catch(() => {
      setError(new Error("Failed to recover gateways after websocket reconnect"));
    });
  }, [connectionState, getGateways]);

  useEffect(() => {
    void Promise.resolve()
      .then(getGateways)
      .catch((err) => {
        setGateways([]);
        setError(
          err instanceof Error ? err : new Error("Failed to load gateways"),
        );
      });
  }, [getGateways]);

  const value = useMemo((): GatewayServiceContent => ({
    createGateway,
    getGateway,
    getGateways,
    updateGateway,
    deleteGateway,
    gateways,
    error,
  }), [createGateway, getGateway, getGateways, updateGateway, deleteGateway, gateways, error]);

  return <GatewayServiceContext.Provider value={value}>{props.children}</GatewayServiceContext.Provider>;
};

export default GatewayServiceProvider;
