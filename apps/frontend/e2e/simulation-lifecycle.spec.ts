import { expect, test } from "@playwright/test";
import { mockBackend } from "./support/mockBackend";

test("runs the complete simulation control lifecycle", async ({ page }) => {
  await mockBackend(page, { devices: "active" });
  await page.goto("/simulation/dashboard");

  const startButton = page.getByRole("button", { name: "Start simulation" });
  await expect(startButton).toBeEnabled({ timeout: 15_000 });

  await page.getByLabel("Seed (optional)").fill("42");
  await startButton.click();

  await expect(page.getByText("Running", { exact: true })).toBeVisible();
  await expect(page.getByRole("button", { name: "Pause" })).toBeVisible();

  await page.getByRole("button", { name: "Pause" }).click();
  await expect(page.getByText("Paused", { exact: true })).toBeVisible();
  await expect(page.getByRole("button", { name: "Resume" })).toBeVisible();

  await page.getByRole("button", { name: "Resume" }).click();
  await expect(page.getByText("Running", { exact: true })).toBeVisible();

  await page.getByRole("button", { name: "Stop" }).click();
  await expect(page.getByText("Stopped", { exact: true })).toBeVisible();
  await expect(startButton).toBeEnabled();
});
