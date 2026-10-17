import { expect, test } from "@playwright/test";
import { mockBackend } from "./support/mockBackend";

test("creates a virtual gateway from the hardware form", async ({ page }) => {
  await mockBackend(page);
  await page.goto("/hardware/gateways/new");

  await page.getByLabel("Name").fill("Catania Lab Gateway");
  await page.getByText("VIRTUAL", { exact: true }).click();
  await page.getByLabel("MAC Address").fill("02:00:00:00:00:01");
  await page.getByLabel("Gateway EUI").fill("0102030405060708");
  await page.getByLabel("Keep Alive (seconds)").fill("30");
  await page.getByLabel("Latitude").fill("37.5079");
  await page.getByLabel("Longitude").fill("15.083");
  await page.getByLabel("Altitude").fill("20");
  await page.getByRole("button", { name: "Create" }).click();

  await expect(page).toHaveURL(/\/hardware\/gateways\/?$/);
});

test("keeps invalid device identifiers from being submitted", async ({ page }) => {
  await mockBackend(page);
  await page.goto("/hardware/devices/new");

  await page.getByLabel("DevEUI").fill("1234");
  await page.getByRole("button", { name: "Create" }).click();

  await expect(page.getByText("Invalid DevEUI", { exact: true })).toBeVisible();
  await expect(page).toHaveURL(/\/hardware\/devices\/new\/?$/);
});
