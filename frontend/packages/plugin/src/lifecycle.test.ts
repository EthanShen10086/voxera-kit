import { describe, expect, it } from "vitest";

import { PluginLifecycle, validateTransition } from "./lifecycle.js";

describe("plugin lifecycle", () => {
  it("allows registered to initialized", () => {
    expect(validateTransition("registered", "initialized")).toBe(true);
  });

  it("rejects mounted to registered", () => {
    expect(validateTransition("mounted", "registered")).toBe(false);
  });

  it("exposes lifecycle enum values", () => {
    expect(PluginLifecycle.Mounted).toBe("mounted");
  });
});
