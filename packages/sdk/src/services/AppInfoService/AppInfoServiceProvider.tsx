"use client";

import { useCallback, useMemo } from "react";
import {
  AppInfoServiceContent,
  AppInfoServiceProviderProps,
} from "./AppInfoService.types";
import AppInfoServiceContext from "./AppInfoServiceContext";
import { AppInfoService } from "./AppInfoService";
import { AppInfoResponse, StatusResponse } from "@lwn-simulator/contracts";

const AppInfoServiceProvider = (props: AppInfoServiceProviderProps) => {
  // States

  // Hooks

  // Memos
  const service = useMemo(
    () => new AppInfoService(props.baseUrl),
    [props.baseUrl],
  );

  // Callbacks
  const status = useCallback(async (): Promise<StatusResponse> => {
    return service.status();
  }, [service]);

  const appInfo = useCallback(async (): Promise<AppInfoResponse> => {
    return service.appInfo();
  }, [service]);

  // Memos
  const value = useMemo(
    (): AppInfoServiceContent => ({
      status,
      appInfo,
    }),
    [status, appInfo],
  );

  // Effects

  return (
    <AppInfoServiceContext.Provider value={value}>
      {props.children}
    </AppInfoServiceContext.Provider>
  );
};

export default AppInfoServiceProvider;
