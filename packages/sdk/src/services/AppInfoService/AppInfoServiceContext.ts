import { createContext } from "react";
import { AppInfoServiceContent } from "./AppInfoService.types";

const AppInfoServiceContext = createContext<AppInfoServiceContent>({});

export default AppInfoServiceContext;
