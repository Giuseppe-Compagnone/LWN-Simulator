import { ipcMain } from "electron";
import {
  startBackend,
  stopBackend,
  waitForServer,
} from "../backend/backend-manager";

import {
  getLauncherPath,
  getMainWindow,
  setConnected,
} from "../window/window-manager";

import { validateRemote } from "../remote/remote-validator";

export function registerIpcHandlers() {
  ipcMain.handle("connect-local", async () => {
    const window = getMainWindow();

    if (!window) {
      return;
    }

    const port = await startBackend();

    await waitForServer();

    await window.loadURL(`http://localhost:${port}`);

    setConnected(true);
  });

  ipcMain.handle("connect-remote", async (_, url: string) => {
    const window = getMainWindow();

    if (!window) {
      return;
    }

    try {
      await validateRemote(url);

      await window.loadURL(url);

      setConnected(true);

      return {
        success: true,
      };
    } catch (err) {
      return {
        success: false,
        message: err instanceof Error ? err.message : String(err),
      };
    }
  });

  ipcMain.handle("disconnect", async () => {
    stopBackend();

    const window = getMainWindow();

    if (window) {
      window.loadFile(getLauncherPath());
    }

    setConnected(false);
  });
}
