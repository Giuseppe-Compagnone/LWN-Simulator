"use client";

import { useCallback, useEffect, useMemo, useState } from "react";
import { AppInfoServiceProviderProps } from "./AppInfoService.types";
import AppInfoServiceContext from "./AppInfoServiceContext";
import { ApiCaller } from "../../models";
import { AppInfoService } from "./AppInfoService";
import { StatusResponse } from "@lwn-simulator/contracts";

const AppInfoServiceProvider = (props: AppInfoServiceProviderProps) => {
  ApiCaller.baseUrl = props.baseUrl;

  const status = useCallback(async (): Promise<StatusResponse> => {
    return AppInfoService.instance.status();
  }, []);

  const value = useMemo(
    () => ({
      status,
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
