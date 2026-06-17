import { describe, expect, it } from "vitest";

import { lightPreset } from "./presets/light.js";
import { defaultTokens } from "./tokens.js";

describe("theme tokens", () => {
  it("defines a primary color scale", () => {
    expect(defaultTokens.colors.primary.main).toBe("#6366f1");
  });

  it("ships a light preset backed by default tokens", () => {
    expect(lightPreset.name).toBe("light");
    expect(lightPreset.tokens).toEqual(defaultTokens);
  });
});
