import type { ThemePreset } from "../types.js";
import { defaultTokens } from "../tokens.js";

/** Light theme preset – the default theme using the standard token palette. */
export const lightPreset: ThemePreset = {
  name: "light",
  tokens: { ...defaultTokens },
};
