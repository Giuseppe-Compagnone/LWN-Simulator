import storage from "@/utils/storage";
import { Theme } from "./ThemeService.types";

export class ThemeService {
  private static _instance: ThemeService | null = null;

  public static get instance() {
    if (!this._instance) {
      this._instance = new ThemeService();
    }

    return this._instance;
  }

  getDefaultTheme = async (): Promise<Theme> => {
    const cached = await storage.get("theme");

    if (cached) {
      return cached as Theme;
    }

    const prefersDark = window.matchMedia(
      "(prefers-color-scheme: dark)",
    ).matches;

    return prefersDark ? Theme.Dark : Theme.Light;
  };
}
