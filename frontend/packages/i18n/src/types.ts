export type Locale = string;

export type TranslationMap = Record<string, string | TranslationMap>;

export type InterpolationParams = Record<string, string | number>;

export interface PluralRule {
  zero?: string;
  one: string;
  other: string;
  few?: string;
  many?: string;
}

export interface II18nEngine {
  loadLocale(locale: Locale, translations: TranslationMap): void;
  t(key: string, params?: InterpolationParams): string;
  setLocale(locale: Locale): void;
  getLocale(): Locale;
  getAvailableLocales(): Locale[];
  hasKey(key: string, locale?: Locale): boolean;
  onLocaleChange(callback: (locale: Locale) => void): () => void;
}
