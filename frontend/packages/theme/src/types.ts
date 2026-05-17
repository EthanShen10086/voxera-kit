/** Color scale with main, light, and dark variants. */
export interface ColorScale {
  main: string;
  light: string;
  dark: string;
}

/** Complete color token definitions. */
export interface ColorTokens {
  primary: ColorScale;
  secondary: ColorScale;
  background: ColorScale;
  surface: ColorScale;
  text: ColorScale;
  error: ColorScale;
  warning: ColorScale;
  success: ColorScale;
  info: ColorScale;
}

/** Typography token definitions. */
export interface TypographyTokens {
  fontFamily: {
    sans: string;
    serif: string;
    mono: string;
  };
  fontSize: {
    xs: string;
    sm: string;
    base: string;
    lg: string;
    xl: string;
    "2xl": string;
    "3xl": string;
  };
  fontWeight: {
    light: string;
    normal: string;
    medium: string;
    semibold: string;
    bold: string;
  };
  lineHeight: {
    tight: string;
    normal: string;
    relaxed: string;
  };
}

/** Complete design token set for a theme. */
export interface ThemeTokens {
  colors: ColorTokens;
  typography: TypographyTokens;
  spacing: Record<string, string>;
  shadows: {
    sm: string;
    md: string;
    lg: string;
    xl: string;
    none: string;
  };
  radii: {
    none: string;
    sm: string;
    md: string;
    lg: string;
    xl: string;
    full: string;
  };
  transitions: {
    fast: string;
    normal: string;
    slow: string;
  };
}

/** A named theme preset containing a full token set. */
export interface ThemePreset {
  /** Human-readable name of the preset (e.g. "light", "dark"). */
  name: string;
  /** The complete design tokens for this preset. */
  tokens: ThemeTokens;
}

/** Theme change callback signature. */
export type ThemeChangeCallback = (preset: ThemePreset) => void;

/** Framework-agnostic theme engine interface. */
export interface IThemeEngine {
  /** Apply a theme preset, setting CSS custom properties on the document root. */
  apply(preset: ThemePreset): void;

  /** Read a token value by dot-separated path (e.g. "colors.primary.main"). */
  getToken(path: string): string | undefined;

  /** Get the currently applied preset, or null if none has been applied. */
  getCurrentPreset(): ThemePreset | null;

  /**
   * Subscribe to theme changes.
   * @returns An unsubscribe function.
   */
  onThemeChange(callback: ThemeChangeCallback): () => void;
}
