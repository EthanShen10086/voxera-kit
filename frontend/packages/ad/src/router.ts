import type { IAdProvider } from './provider';
import type { Ad, AdConfig, AdRequest } from './types';

export class AdRouter {
  private providers: IAdProvider[] = [];
  private config: AdConfig;

  constructor(config: AdConfig) {
    this.config = config;
  }

  register(provider: IAdProvider): void {
    this.providers.push(provider);
    this.providers.sort((a, b) => {
      const pa = this.config.providers.find((p) => p.name === a.name);
      const pb = this.config.providers.find((p) => p.name === b.name);
      return (pa?.priority ?? 99) - (pb?.priority ?? 99);
    });
  }

  async fetch(
    request: AdRequest,
    options?: { isPaidUser?: boolean; isMinor?: boolean },
  ): Promise<Ad | null> {
    if (!this.config.enabled) return null;
    if (options?.isPaidUser) return null;
    if (options?.isMinor && this.config.minorPolicy === 'hide') return null;

    for (const provider of this.providers) {
      if (!provider.available()) continue;

      try {
        const ad = await provider.fetchAd(request);
        if (ad) return ad;
      } catch {
        continue;
      }
    }

    return null;
  }
}
