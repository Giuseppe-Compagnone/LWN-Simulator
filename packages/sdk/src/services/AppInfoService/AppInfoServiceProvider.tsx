"use client";

import { useEffect, useState } from "react";
import { AppInfoServiceProviderProps } from "./AppInfoService.types";
import AppInfoServiceContext from "./AppInfoServiceContext";
import { ApiCaller } from "../../models";
import { AppInfoService } from "./AppInfoService";
import { StatusResponse } from "@lwn-simulator/contracts";

const AppInfoServiceProvider = (props: AppInfoServiceProviderProps) => {
  // States
  const [canMount, setCanMount] = useState<boolean>(false);

  // Effects
  useEffect(() => {
    ApiCaller.baseUrl = props.baseUrl;
    setCanMount(true);
  }, [props.baseUrl]);

  // Functions
  const status = async (): Promise<StatusResponse> => {
    return await AppInfoService.instance.status();
  };

  return canMount ? (
    <AppInfoServiceContext.Provider value={{ status }}>
      {props.children}
    </AppInfoServiceContext.Provider>
  ) : (
    <></>
  );
};

export default AppInfoServiceProvider;
