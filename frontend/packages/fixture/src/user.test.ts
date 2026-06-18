import { describe, expect, it } from "vitest";

import { createUser, resetUserFixtures } from "./user.js";

describe("createUser", () => {
  it("returns deterministic users after reset", () => {
    resetUserFixtures();
    expect(createUser()).toEqual({
      id: "user-1",
      email: "user1@example.com",
      name: "Test User 1",
      role: "member",
    });
    expect(createUser({ role: "admin" }).role).toBe("admin");
  });
});
