import { expect, test } from "@playwright/test";

test("the application root reaches the simulation dashboard", async ({ page }) => {
  await page.goto("/");

  await expect(page).toHaveURL(/\/simulation\/dashboard\/?$/);
});
