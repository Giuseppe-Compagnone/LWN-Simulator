import { ApiCaller, AppInfoService } from "@lwn-simulator/sdk";

export async function validateRemote(url: string) {
  ApiCaller.baseUrl = new URL("/api", url).toString();

  try {
    const info = await AppInfoService.instance.appInfo();

    if (info.app !== "lwn-simulator") {
      throw new Error("Server isn't a LWN Simulator instance");
    }

    const version = process.env.APP_ENV || "dev";

    if (version !== "dev" && info.version !== version) {
      throw new Error("Incompatible version");
    }

    return info;
  } catch (err) {
    throw new Error("Server unreachable", {
      cause: err,
    });
  }
}
