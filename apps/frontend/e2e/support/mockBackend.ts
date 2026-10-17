import type { Page } from "@playwright/test";

export type MockBackendOptions = {
  logs?: "success" | "empty" | "error";
  devices?: "empty" | "active";
};

const profile = { id: "profile-1", name: "Default profile" };
const secondProfile = { id: "profile-2", name: "Test profile" };

const activeDevice = {
  id: "device-1",
  active: true,
  name: "Catania Weather Station",
  devEUI: "70B3D57ED0000001",
  class: "class-a",
  activation: "oota",
  OOTAConfig: {
    joinEUI: "0000000000000001",
    appKey: "00000000000000000000000000000001",
  },
  locationConfig: {
    latitude: 37.5079,
    longitude: 15.083,
    altitude: 15,
    region: "EU868",
  },
  RX1Config: { delay: 1, duration: 1, dataRateOffset: 0 },
  RX2Config: {
    delay: 2,
    duration: 1,
    channelFrequency: 869525000,
    dataRate: 0,
    ACKTimeout: 2,
  },
  advancedConfig: { antennaRange: 5000, ADREnabled: true },
  frameConfig: {
    fPort: 1,
    retransmission: 1,
    disableFrameCounterValidation: false,
  },
  payloadConfig: {
    uplinkInterval: 60000,
    oversizedPayloadBehavior: "truncate",
    MType: "unconfirmed-data-up",
    payload: "AQID",
    base64Encoded: true,
  },
};

const activeGateway = {
  id: "gateway-1",
  active: true,
  name: "Catania Lab Gateway",
  type: "virtual",
  macAddress: "02:00:00:00:00:01",
  gatewayEUI: "0102030405060708",
  keepAlive: 30,
  latitude: 37.5079,
  longitude: 15.083,
  altitude: 20,
};

const runs = [
  {
    summary: {
      id: "run-new",
      startedAtMilliseconds: 1_800_000_000_000,
      endedAtMilliseconds: 1_800_000_001_000,
      durationMilliseconds: 1_000,
      status: "stopped",
      speed: 1,
      seed: 42,
      deviceCount: 2,
      gatewayCount: 1,
      eventCount: 12,
      packetSuccessRate: 0.75,
    },
    objects: [],
  },
  {
    summary: {
      id: "run-old",
      startedAtMilliseconds: 1_700_000_000_000,
      endedAtMilliseconds: 1_700_000_001_000,
      durationMilliseconds: 1_000,
      status: "failed",
      speed: 1,
      seed: 7,
      deviceCount: 1,
      gatewayCount: 1,
      eventCount: 3,
      packetSuccessRate: 0,
    },
    objects: [],
  },
];

const json = (data: unknown) => ({
  status: 200,
  contentType: "application/json",
  headers: { "access-control-allow-origin": "*" },
  body: JSON.stringify(data),
});

export const mockBackend = async (
  page: Page,
  options: MockBackendOptions = {},
): Promise<void> => {
  let gatewayBridge = {
    enabled: false,
    address: "0.0.0.0",
    port: 1700,
  };
  let simulationStatus = "idle";

  const simulationSnapshot = () => ({
    state: {
      status: simulationStatus,
      speed: 1,
      elapsedMilliseconds: simulationStatus === "idle" ? 0 : 1000,
      eventSequence: simulationStatus === "idle" ? 0 : 1,
      deviceCount: options.devices === "active" ? 1 : 0,
      gatewayCount: options.devices === "active" ? 1 : 0,
    },
    devices: [],
    gateways: [],
    metrics: {
      totalUplinks: 0,
      successfulUplinks: 0,
      droppedUplinks: 0,
      totalPacketsReceived: 0,
      packetSuccessRate: 0,
      activeDevices: options.devices === "active" ? 1 : 0,
      activeGateways: options.devices === "active" ? 1 : 0,
    },
  });

  await page.route("**/*", async (route) => {
    const url = new URL(route.request().url());
    const isBackendRequest =
      url.port === "8080" ||
      (url.port === "3000" && url.pathname.startsWith("/api/"));
    if (!isBackendRequest) {
      await route.continue();
      return;
    }

    const path = url.pathname.replace(/^\/api/, "");

    if (path.endsWith("/profile/get-profiles")) {
      await route.fulfill(json({ profiles: [profile, secondProfile], total: 2 }));
      return;
    }

    if (path.endsWith("/simulation/activity")) {
      await route.fulfill(json({ active: false, status: "idle" }));
      return;
    }

    if (path.endsWith("/device/get-devices")) {
      await route.fulfill(
        json({ devices: options.devices === "active" ? [activeDevice] : [] }),
      );
      return;
    }

    if (path.endsWith("/gateway/get-gateways")) {
      await route.fulfill(
        json({ gateways: options.devices === "active" ? [activeGateway] : [] }),
      );
      return;
    }

    if (path.endsWith("/configuration/gateway-bridge")) {
      if (route.request().method() !== "GET") {
        try {
          gatewayBridge = {
            ...gatewayBridge,
            ...(route.request().postDataJSON() as Partial<typeof gatewayBridge>),
          };
        } catch {
          // Keep the previous bridge state when the request has no JSON body.
        }
      }

      await route.fulfill(json(gatewayBridge));
      return;
    }

    if (path.endsWith("/simulation/logs")) {
      if (options.logs === "error") {
        await route.fulfill({
          status: 503,
          contentType: "application/json",
          headers: { "access-control-allow-origin": "*" },
          body: JSON.stringify({ error: "Backend unavailable" }),
        });
        return;
      }

      if (options.logs === "empty") {
        await route.fulfill(json({ runs: [], total: 0 }));
        return;
      }

      await route.fulfill(json({ runs, total: runs.length }));
      return;
    }

    if (path.endsWith("/simulation/snapshot")) {
      await route.fulfill(json(simulationSnapshot()));
      return;
    }

    if (path.endsWith("/simulation/start")) {
      simulationStatus = "running";
      await route.fulfill(json(simulationSnapshot()));
      return;
    }

    if (path.endsWith("/simulation/pause")) {
      simulationStatus = "paused";
      await route.fulfill(json(simulationSnapshot()));
      return;
    }

    if (path.endsWith("/simulation/resume")) {
      simulationStatus = "running";
      await route.fulfill(json(simulationSnapshot()));
      return;
    }

    if (path.endsWith("/simulation/stop")) {
      simulationStatus = "stopped";
      await route.fulfill(json(simulationSnapshot()));
      return;
    }

    if (path.endsWith("/simulation/events")) {
      await route.fulfill(json({ events: [], lastSequence: 0 }));
      return;
    }

    await route.fulfill(json({}));
  });
};
