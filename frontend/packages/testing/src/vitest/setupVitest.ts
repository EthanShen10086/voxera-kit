import type { UserConfig } from "vitest/config";

/** Default coverage thresholds aligned with kit quality goals. */
export const defaultCoverageThresholds = {
  lines: 80,
  functions: 80,
  branches: 70,
  statements: 80,
} as const;

const baseCoverage = {
  provider: "v8" as const,
  reporter: ["text", "json-summary"],
  thresholds: { ...defaultCoverageThresholds },
};

/** Creates a baseline Vitest config fragment for kit packages and products. */
export function createVitestConfig(overrides: UserConfig = {}): UserConfig {
  return mergeVitestConfig(
    {
      test: {
        globals: false,
        environment: "node",
        coverage: baseCoverage,
      },
    },
    overrides,
  );
}

type CoverageConfig = {
  thresholds?: Record<string, number>;
  [key: string]: unknown;
};

function mergeCoverage(base: CoverageConfig | undefined, overrides: CoverageConfig | undefined) {
  return {
    ...(base ?? {}),
    ...(overrides ?? {}),
    thresholds: {
      ...(base?.thresholds ?? {}),
      ...(overrides?.thresholds ?? {}),
    },
  };
}

/** Deep-merges Vitest configs while preserving nested `test.coverage`. */
export function mergeVitestConfig(base: UserConfig, overrides: UserConfig): UserConfig {
  const baseTest = base.test ?? {};
  const overrideTest = overrides.test ?? {};

  return {
    ...base,
    ...overrides,
    test: {
      ...baseTest,
      ...overrideTest,
      coverage: mergeCoverage(
        baseTest.coverage as CoverageConfig | undefined,
        overrideTest.coverage as CoverageConfig | undefined,
      ),
    },
  };
}

/** Alias for createVitestConfig to match Wave T naming. */
export const setupVitest = createVitestConfig;
