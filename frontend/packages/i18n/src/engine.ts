import type {
  II18nEngine,
  InterpolationParams,
  Locale,
  TranslationMap,
} from "./types.js";
import { interpolate } from "./interpolation.js";

/**
 * Deep-merge `source` into `target`, mutating `target` in place.
 */
function deepMerge(target: TranslationMap, source: TranslationMap): void {
  for (const key of Object.keys(source)) {
    const srcVal = source[key];
    const tgtVal = target[key];

    if (
      typeof srcVal === "object" &&
      srcVal !== null &&
      typeof tgtVal === "object" &&
      tgtVal !== null
    ) {
      deepMerge(tgtVal as TranslationMap, srcVal as TranslationMap);
    } else {
      target[key] = srcVal;
    }
  }
}

/**
 * Resolve a dot-notation key against a nested `TranslationMap`.
 * Returns `undefined` when the path does not resolve to a string leaf.
 */
function resolveKey(
  translations: TranslationMap,
  key: string,
): string | undefined {
  const segments = key.split(".");
  let current: string | TranslationMap = translations;

  for (const segment of segments) {
    if (typeof current === "string" || current === undefined) {
      return undefined;
    }
    current = (current as TranslationMap)[segment];
  }

  return typeof current === "string" ? current : undefined;
}

export class I18nEngine implements II18nEngine {
  private readonly store = new Map<Locale, TranslationMap>();
  private currentLocale: Locale;
  private readonly listeners = new Set<(locale: Locale) => void>();

  constructor(defaultLocale: Locale = "en") {
    this.currentLocale = defaultLocale;
  }

  loadLocale(locale: Locale, translations: TranslationMap): void {
    const existing = this.store.get(locale);
    if (existing) {
      deepMerge(existing, translations);
    } else {
      this.store.set(locale, structuredClone(translations));
    }
  }

  t(key: string, params?: InterpolationParams): string {
    const translations = this.store.get(this.currentLocale);
    if (!translations) {
      return key;
    }

    const value = resolveKey(translations, key);
    if (value === undefined) {
      return key;
    }

    return params ? interpolate(value, params) : value;
  }

  setLocale(locale: Locale): void {
    if (locale === this.currentLocale) {
      return;
    }
    this.currentLocale = locale;
    for (const listener of this.listeners) {
      listener(locale);
    }
  }

  getLocale(): Locale {
    return this.currentLocale;
  }

  getAvailableLocales(): Locale[] {
    return [...this.store.keys()];
  }

  hasKey(key: string, locale?: Locale): boolean {
    const target = locale ?? this.currentLocale;
    const translations = this.store.get(target);
    if (!translations) {
      return false;
    }
    return resolveKey(translations, key) !== undefined;
  }

  onLocaleChange(callback: (locale: Locale) => void): () => void {
    this.listeners.add(callback);
    return () => {
      this.listeners.delete(callback);
    };
  }
}
