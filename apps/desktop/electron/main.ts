import { app, BrowserWindow, ipcMain } from "electron";
import { ChildProcess, spawn } from "child_process";
import path from "path";
import fs from "fs";
import os from "os";
import { exec } from "child_process";
import { AppInfoResponse } from "@lwn-simulator/contracts";
import net from "net";

let backendProcess: ChildProcess | undefined;
let mainWindow: BrowserWindow | undefined;

function isPortAvailable(port: number) {
  return new Promise((resolve) => {
    const server = net.createServer();

    server.once("error", () => {
      resolve(false);
    });

    server.once("listening", () => {
      server.close(() => resolve(true));
    });

    server.listen(port, "127.0.0.1");
  });
}

async function findAvailablePort(startPort = 8080) {
  let port = startPort;

  while (!(await isPortAvailable(port))) {
    port++;
  }

  return port;
}

async function startBackend() {
  const backendPath = app.isPackaged
    ? path.join(process.resourcesPath, "lwn-server")
    : path.join(process.cwd(), "assets", "lwn-server");

  const port = await findAvailablePort(8080);

  backendProcess = spawn(backendPath, ["-p", String(port)], {
    stdio: "inherit",
  });

  console.log(`Backend avviato sulla porta ${port}`);
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

async function validateRemote(url: string) {
  const response = await fetch(path.join(url, "/api/info"));

  if (!response.ok) {
    throw new Error("Server non raggiungibile");
  }

  const info: AppInfoResponse = await response.json();

  if (info.app !== "lwn-simulator") {
    throw new Error("Il server non è un'istanza LWN Simulator");
  }

  const version = process.env.APP_ENV || "dev";

  if (version !== "dev" && info.version !== version) {
    throw new Error("Versione API non compatibile");
  }

  return info;
}

ipcMain.handle("connect-remote", async (_, url: string) => {
  if (!mainWindow) {
    return;
  }

  try {
    await validateRemote(url);

    await mainWindow.loadURL(url);

    return {
      success: true,
    };
  } catch (error) {
    return {
      success: false,
      message: "Connection failed",
    };
  }
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
