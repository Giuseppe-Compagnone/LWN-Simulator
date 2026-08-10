"use client";

import { createContext } from "react";
import { AppInfoServiceContent } from "./AppInfoService.types";

const AppInfoServiceContext = createContext<AppInfoServiceContent>({
  status: () => new Promise(() => {}),
});

export default AppInfoServiceContext;
