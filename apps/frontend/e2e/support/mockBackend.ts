import type { Page } from "@playwright/test";

export type MockBackendOptions = {
  logs?: "success" | "empty" | "error";
};

const profile = { id: "profile-1", name: "Default profile" };
const secondProfile = { id: "profile-2", name: "Test profile" };

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
      await route.fulfill(json({ devices: [] }));
      return;
    }

    if (path.endsWith("/gateway/get-gateways")) {
      await route.fulfill(json({ gateways: [] }));
      return;
    }

    if (path.endsWith("/configuration/gateway-bridge")) {
      await route.fulfill(
        json({ enabled: false, address: "0.0.0.0", port: 1700 }),
      );
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
      await route.fulfill(json({
        state: {
          status: "idle",
          speed: 1,
          elapsedMilliseconds: 0,
          eventSequence: 0,
          deviceCount: 0,
          gatewayCount: 0,
        },
        devices: [],
        gateways: [],
        metrics: {
          totalUplinks: 0,
          successfulUplinks: 0,
          droppedUplinks: 0,
          totalPacketsReceived: 0,
          packetSuccessRate: 0,
        },
      }));
      return;
    }

    if (path.endsWith("/simulation/events")) {
      await route.fulfill(json({ events: [], lastSequence: 0 }));
      return;
    }

    await route.fulfill(json({}));
  });
};
