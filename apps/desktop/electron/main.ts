import { app, session } from "electron";
import { registerIpcHandlers } from "./ipc";
import { installLinuxDesktopEntry } from "./system";
import { createMainWindow } from "./window";
import { stopBackend } from "./backend";

const isGeolocationPermission = (permission: string): boolean =>
  permission === "geolocation" || permission === "geolocation-approximate";

app.whenReady().then(() => {
  session.defaultSession.setPermissionCheckHandler(
    (webContents, permission) => {
      if (isGeolocationPermission(permission)) {
        return true;
      }
      return false;
    },
  );

  session.defaultSession.setPermissionRequestHandler(
    (webContents, permission, callback) => {
      if (isGeolocationPermission(permission)) {
        callback(true);
      } else {
        callback(false);
      }
    },
  );

  registerIpcHandlers();

  if (app.isPackaged) {
    installLinuxDesktopEntry();
  }

  createMainWindow();
});

app.on("window-all-closed", () => {
  stopBackend();

  if (process.platform !== "darwin") {
    app.quit();
  }
});
