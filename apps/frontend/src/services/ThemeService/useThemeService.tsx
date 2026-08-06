"use client";

import { useContext } from "react";
import { ThemeServiceContext } from "./ThemeServiceContext";

export const useThemeService = () => {
  const context = useContext(ThemeServiceContext);

  if (context === "undefinded") {
    throw new Error(
      "useThemeService must be use inside `ThemeServiceProvider`",
    );
  }

  return context;
};
