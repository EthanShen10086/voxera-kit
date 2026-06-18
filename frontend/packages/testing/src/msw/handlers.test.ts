import { afterAll, afterEach, beforeAll, describe, expect, it } from "vitest";

import { API_BASE, createMockServer } from "./server.js";

const server = createMockServer();

describe("mswHandlers", () => {
  beforeAll(() => server.listen());
  afterEach(() => server.resetHandlers());
  afterAll(() => server.close());

  it("handles auth login", async () => {
    const response = await fetch(`${API_BASE}/auth/login`, { method: "POST" });
    expect(response.status).toBe(200);
    const body = (await response.json()) as { data: { user: { id: string } } };
    expect(body.data.user.id).toMatch(/^user-/);
  });

  it("handles paginated items", async () => {
    const response = await fetch(`${API_BASE}/items?page=2&pageSize=5`);
    expect(response.status).toBe(200);
    const body = (await response.json()) as { page: number; pageSize: number; data: unknown[] };
    expect(body.page).toBe(2);
    expect(body.pageSize).toBe(5);
    expect(body.data.length).toBeGreaterThan(0);
  });

  it("handles error codes", async () => {
    const response = await fetch(`${API_BASE}/errors/unauthorized`);
    expect(response.status).toBe(401);
    const body = (await response.json()) as { code: string };
    expect(body.code).toBe("UNAUTHORIZED");
  });
});
