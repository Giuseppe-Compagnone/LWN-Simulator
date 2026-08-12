import { useContext } from "react";
import AppInfoServiceContext from "./AppInfoServiceContext";

export const useAppInfoService = () => {
  const context = useContext(AppInfoServiceContext);

  if (context === undefined) {
    throw new Error(
      "useThemeService must be used inside `ThemeServiceProvider`",
    );
  }

  return context;
};
