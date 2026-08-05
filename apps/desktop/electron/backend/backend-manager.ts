import { app } from "electron";
import { ChildProcess, spawn } from "child_process";
import path from "path";
import { findAvailablePort } from "./port-utils";

let backendProcess: ChildProcess | undefined;

export async function startBackend() {
  const backendPath = app.isPackaged
    ? path.join(process.resourcesPath, "lwn-server")
    : path.join(process.cwd(), "assets", "lwn-server");

  const port = await findAvailablePort();

  backendProcess = spawn(backendPath, ["-p", String(port)], {
    stdio: "inherit",
  });

  console.log(`Backend running on port: ${port}`);

  return port;
}

export function stopBackend() {
  backendProcess?.kill();
}

export function waitForServer() {
  return new Promise<void>((resolve) => {
    setTimeout(resolve, 1500);
  });
}
