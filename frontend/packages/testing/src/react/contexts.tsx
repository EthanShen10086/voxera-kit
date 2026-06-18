import type { Container } from "@voxera-kit/di";
import type { I18nEngine } from "@voxera-kit/i18n";
import type { ThemeEngine } from "@voxera-kit/theme";
import { createContext, useContext } from "react";

export const ThemeTestContext = createContext<ThemeEngine | null>(null);
export const I18nTestContext = createContext<I18nEngine | null>(null);
export const ContainerTestContext = createContext<Container | null>(null);

/** Returns the ThemeEngine from renderWithProviders. */
export function useTestTheme(): ThemeEngine {
  const engine = useContext(ThemeTestContext);
  if (!engine) {
    throw new Error("useTestTheme must be used within renderWithProviders");
  }
  return engine;
}

/** Returns the I18nEngine from renderWithProviders. */
export function useTestI18n(): I18nEngine {
  const engine = useContext(I18nTestContext);
  if (!engine) {
    throw new Error("useTestI18n must be used within renderWithProviders");
  }
  return engine;
}

/** Returns the DI Container from renderWithProviders. */
export function useTestContainer(): Container {
  const container = useContext(ContainerTestContext);
  if (!container) {
    throw new Error("useTestContainer must be used within renderWithProviders");
  }
  return container;
}
