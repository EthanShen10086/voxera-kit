import { describe, expect, it } from "vitest";

import { interpolate } from "./interpolation.js";

describe("interpolate", () => {
  it("replaces placeholders with params", () => {
    expect(interpolate("Hello {name}", { name: "world" })).toBe("Hello world");
  });

  it("keeps unknown placeholders", () => {
    expect(interpolate("Hello {missing}", {})).toBe("Hello {missing}");
  });
});
