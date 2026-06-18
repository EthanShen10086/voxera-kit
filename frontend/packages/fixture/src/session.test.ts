import { describe, expect, it } from "vitest";

import { createSession, resetSessionFixtures } from "./session.js";

describe("createSession", () => {
  it("returns deterministic sessions after reset", () => {
    resetSessionFixtures();
    expect(createSession()).toEqual({
      token: "session-token-1",
      userId: "user-1",
      expiresAt: "2026-12-31T23:59:59Z",
    });
  });
});
