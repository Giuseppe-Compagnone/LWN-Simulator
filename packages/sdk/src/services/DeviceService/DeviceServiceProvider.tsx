import { useMemo } from "react";
import {
  DeviceServiceContent,
  DeviceServiceProviderProps,
} from "./DeviceService.types";
import DeviceServiceContext from "./DeviceServiceContext";

const DeviceServiceProvider = (props: DeviceServiceProviderProps) => {
  // Memos
  const value = useMemo((): DeviceServiceContent => ({}), []);
  return (
    <DeviceServiceContext.Provider value={value}>
      {props.children}
    </DeviceServiceContext.Provider>
  );
};

export default DeviceServiceProvider;
