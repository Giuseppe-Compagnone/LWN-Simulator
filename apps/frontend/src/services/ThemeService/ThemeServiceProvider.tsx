"use client";

import { ThemeServiceProviderProps } from "./ThemeService.types";
import { ThemeServiceContext } from "./ThemeServiceContext";

const ThemeServiceProvider = (props: ThemeServiceProviderProps) => {
  return (
    <ThemeServiceContext.Provider value={{}}>
      {props.children}
    </ThemeServiceContext.Provider>
  );
};

export default ThemeServiceProvider;
