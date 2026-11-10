import { expect, test } from "@playwright/test";
import { readFile } from "node:fs/promises";
import { mockBackend } from "./support/mockBackend";

test("downloads all logs as a ZIP archive", async ({ page }) => {
  await mockBackend(page);
  await page.goto("/logs");
  const downloadPromise = page.waitForEvent("download");
  await page.getByRole("button", { name: "Export all (ZIP)" }).click();
  const download = await downloadPromise;
  expect(download.suggestedFilename()).toBe("simulation-logs.zip");
  const path = await download.path();
  expect(await readFile(path!, "utf8")).toBe("PK-test-archive");
  await expect(page).toHaveURL(/\/logs\/?$/);
});

test("downloads the selected run as JSON", async ({ page }) => {
  await mockBackend(page);
  await page.goto("/logs/detail?runId=run-new");
  const downloadPromise = page.waitForEvent("download");
  await page.getByRole("button", { name: "Export JSON" }).click();
  const download = await downloadPromise;
  expect(download.suggestedFilename()).toBe("simulation-run-new.json");
  const path = await download.path();
  const exported = JSON.parse(await readFile(path!, "utf8"));
  expect(exported.run.summary.id).toBe("run-new");
  expect(exported.events[0].message).toBe("Exported uplink");
  await expect(page).toHaveURL(/runId=run-new/);
});

test("disables the archive button when there are no logs", async ({ page }) => {
  await mockBackend(page, { logs: "empty" });
  await page.goto("/logs");
  await expect(page.getByRole("button", { name: "Export all (ZIP)" })).toBeDisabled();
});
