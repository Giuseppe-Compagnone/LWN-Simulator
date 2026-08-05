import { BrowserWindow, app } from "electron";
import path from "path";

let mainWindow: BrowserWindow | undefined;

export function createMainWindow() {
  mainWindow = new BrowserWindow({
    width: 1280,
    height: 800,
    webPreferences: {
      preload: path.join(__dirname, "..", "preload.js"),
    },
  });

  mainWindow.loadFile(getLauncherPath());
}

export function getMainWindow() {
  return mainWindow;
}

function getLauncherPath() {
  if (app.isPackaged) {
    return path.join(process.resourcesPath, "launcher", "out", "index.html");
  }

  return path.join(process.cwd(), "..", "launcher", "out", "index.html");
}
