import { describe, expect, it } from "vitest";
import { createFaker } from "./factory.js";
import { createLightweightFaker } from "./lightweight.js";

describe("createFaker", () => {
  it("lightweight is deterministic with seed", () => {
    const a = createLightweightFaker(7);
    const b = createLightweightFaker(7);
    expect(a.email()).toBe(b.email());
  });

  it("factory supports lightweight provider", () => {
    const f = createFaker({ provider: "lightweight", seed: 1 });
    expect(f.uuid()).toMatch(/^[0-9a-f-]{36}$/);
  });

  it("factory supports fakerjs provider", () => {
    const f = createFaker({ provider: "fakerjs", seed: 99 });
    expect(f.sentence(3).length).toBeGreaterThan(0);
  });
});
