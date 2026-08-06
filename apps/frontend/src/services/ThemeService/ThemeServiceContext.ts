"use client";

import { createContext } from "react";
import { ThemeServiceContent } from "./ThemeService.types";

export const ThemeServiceContext = createContext<ThemeServiceContent>({});
