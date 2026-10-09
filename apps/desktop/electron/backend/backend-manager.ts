import { app } from "electron";
import { ChildProcess, spawn } from "child_process";
import path from "path";
import { findAvailablePort } from "./port-utils";

let backendProcess: ChildProcess | undefined;

export async function startBackend() {
  const isDevelopment = !app.isPackaged;
  const backendPath = app.isPackaged
    ? path.join(process.resourcesPath, "lwn-server")
    : path.join(app.getAppPath(), "assets", "lwn-server");

  const port = await findAvailablePort();

  backendProcess = spawn(backendPath, ["-p", String(port)], {
    cwd: isDevelopment
      ? path.resolve(app.getAppPath(), "..", "backend")
      : undefined,
    env: {
      ...process.env,
      ...(isDevelopment ? { LWN_ENV: "development" } : {}),
    },
    stdio: "inherit",
  });

  console.log(`Backend running on port: ${port}`);

  return port;
}

export function stopBackend() {
  backendProcess?.kill();
  backendProcess = undefined;
}

export async function waitForServer(
  port: number,
  timeoutMs = 10000,
): Promise<void> {
  const deadline = Date.now() + timeoutMs;

  while (Date.now() < deadline) {
    try {
      const response = await fetch(
        `http://127.0.0.1:${port}/api/app-info/status`,
      );

      if (response.ok) {
        return;
      }
    } catch {
      // The backend may still be starting.
    }

    await new Promise((resolve) => setTimeout(resolve, 150));
  }

  throw new Error("Local backend did not become ready in time");
}
