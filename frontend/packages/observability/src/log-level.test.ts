import { describe, expect, it } from "vitest";

import { LogLevel } from "./types.js";

describe("LogLevel", () => {
  it("defines severity ordering", () => {
    expect(LogLevel.Error).toBeGreaterThan(LogLevel.Info);
  });
});
