"use client";

import { useEffect, useState } from "react";
import { Theme, ThemeServiceProviderProps } from "./ThemeService.types";
import { ThemeServiceContext } from "./ThemeServiceContext";
import cn from "classnames";
import storage from "@/utils/storage";

const ThemeServiceProvider = (props: ThemeServiceProviderProps) => {
  // States
  const [theme, setTheme] = useState<Theme>(Theme.Dark);

  // Effects
  useEffect(() => {
    const cached = storage.get("theme");

    if (cached) {
      // eslint-disable-next-line react-hooks/set-state-in-effect
      setTheme(cached as Theme);
      return;
    }

    const prefersDark = window.matchMedia(
      "(prefers-color-scheme: dark)",
    ).matches;

    setTheme(prefersDark ? Theme.Dark : Theme.Light);
  }, []);

  useEffect(() => {
    document.body.className = cn(
      document.body.className.replace(/(dark|light)-theme/, "").trim(),
      `${theme}-theme`,
    );
    storage.set("theme", theme);
  }, [theme]);

  return (
    <ThemeServiceContext.Provider value={{ theme, setTheme }}>
      {props.children}
    </ThemeServiceContext.Provider>
  );
};

export default ThemeServiceProvider;
