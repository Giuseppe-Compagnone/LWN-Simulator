import { app, BrowserWindow, ipcMain } from "electron";
import { ChildProcess, spawn } from "child_process";
import path from "path";
import fs from "fs";
import os from "os";
import { exec } from "child_process";

let backendProcess: ChildProcess | undefined;
let mainWindow: BrowserWindow | undefined;

function startBackend() {
  const backendPath = app.isPackaged
    ? path.join(process.resourcesPath, "lwn-server")
    : path.join(process.cwd(), "assets", "lwn-server");

  backendProcess = spawn(backendPath, [], {
    stdio: "inherit",
  });
}

function waitForServer() {
  return new Promise<void>((resolve) => {
    setTimeout(resolve, 1500);
  });
}

function installLinuxDesktopEntry() {
  if (process.platform !== "linux") return;

  const applicationsDir = path.join(
    os.homedir(),
    ".local",
    "share",
    "applications",
  );

  const iconsDir = path.join(os.homedir(), ".local", "share", "icons");

  fs.mkdirSync(applicationsDir, {
    recursive: true,
  });

  fs.mkdirSync(iconsDir, {
    recursive: true,
  });

  const desktopPath = path.join(applicationsDir, "lwn-simulator.desktop");

  const iconPath = path.join(iconsDir, "lwn-simulator.png");

  fs.copyFileSync(path.join(process.resourcesPath, "icon.png"), iconPath);

  const desktopContent = `
    [Desktop Entry]
    Name=LWN-Simulator
    Exec=${process.env.APPIMAGE || app.getPath("exe")}
    Terminal=false
    Type=Application
    Icon=${iconPath}
    StartupWMClass=@lwn-simulator/desktop
    Categories=Utility;
  `;

  fs.writeFileSync(desktopPath, desktopContent.trim());

  fs.chmodSync(desktopPath, 0o755);

  exec("update-desktop-database ~/.local/share/applications", (error) => {
    if (error) {
      console.log("update-desktop-database unavailable:", error.message);
    }
  });
}

function getLauncherPath() {
  if (app.isPackaged) {
    return path.join(process.resourcesPath, "launcher", "out", "index.html");
  }

  return path.join(process.cwd(), "..", "launcher", "out", "index.html");
}

function createWindow() {
  mainWindow = new BrowserWindow({
    width: 1280,
    height: 800,
    webPreferences: {
      preload: path.join(__dirname, "preload.js"),
    },
  });

  mainWindow.loadFile(getLauncherPath());
}

ipcMain.handle("connect-local", async () => {
  if (!mainWindow) {
    return;
  }

  startBackend();

  await waitForServer();

  mainWindow.loadURL("http://localhost:8080");
});

ipcMain.handle("connect-remote", async (_, url: string) => {
  if (!mainWindow) {
    return;
  }

  mainWindow.loadURL(url);
});

app.whenReady().then(() => {
  if (app.isPackaged) {
    installLinuxDesktopEntry();
  }

  createWindow();
});

app.on("window-all-closed", () => {
  if (backendProcess) {
    backendProcess.kill();
  }

  if (process.platform !== "darwin") {
    app.quit();
  }
});
