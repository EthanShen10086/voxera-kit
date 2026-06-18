import type { Page } from "@playwright/test";

export class LoginPage {
  constructor(private readonly page: Page) {}

  async goto() {
    await this.page.goto("/login.html");
  }

  async login(email: string, password: string) {
    await this.page.getByTestId("email").fill(email);
    await this.page.getByTestId("password").fill(password);
    await this.page.getByTestId("submit").click();
  }
}
