import type { FixtureSession } from "./types.js";

let sessionSeq = 0;

/** Resets the session fixture counter for deterministic tests. */
export function resetSessionFixtures(): void {
  sessionSeq = 0;
}

/** Creates a session JSON object with optional overrides. */
export function createSession(overrides: Partial<FixtureSession> = {}): FixtureSession {
  sessionSeq += 1;
  return {
    token: `session-token-${sessionSeq}`,
    userId: overrides.userId ?? `user-${sessionSeq}`,
    expiresAt: "2026-12-31T23:59:59Z",
    ...overrides,
  };
}
