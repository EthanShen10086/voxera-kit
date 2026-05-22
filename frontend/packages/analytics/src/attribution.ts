const FIRST_TOUCH_KEY = 'voxera_first_touch';
const LAST_TOUCH_KEY = 'voxera_last_touch';

export interface Attribution {
  utmSource?: string;
  utmMedium?: string;
  utmCampaign?: string;
  utmTerm?: string;
  utmContent?: string;
  timestamp: number;
}

export class AttributionTracker {
  constructor() {
    this.captureFromUrl();
  }

  getFirstTouch(): Attribution | null {
    return this.load(FIRST_TOUCH_KEY);
  }

  getLastTouch(): Attribution | null {
    return this.load(LAST_TOUCH_KEY);
  }

  getCurrentAttribution(): Attribution | null {
    return this.parseUrlParams();
  }

  private captureFromUrl(): void {
    const attribution = this.parseUrlParams();
    if (!attribution) return;

    if (!this.load(FIRST_TOUCH_KEY)) {
      this.save(FIRST_TOUCH_KEY, attribution);
    }
    this.save(LAST_TOUCH_KEY, attribution);
  }

  private parseUrlParams(): Attribution | null {
    try {
      const params = new URLSearchParams(window.location.search);
      const utmSource = params.get('utm_source') ?? undefined;
      const utmMedium = params.get('utm_medium') ?? undefined;
      const utmCampaign = params.get('utm_campaign') ?? undefined;
      const utmTerm = params.get('utm_term') ?? undefined;
      const utmContent = params.get('utm_content') ?? undefined;

      if (!utmSource && !utmMedium && !utmCampaign && !utmTerm && !utmContent) {
        return null;
      }

      return { utmSource, utmMedium, utmCampaign, utmTerm, utmContent, timestamp: Date.now() };
    } catch {
      return null;
    }
  }

  private save(key: string, attribution: Attribution): void {
    try {
      localStorage.setItem(key, JSON.stringify(attribution));
    } catch {
      // localStorage unavailable
    }
  }

  private load(key: string): Attribution | null {
    try {
      const raw = localStorage.getItem(key);
      if (!raw) return null;
      return JSON.parse(raw) as Attribution;
    } catch {
      return null;
    }
  }
}
