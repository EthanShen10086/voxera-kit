import type { FixtureUser } from "./types.js";

let userSeq = 0;

/** Resets the user fixture counter for deterministic tests. */
export function resetUserFixtures(): void {
  userSeq = 0;
}

/** Creates a user JSON object with optional overrides. */
export function createUser(overrides: Partial<FixtureUser> = {}): FixtureUser {
  userSeq += 1;
  return {
    id: `user-${userSeq}`,
    email: `user${userSeq}@example.com`,
    name: `Test User ${userSeq}`,
    role: "member",
    ...overrides,
  };
}
