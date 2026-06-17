import { describe, expect, it } from "vitest";

import type { ExperimentVariant } from "./types.js";

describe("experiment types", () => {
  it("accepts variant metadata", () => {
    const variant: ExperimentVariant = {
      key: "control",
      weight: 50,
      payload: { color: "blue" },
    };
    expect(variant.key).toBe("control");
  });
});
