import { useCallback, useEffect, useMemo, useState } from "react";
import {
  CreateGatewayRequest,
  DeleteGatewayRequest,
  Gateway,
  GetGatewayRequest,
  UpdateGatewayRequest,
} from "@lwn-simulator/contracts";
import GatewayServiceContext from "./GatewayServiceContext";
import { GatewayService } from "./GatewayService";
import {
  GatewayServiceContent,
  GatewayServiceProviderProps,
} from "./GatewayService.types";

const GatewayServiceProvider = (props: GatewayServiceProviderProps) => {
  const [gateways, setGateways] = useState<Array<Gateway> | null>(null);
  const [error, setError] = useState<Error | null>(null);
  const service = useMemo(() => new GatewayService(props.baseUrl), [props.baseUrl]);

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
