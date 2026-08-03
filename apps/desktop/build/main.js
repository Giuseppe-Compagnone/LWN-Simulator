"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
const electron_1 = require("electron");
const child_process_1 = require("child_process");
const path_1 = __importDefault(require("path"));
const fs_1 = __importDefault(require("fs"));
const os_1 = __importDefault(require("os"));
const child_process_2 = require("child_process");
let backendProcess;
function startBackend() {
    const backendPath = path_1.default.join(process.resourcesPath, "lwn-server");
    backendProcess = (0, child_process_1.spawn)(backendPath, [], {
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
    if (process.platform !== "linux")
        return;
    const applicationsDir = path_1.default.join(os_1.default.homedir(), ".local", "share", "applications");
    const iconsDir = path_1.default.join(os_1.default.homedir(), ".local", "share", "icons");
    fs_1.default.mkdirSync(applicationsDir, { recursive: true });
    fs_1.default.mkdirSync(iconsDir, { recursive: true });
    const desktopPath = path_1.default.join(applicationsDir, "lwn-simulator.desktop");
    const iconPath = path_1.default.join(iconsDir, "lwn-simulator.png");
    fs_1.default.copyFileSync(path_1.default.join(process.resourcesPath, "icon.png"), iconPath);
    const desktopContent = `
    [Desktop Entry]
    Name=LWN-Simulator
    Exec=${process.env.APPIMAGE || electron_1.app.getPath("exe")}
    Terminal=false
    Type=Application
    Icon=${iconPath}
    StartupWMClass=lwn-simulator
    Categories=Utility;
    StartupWMClass=desktop
  `;
    fs_1.default.writeFileSync(desktopPath, desktopContent.trim());
    fs_1.default.chmodSync(desktopPath, 0o755);
    (0, child_process_2.exec)("update-desktop-database ~/.local/share/applications", (error) => {
        if (error) {
            console.log("update-desktop-database unavailable:", error.message);
        }
    });
}
function createWindow() {
    const win = new electron_1.BrowserWindow({
        width: 1280,
        height: 800,
    });
    win.loadURL("http://localhost:8080");
}
electron_1.app.whenReady().then(async () => {
    installLinuxDesktopEntry();
    startBackend();
    await waitForServer();
    createWindow();
});
electron_1.app.on("window-all-closed", () => {
    if (backendProcess) {
        backendProcess.kill();
    }
    if (process.platform !== "darwin") {
        electron_1.app.quit();
    }
});
