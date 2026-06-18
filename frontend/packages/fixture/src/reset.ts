import { resetApiFixtures } from "./api.js";
import { resetSessionFixtures } from "./session.js";
import { resetUserFixtures } from "./user.js";

/** Resets all fixture counters. */
export function resetFixtures(): void {
  resetUserFixtures();
  resetSessionFixtures();
  resetApiFixtures();
}
