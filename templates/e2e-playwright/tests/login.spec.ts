import { expect, test } from "@playwright/test";

import { LoginPage } from "../pages/LoginPage.js";

test.describe("login flow", () => {
  test("redirects to dashboard after valid credentials", async ({ page }) => {
    const login = new LoginPage(page);
    await login.goto();
    await login.login("demo@voxera.test", "demo-pass");

    await expect(page).toHaveURL(/\/dashboard(\.html)?$/);
    await expect(page.getByTestId("welcome")).toBeVisible();
    await expect(page.getByTestId("user-email")).toHaveText("demo@voxera.test");
  });

  test("shows error for invalid credentials", async ({ page }) => {
    const login = new LoginPage(page);
    await login.goto();
    await login.login("wrong@example.com", "bad");

    await expect(page.getByTestId("error")).toBeVisible();
    await expect(page.getByTestId("error")).toHaveText("Invalid email or password");
  });

  test("logout returns to login", async ({ page }) => {
    const login = new LoginPage(page);
    await login.goto();
    await login.login("demo@voxera.test", "demo-pass");
    await page.getByTestId("logout").click();

    await expect(page).toHaveURL(/\/login(\.html)?$/);
  });
});
