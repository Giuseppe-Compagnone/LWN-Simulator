"use client";

import { useCallback, useEffect, useMemo, useState } from "react";
import { AppInfoServiceProviderProps } from "./AppInfoService.types";
import AppInfoServiceContext from "./AppInfoServiceContext";
import { ApiCaller } from "../../models";
import { AppInfoService } from "./AppInfoService";
import { AppInfoResponse, StatusResponse } from "@lwn-simulator/contracts";

const AppInfoServiceProvider = (props: AppInfoServiceProviderProps) => {
  ApiCaller.baseUrl = props.baseUrl;

  // Functions
  const status = useCallback(async (): Promise<StatusResponse> => {
    return AppInfoService.instance.status();
  }, []);

  const appInfo = useCallback(async (): Promise<AppInfoResponse> => {
    return AppInfoService.instance.appInfo();
  }, []);

  // Effects
  useEffect(() => {
    ApiCaller.baseUrl = props.baseUrl;
  }, [props.baseUrl]);

  // Memos

  const value = useMemo(
    () => ({
      status,
      appInfo,
    }),
    [status],
  );

  return (
    <AppInfoServiceContext.Provider value={value}>
      {props.children}
    </AppInfoServiceContext.Provider>
  );
};

export default AppInfoServiceProvider;
