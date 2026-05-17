import type { ThemePreset } from "../types.js";
import { defaultTokens } from "../tokens.js";

/** High-contrast preset optimised for accessibility (WCAG AAA targets). */
export const highContrastPreset: ThemePreset = {
  name: "high-contrast",
  tokens: {
    ...defaultTokens,
    colors: {
      primary: {
        main: "#1d4ed8",
        light: "#3b82f6",
        dark: "#1e3a8a",
      },
      secondary: {
        main: "#7c3aed",
        light: "#8b5cf6",
        dark: "#5b21b6",
      },
      background: {
        main: "#ffffff",
        light: "#ffffff",
        dark: "#f3f4f6",
      },
      surface: {
        main: "#ffffff",
        light: "#ffffff",
        dark: "#e5e7eb",
      },
      text: {
        main: "#000000",
        light: "#1f2937",
        dark: "#000000",
      },
      error: {
        main: "#b91c1c",
        light: "#dc2626",
        dark: "#991b1b",
      },
      warning: {
        main: "#92400e",
        light: "#b45309",
        dark: "#78350f",
      },
      success: {
        main: "#047857",
        light: "#059669",
        dark: "#065f46",
      },
      info: {
        main: "#1d4ed8",
        light: "#2563eb",
        dark: "#1e3a8a",
      },
    },
    shadows: {
      none: "none",
      sm: "0 1px 2px 0 rgb(0 0 0 / 0.15)",
      md: "0 4px 6px -1px rgb(0 0 0 / 0.2), 0 2px 4px -2px rgb(0 0 0 / 0.15)",
      lg: "0 10px 15px -3px rgb(0 0 0 / 0.2), 0 4px 6px -4px rgb(0 0 0 / 0.15)",
      xl: "0 20px 25px -5px rgb(0 0 0 / 0.25), 0 8px 10px -6px rgb(0 0 0 / 0.2)",
    },
  },
};
