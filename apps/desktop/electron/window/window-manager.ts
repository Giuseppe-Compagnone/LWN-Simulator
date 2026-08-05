import { BrowserWindow, app } from "electron";
import path from "path";
import { createMenu } from "./window-menu";

let mainWindow: BrowserWindow | undefined;
let disconnectItem: Electron.MenuItem | null;

export function createMainWindow() {
  mainWindow = new BrowserWindow({
    width: 1280,
    height: 800,
    webPreferences: {
      preload: path.join(__dirname, "..", "preload.js"),
    },
  });

  disconnectItem = createMenu();

  mainWindow.loadFile(getLauncherPath());
}

export function setConnected(value: boolean) {
  if (disconnectItem) {
    disconnectItem.enabled = value;
  }
}

export function getMainWindow() {
  return mainWindow;
}

export function getLauncherPath() {
  if (app.isPackaged) {
    return path.join(process.resourcesPath, "launcher", "out", "index.html");
  }

  return path.join(process.cwd(), "..", "launcher", "out", "index.html");
}
