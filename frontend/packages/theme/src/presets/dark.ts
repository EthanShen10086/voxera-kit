import type { ThemePreset } from "../types.js";
import { defaultTokens } from "../tokens.js";

/** Dark theme preset with inverted luminance and softened accent colours. */
export const darkPreset: ThemePreset = {
  name: "dark",
  tokens: {
    ...defaultTokens,
    colors: {
      primary: {
        main: "#818cf8",
        light: "#a5b4fc",
        dark: "#6366f1",
      },
      secondary: {
        main: "#a78bfa",
        light: "#c4b5fd",
        dark: "#8b5cf6",
      },
      background: {
        main: "#0f172a",
        light: "#1e293b",
        dark: "#020617",
      },
      surface: {
        main: "#1e293b",
        light: "#334155",
        dark: "#0f172a",
      },
      text: {
        main: "#f1f5f9",
        light: "#94a3b8",
        dark: "#f8fafc",
      },
      error: {
        main: "#f87171",
        light: "#fca5a5",
        dark: "#ef4444",
      },
      warning: {
        main: "#fbbf24",
        light: "#fde68a",
        dark: "#f59e0b",
      },
      success: {
        main: "#34d399",
        light: "#6ee7b7",
        dark: "#10b981",
      },
      info: {
        main: "#60a5fa",
        light: "#93c5fd",
        dark: "#3b82f6",
      },
    },
    shadows: {
      none: "none",
      sm: "0 1px 2px 0 rgb(0 0 0 / 0.3)",
      md: "0 4px 6px -1px rgb(0 0 0 / 0.4), 0 2px 4px -2px rgb(0 0 0 / 0.3)",
      lg: "0 10px 15px -3px rgb(0 0 0 / 0.4), 0 4px 6px -4px rgb(0 0 0 / 0.3)",
      xl: "0 20px 25px -5px rgb(0 0 0 / 0.5), 0 8px 10px -6px rgb(0 0 0 / 0.4)",
    },
  },
};
