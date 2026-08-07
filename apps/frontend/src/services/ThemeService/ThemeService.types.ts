import { PropsWithChildren } from "react";

export enum Theme {
  Light = "light",
  Dark = "dark",
}

export interface ThemeServiceContent {
  theme: Theme;
  setTheme: (theme: Theme) => void;
}

export interface ThemeServiceProviderProps extends PropsWithChildren {}
