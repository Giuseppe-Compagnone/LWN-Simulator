import { AppInfoResponse } from "@lwn-simulator/contracts";

export async function validateRemote(url: string) {
  const endpoint = new URL("/api/info", url);

  let response: Response;

  try {
    response = await fetch(endpoint);
  } catch (err) {
    throw new Error("Server unreachable", {
      cause: err,
    });
  }

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
