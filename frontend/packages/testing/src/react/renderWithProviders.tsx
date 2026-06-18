import { cleanup, render, type RenderOptions, type RenderResult } from "@testing-library/react";
import type { ReactElement } from "react";
import { afterEach } from "vitest";

import { TestProviders, type TestProviderOptions } from "./providers.js";

/** Registers RTL cleanup after each Vitest test. Import from a Vitest setup file. */
export function registerReactTestingCleanup(): void {
  afterEach(() => {
    cleanup();
  });
}

/** Renders UI wrapped with kit theme/i18n/DI providers. */
export function renderWithProviders(
  ui: ReactElement,
  options?: TestProviderOptions & { renderOptions?: Omit<RenderOptions, "wrapper"> },
): RenderResult {
  const { renderOptions, ...providerOptions } = options ?? {};
  return render(ui, {
    ...renderOptions,
    wrapper: ({ children }) => (
      <TestProviders options={providerOptions}>{children}</TestProviders>
    ),
  });
}

export {
  ContainerTestContext,
  I18nTestContext,
  ThemeTestContext,
  useTestContainer,
  useTestI18n,
  useTestTheme,
} from "./contexts.js";
export { TestProviders, createDefaultTestProviders, type TestProviderOptions } from "./providers.js";
