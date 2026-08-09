import { useContext } from "react";
import AppInfoServiceContext from "./AppInfoServiceContext";

const useAppInfoService = () => {
  const context = useContext(AppInfoServiceContext);

  if (context === undefined) {
    throw new Error(
      "useThemeService must be use inside `ThemeServiceProvider`",
    );
  }

  return context;
};
