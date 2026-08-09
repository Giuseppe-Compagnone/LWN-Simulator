import { AppInfoServiceProviderProps } from "./AppInfoService.types";
import AppInfoServiceContext from "./AppInfoServiceContext";

const AppInfoServiceProvider = (props: AppInfoServiceProviderProps) => {
  return (
    <AppInfoServiceContext.Provider value={{}}>
      {props.children}
    </AppInfoServiceContext.Provider>
  );
};

export default AppInfoServiceProvider;
