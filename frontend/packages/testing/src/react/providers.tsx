import { Container } from "@voxera-kit/di";
import { I18nEngine } from "@voxera-kit/i18n";
import { lightPreset, ThemeEngine } from "@voxera-kit/theme";
import type { ReactNode } from "react";

import {
  ContainerTestContext,
  I18nTestContext,
  ThemeTestContext,
} from "./contexts.js";

/** Options for kit test providers. */
export interface TestProviderOptions {
  themeEngine?: ThemeEngine;
  i18nEngine?: I18nEngine;
  container?: Container;
  /** Applies the light theme preset on mount when true (default). */
  applyLightTheme?: boolean;
}

/** Creates default theme/i18n/container instances for tests. */
export function createDefaultTestProviders(options: { applyLightTheme?: boolean } = {}) {
  const applyLightTheme = options.applyLightTheme ?? true;
  const themeEngine = new ThemeEngine();
  if (applyLightTheme) {
    themeEngine.apply(lightPreset);
  }
  const i18nEngine = new I18nEngine("en");
  i18nEngine.loadLocale("en", {
    common: {
      hello: "Hello",
    },
  });
  return {
    themeEngine,
    i18nEngine,
    container: new Container(),
  };
}

/** Wraps children with optional theme, i18n, and DI providers for component tests. */
export function TestProviders({
  children,
  options = {},
}: {
  children: ReactNode;
  options?: TestProviderOptions;
}) {
  const applyLightTheme = options.applyLightTheme ?? true;
  const defaults = createDefaultTestProviders({ applyLightTheme });
  const themeEngine = options.themeEngine ?? defaults.themeEngine;
  const i18nEngine = options.i18nEngine ?? defaults.i18nEngine;
  const container = options.container ?? defaults.container;

  return (
    <ThemeTestContext.Provider value={themeEngine}>
      <I18nTestContext.Provider value={i18nEngine}>
        <ContainerTestContext.Provider value={container}>{children}</ContainerTestContext.Provider>
      </I18nTestContext.Provider>
    </ThemeTestContext.Provider>
  );
}
