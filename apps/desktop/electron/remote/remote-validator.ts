import path from "path";
import { AppInfoResponse } from "@lwn-simulator/contracts";

export async function validateRemote(url: string) {
  const response = await fetch(path.join(url, "/api/info"));

  if (!response.ok) {
    throw new Error("Server unreachable");
  }

  const info: AppInfoResponse = await response.json();

  if (info.app !== "lwn-simulator") {
    throw new Error("Server isn't a LWN Simulator instance");
  }

  const version = process.env.APP_ENV || "dev";

  if (version !== "dev" && info.version !== version) {
    throw new Error("Incompatible version");
  }

  return info;
}
