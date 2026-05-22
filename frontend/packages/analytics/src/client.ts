import type { AnalyticsConfig, AnalyticsEvent, EventContext, UserProperties } from './types';
import type { IAnalyticsProvider } from './provider';
import { SessionManager } from './session';
import { AttributionTracker } from './attribution';
import { AutoTracker } from './auto-track';
import { HttpProvider } from './http-provider';

const DEFAULT_BATCH_SIZE = 20;
const DEFAULT_FLUSH_INTERVAL = 10_000;
const DEVICE_ID_KEY = 'voxera_device_id';

export class AnalyticsClient {
  private config: AnalyticsConfig;
  private provider: IAnalyticsProvider;
  private session: SessionManager;
  private attribution: AttributionTracker;
  private autoTracker: AutoTracker | null = null;
  private buffer: AnalyticsEvent[] = [];
  private flushTimer: ReturnType<typeof setInterval> | null = null;
  private userId?: string;
  private userProperties: UserProperties = {};
  private deviceId: string;
  private disposed = false;

  constructor(config: AnalyticsConfig, provider?: IAnalyticsProvider) {
    this.config = config;
    this.provider = provider ?? new HttpProvider(config.endpoint, config.apiKey);
    this.session = new SessionManager(config.sessionTimeout);
    this.attribution = new AttributionTracker();
    this.deviceId = this.getOrCreateDeviceId();

    if (config.autoTrack) {
      this.autoTracker = new AutoTracker(config, (name, props) => this.track(name, props));
      this.autoTracker.start();
    }

    const flushInterval = config.flushInterval ?? DEFAULT_FLUSH_INTERVAL;
    this.flushTimer = setInterval(() => void this.flush(), flushInterval);

    if (typeof window !== 'undefined') {
      window.addEventListener('visibilitychange', this.onVisibilityChange);
      window.addEventListener('pagehide', this.onPageHide);
    }
  }

  track(name: string, properties?: Record<string, unknown>): void {
    if (this.disposed) return;
    this.session.touch();

    const event: AnalyticsEvent = {
      name,
      properties,
      userId: this.userId,
      sessionId: this.session.getSessionId(),
      timestamp: Date.now(),
      context: this.buildContext(),
    };

    this.buffer.push(event);
    this.log('track', event);

    if (this.buffer.length >= (this.config.batchSize ?? DEFAULT_BATCH_SIZE)) {
      void this.flush();
    }
  }

  identify(userId: string, properties?: UserProperties): void {
    this.userId = userId;
    if (properties) Object.assign(this.userProperties, properties);
    void this.provider.identify(userId, properties);
    this.log('identify', { userId, properties });
  }

  alias(previousId: string, newId: string): void {
    void this.provider.alias(previousId, newId);
  }

  group(groupId: string, properties?: Record<string, unknown>): void {
    void this.provider.group(groupId, properties);
  }

  page(name?: string, properties?: Record<string, unknown>): void {
    this.track('$pageview', {
      ...properties,
      pageName: name,
      url: location.href,
      title: document.title,
      referrer: document.referrer,
    });
  }

  setUserProperties(props: UserProperties): void {
    Object.assign(this.userProperties, props);
  }

  reset(): void {
    this.userId = undefined;
    this.userProperties = {};
    this.session.reset();
    this.provider.reset();
    this.buffer = [];
  }

  async flush(): Promise<void> {
    if (this.buffer.length === 0) return;
    const events = this.buffer.splice(0);
    try {
      await this.provider.trackBatch(events);
    } catch {
      this.buffer.unshift(...events);
    }
  }

  dispose(): void {
    this.disposed = true;
    this.autoTracker?.stop();
    if (this.flushTimer) clearInterval(this.flushTimer);
    if (typeof window !== 'undefined') {
      window.removeEventListener('visibilitychange', this.onVisibilityChange);
      window.removeEventListener('pagehide', this.onPageHide);
    }
    void this.flush();
  }

  private buildContext(): EventContext {
    const lastTouch = this.attribution.getLastTouch();
    try {
      return {
        platform: 'web',
        locale: navigator.language,
        userAgent: navigator.userAgent,
        pageUrl: location.href,
        referrer: document.referrer || undefined,
        utmSource: lastTouch?.utmSource,
        utmMedium: lastTouch?.utmMedium,
        utmCampaign: lastTouch?.utmCampaign,
        utmTerm: lastTouch?.utmTerm,
        utmContent: lastTouch?.utmContent,
        screenWidth: window.screen.width,
        screenHeight: window.screen.height,
        deviceId: this.deviceId,
      };
    } catch {
      return { deviceId: this.deviceId };
    }
  }

  private getOrCreateDeviceId(): string {
    try {
      const existing = localStorage.getItem(DEVICE_ID_KEY);
      if (existing) return existing;
      const id = crypto.randomUUID();
      localStorage.setItem(DEVICE_ID_KEY, id);
      return id;
    } catch {
      return crypto.randomUUID();
    }
  }

  private onVisibilityChange = (): void => {
    if (document.hidden) void this.flush();
  };

  private onPageHide = (): void => {
    void this.flush();
  };

  private log(action: string, data: unknown): void {
    if (this.config.debug) {
      console.debug(`[voxera-analytics] ${action}`, data);
    }
  }
}
