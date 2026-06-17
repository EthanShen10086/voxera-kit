import { describe, expect, it } from "vitest";

import { SessionManager } from "./session.js";

describe("SessionManager", () => {
  it("returns a stable session id within timeout", () => {
    const session = new SessionManager(60_000);
    expect(session.getSessionId()).toBe(session.getSessionId());
  });
});
