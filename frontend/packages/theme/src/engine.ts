import type { IThemeEngine, ThemeChangeCallback, ThemePreset } from "./types.js";

/**
 * Framework-agnostic theme engine that converts design tokens to CSS custom
 * properties on `document.documentElement`.
 *
 * SSR-safe: operations that touch the DOM are guarded by a `typeof document` check.
 */
export class ThemeEngine implements IThemeEngine {
  private currentPreset: ThemePreset | null = null;
  private readonly subscribers = new Set<ThemeChangeCallback>();

  apply(preset: ThemePreset): void {
    this.currentPreset = preset;
    const properties = this.flattenTokens(preset.tokens as unknown as Record<string, unknown>);

    if (typeof document !== "undefined") {
      const root = document.documentElement;
      for (const [key, value] of properties) {
        root.style.setProperty(key, value);
      }
    }

    for (const cb of this.subscribers) {
      try {
        cb(preset);
      } catch {
        // subscriber errors must not break the engine
      }
    }
  }

  getToken(path: string): string | undefined {
    const varName = `--vk-${path.replace(/\./g, "-")}`;

    if (typeof document !== "undefined") {
      const value = getComputedStyle(document.documentElement)
        .getPropertyValue(varName)
        .trim();
      return value || undefined;
    }

    if (!this.currentPreset) return undefined;
    return this.resolveTokenPath(this.currentPreset.tokens as unknown as Record<string, unknown>, path);
  }

  getCurrentPreset(): ThemePreset | null {
    return this.currentPreset;
  }

  onThemeChange(callback: ThemeChangeCallback): () => void {
    this.subscribers.add(callback);
    return () => {
      this.subscribers.delete(callback);
    };
  }

  // ── Private helpers ──────────────────────────────────────────────

  /**
   * Flatten a nested token object into CSS custom property entries.
   * e.g. `{ colors: { primary: { main: '#fff' } } }` → `[['--vk-colors-primary-main', '#fff']]`
   */
  private flattenTokens(
    obj: Record<string, unknown>,
    prefix = "--vk",
  ): Array<[string, string]> {
    const entries: Array<[string, string]> = [];

    for (const [key, value] of Object.entries(obj)) {
      const cssKey = `${prefix}-${key}`;

      if (typeof value === "object" && value !== null && !Array.isArray(value)) {
        entries.push(
          ...this.flattenTokens(value as Record<string, unknown>, cssKey),
        );
      } else {
        entries.push([cssKey, String(value)]);
      }
    }

    return entries;
  }

  /** Resolve a dot-separated path against an object tree. */
  private resolveTokenPath(
    obj: Record<string, unknown>,
    path: string,
  ): string | undefined {
    const segments = path.split(".");
    let current: unknown = obj;

    for (const segment of segments) {
      if (current === null || current === undefined || typeof current !== "object") {
        return undefined;
      }
      current = (current as Record<string, unknown>)[segment];
    }

    return current !== undefined && current !== null
      ? String(current)
      : undefined;
  }
}
