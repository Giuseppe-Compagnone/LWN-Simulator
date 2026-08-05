import { app } from "electron";

import { exec } from "child_process";

import fs from "fs";
import os from "os";
import path from "path";

export function installLinuxDesktopEntry() {
  if (process.platform !== "linux") {
    return;
  }

  const applicationsDir = path.join(os.homedir(), ".local/share/applications");

  const iconsDir = path.join(os.homedir(), ".local/share/icons");

  fs.mkdirSync(applicationsDir, { recursive: true });

  fs.mkdirSync(iconsDir, { recursive: true });

  const iconPath = path.join(iconsDir, "lwn-simulator.png");

  fs.copyFileSync(path.join(process.resourcesPath, "icon.png"), iconPath);

  const desktopPath = path.join(applicationsDir, "lwn-simulator.desktop");

  fs.writeFileSync(
    desktopPath,
    `
[Desktop Entry]
Name=LWN-Simulator
Exec=${process.env.APPIMAGE || app.getPath("exe")}
Terminal=false
Type=Application
Icon=${iconPath}
Categories=Utility;
StartupWMClass=@lwn-simulator/desktop
`.trim(),
  );

  fs.chmodSync(desktopPath, 0o755);

  exec("update-desktop-database ~/.local/share/applications");
}
