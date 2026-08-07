"use client";

import { useEffect, useState } from "react";
import { Theme, ThemeServiceProviderProps } from "./ThemeService.types";
import { ThemeServiceContext } from "./ThemeServiceContext";
import { Inter, JetBrains_Mono } from "next/font/google";
import cn from "classnames";

const inter = Inter({
  subsets: ["latin"],
  variable: "--primary-font",
});

const jetBrains = JetBrains_Mono({
  subsets: ["latin"],
  variable: "--secondary-font",
});

const ThemeServiceProvider = (props: ThemeServiceProviderProps) => {
  // States
  const [theme, setTheme] = useState<Theme>(Theme.Light);

  // Effects
  useEffect(() => {
    document.body.className = cn(
      `${theme}-theme`,
      inter.variable,
      jetBrains.variable,
    );
  }, [theme]);

  return (
    <ThemeServiceContext.Provider value={{ theme, setTheme }}>
      {props.children}
    </ThemeServiceContext.Provider>
  );
};

export default ThemeServiceProvider;
