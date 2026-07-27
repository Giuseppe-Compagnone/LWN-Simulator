const { app, BrowserWindow } = require("electron");
const { spawn } = require("child_process");
const path = require("path");

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

function createWindow() {
  const win = new BrowserWindow({
    width: 1280,
    height: 800,
  });

  win.loadURL("http://localhost:8080");
}

app.whenReady().then(async () => {
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
