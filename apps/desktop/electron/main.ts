import { app } from "electron";
import { registerIpcHandlers } from "./ipc";
import { installLinuxDesktopEntry } from "./system";
import { createMainWindow } from "./window";
import { stopBackend } from "./backend";

app.whenReady().then(() => {
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
