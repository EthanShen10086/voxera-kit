import { describe, expect, it } from "vitest";

import { PlayerEvent } from "./events.js";

describe("PlayerEvent", () => {
  it("includes play and pause events", () => {
    expect(PlayerEvent.Play).toBe("play");
    expect(PlayerEvent.Pause).toBe("pause");
  });
});
