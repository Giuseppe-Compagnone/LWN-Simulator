const { app, BrowserWindow } = require("electron");
const { spawn } = require("child_process");
const path = require("path");
const fs = require("fs");
const os = require("os");
const { exec } = require("child_process");

let backendProcess;

function startBackend() {
  const backendPath = path.join(process.resourcesPath, "lwn-server");

  backendProcess = spawn(backendPath, [], {
    stdio: "inherit",
  });
}

function waitForServer() {
  return new Promise((resolve) => {
    setTimeout(() => {
      resolve();
    }, 1500);
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

  fs.mkdirSync(applicationsDir, { recursive: true });
  fs.mkdirSync(iconsDir, { recursive: true });

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
    StartupWMClass=lwn-simulator
    Categories=Utility;
    StartupWMClass=desktop
  `;

  fs.writeFileSync(desktopPath, desktopContent.trim());

  fs.chmodSync(desktopPath, 0o755);

  exec("update-desktop-database ~/.local/share/applications", (error) => {
    if (error) {
      console.log("update-desktop-database unavailable:", error.message);
    }
  });
}

function createWindow() {
  const win = new BrowserWindow({
    width: 1280,
    height: 800,
  });

  win.loadURL("http://localhost:8080");
}

app.whenReady().then(async () => {
  installLinuxDesktopEntry();

  startBackend();

  await waitForServer();

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
