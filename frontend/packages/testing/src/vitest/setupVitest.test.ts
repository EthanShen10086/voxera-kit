import { describe, expect, it } from "vitest";

import { createVitestConfig, defaultCoverageThresholds, mergeVitestConfig } from "./setupVitest.js";

describe("setupVitest", () => {
  it("applies default coverage thresholds", () => {
    const config = createVitestConfig();
    const coverage = config.test?.coverage as { thresholds?: typeof defaultCoverageThresholds };
    expect(coverage?.thresholds).toEqual(defaultCoverageThresholds);
  });

  it("merges overrides", () => {
    const merged = mergeVitestConfig(createVitestConfig(), {
      test: {
        environment: "jsdom",
        coverage: {
          thresholds: {
            lines: 90,
          },
        },
      },
    });
    const coverage = merged.test?.coverage as { thresholds?: Record<string, number> };
    expect(merged.test?.environment).toBe("jsdom");
    expect(coverage?.thresholds?.lines).toBe(90);
    expect(coverage?.thresholds?.functions).toBe(80);
  });
});
