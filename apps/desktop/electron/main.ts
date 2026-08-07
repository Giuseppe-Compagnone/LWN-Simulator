import { app, session } from "electron";
import { registerIpcHandlers } from "./ipc";
import { installLinuxDesktopEntry } from "./system";
import { createMainWindow } from "./window";
import { stopBackend } from "./backend";

app.whenReady().then(() => {
  session.defaultSession.setPermissionCheckHandler(
    (webContents, permission) => {
      if (permission === "geolocation") {
        return true;
      }
      return false;
    },
  );

  session.defaultSession.setPermissionRequestHandler(
    (webContents, permission, callback) => {
      console.log("PERMISSION", permission);
      if (permission === "geolocation") {
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
