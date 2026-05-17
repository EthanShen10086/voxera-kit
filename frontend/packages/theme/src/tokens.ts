import type { ThemeTokens } from "./types.js";

/** Default design tokens – a professional, neutral palette suitable as a starting point. */
export const defaultTokens: ThemeTokens = {
  colors: {
    primary: {
      main: "#6366f1",
      light: "#818cf8",
      dark: "#4f46e5",
    },
    secondary: {
      main: "#8b5cf6",
      light: "#a78bfa",
      dark: "#7c3aed",
    },
    background: {
      main: "#ffffff",
      light: "#f9fafb",
      dark: "#f3f4f6",
    },
    surface: {
      main: "#ffffff",
      light: "#f9fafb",
      dark: "#e5e7eb",
    },
    text: {
      main: "#111827",
      light: "#6b7280",
      dark: "#030712",
    },
    error: {
      main: "#ef4444",
      light: "#fca5a5",
      dark: "#dc2626",
    },
    warning: {
      main: "#f59e0b",
      light: "#fcd34d",
      dark: "#d97706",
    },
    success: {
      main: "#10b981",
      light: "#6ee7b7",
      dark: "#059669",
    },
    info: {
      main: "#3b82f6",
      light: "#93c5fd",
      dark: "#2563eb",
    },
  },

  typography: {
    fontFamily: {
      sans: "'Inter', system-ui, -apple-system, sans-serif",
      serif: "'Merriweather', Georgia, serif",
      mono: "'JetBrains Mono', 'Fira Code', monospace",
    },
    fontSize: {
      xs: "0.75rem",
      sm: "0.875rem",
      base: "1rem",
      lg: "1.125rem",
      xl: "1.25rem",
      "2xl": "1.5rem",
      "3xl": "1.875rem",
    },
    fontWeight: {
      light: "300",
      normal: "400",
      medium: "500",
      semibold: "600",
      bold: "700",
    },
    lineHeight: {
      tight: "1.25",
      normal: "1.5",
      relaxed: "1.75",
    },
  },

  spacing: {
    "0": "0",
    "1": "0.25rem",
    "2": "0.5rem",
    "3": "0.75rem",
    "4": "1rem",
    "5": "1.25rem",
    "6": "1.5rem",
    "8": "2rem",
    "10": "2.5rem",
    "12": "3rem",
    xs: "0.25rem",
    sm: "0.5rem",
    md: "1rem",
    lg: "1.5rem",
    xl: "2rem",
    "2xl": "3rem",
  },

  shadows: {
    none: "none",
    sm: "0 1px 2px 0 rgb(0 0 0 / 0.05)",
    md: "0 4px 6px -1px rgb(0 0 0 / 0.1), 0 2px 4px -2px rgb(0 0 0 / 0.1)",
    lg: "0 10px 15px -3px rgb(0 0 0 / 0.1), 0 4px 6px -4px rgb(0 0 0 / 0.1)",
    xl: "0 20px 25px -5px rgb(0 0 0 / 0.1), 0 8px 10px -6px rgb(0 0 0 / 0.1)",
  },

  radii: {
    none: "0",
    sm: "0.125rem",
    md: "0.375rem",
    lg: "0.5rem",
    xl: "0.75rem",
    full: "9999px",
  },

  transitions: {
    fast: "150ms cubic-bezier(0.4, 0, 0.2, 1)",
    normal: "250ms cubic-bezier(0.4, 0, 0.2, 1)",
    slow: "500ms cubic-bezier(0.4, 0, 0.2, 1)",
  },
};
