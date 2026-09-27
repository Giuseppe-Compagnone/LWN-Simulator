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
import Store from "electron-store";

import { validateRemote } from "../remote/remote-validator";

function isLauncherSender(event: Electron.IpcMainInvokeEvent) {
  return event.sender.getURL().startsWith("file://");
}

export function registerIpcHandlers() {
  const store = new Store();

  ipcMain.handle("connect-local", async (event) => {
    if (!isLauncherSender(event)) {
      throw new Error("Only the launcher can start a connection");
    }

    const window = getMainWindow();

    if (!window) {
      return;
    }

    const port = await startBackend();

    await waitForServer(port);

    await window.loadURL(`http://localhost:${port}`);

    setConnected(true);
  });

  ipcMain.handle("connect-remote", async (event, url: string) => {
    if (!isLauncherSender(event)) {
      throw new Error("Only the launcher can start a connection");
    }

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

  ipcMain.handle("sync-storage", (_, key: string, value: string) => {
    store.set(key, value);
  });

  ipcMain.handle("get-storage", (_, key: string) => {
    return store.get(key);
  });
}
