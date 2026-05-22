import type { AnalyticsConfig } from './types';

type TrackCallback = (name: string, properties?: Record<string, unknown>) => void;

export class AutoTracker {
  private config: AnalyticsConfig;
  private track: TrackCallback;
  private cleanups: Array<() => void> = [];

  constructor(config: AnalyticsConfig, track: TrackCallback) {
    this.config = config;
    this.track = track;
  }

  start(): void {
    const auto = this.config.autoTrack;
    if (!auto) return;

    if (auto.pageViews) this.trackPageViews();
    if (auto.clicks) this.trackClicks();
    if (auto.scrollDepth) this.trackScrollDepth();
    if (auto.formSubmissions) this.trackFormSubmissions();
    if (auto.outboundLinks) this.trackOutboundLinks();
    if (auto.timeOnPage) this.trackTimeOnPage();
    if (auto.errors) this.trackErrors();
  }

  stop(): void {
    for (const cleanup of this.cleanups) cleanup();
    this.cleanups = [];
  }

  private trackPageViews(): void {
    this.track('$pageview', { url: location.href, title: document.title });

    const handler = () => {
      this.track('$pageview', { url: location.href, title: document.title });
    };
    window.addEventListener('popstate', handler);
    this.cleanups.push(() => window.removeEventListener('popstate', handler));
  }

  private trackClicks(): void {
    const handler = (e: MouseEvent) => {
      const target = e.target as HTMLElement | null;
      if (!target) return;

      const trackable = target.closest('a, button, [data-track]') as HTMLElement | null;
      if (!trackable) return;

      const props: Record<string, unknown> = {
        tagName: trackable.tagName.toLowerCase(),
        text: (trackable.textContent ?? '').trim().slice(0, 200),
      };

      if (trackable.hasAttribute('data-track')) {
        props.trackId = trackable.getAttribute('data-track');
      }
      if (trackable instanceof HTMLAnchorElement) {
        props.href = trackable.href;
      }

      this.track('$click', props);
    };
    document.addEventListener('click', handler, true);
    this.cleanups.push(() => document.removeEventListener('click', handler, true));
  }

  private trackScrollDepth(): void {
    const thresholds = [25, 50, 75, 100];
    const reached = new Set<number>();
    let throttleTimer: ReturnType<typeof setTimeout> | null = null;

    const handler = () => {
      if (throttleTimer) return;
      throttleTimer = setTimeout(() => {
        throttleTimer = null;
        const scrollTop = window.scrollY;
        const docHeight = document.documentElement.scrollHeight - window.innerHeight;
        if (docHeight <= 0) return;

        const percent = Math.round((scrollTop / docHeight) * 100);
        for (const t of thresholds) {
          if (percent >= t && !reached.has(t)) {
            reached.add(t);
            this.track('$scroll_depth', { depth: t, url: location.href });
          }
        }
      }, 250);
    };
    window.addEventListener('scroll', handler, { passive: true });
    this.cleanups.push(() => {
      window.removeEventListener('scroll', handler);
      if (throttleTimer) clearTimeout(throttleTimer);
    });
  }

  private trackFormSubmissions(): void {
    const handler = (e: Event) => {
      const form = e.target as HTMLFormElement;
      this.track('$form_submit', {
        formId: form.id || undefined,
        formName: form.name || undefined,
        formAction: form.action || undefined,
      });
    };
    document.addEventListener('submit', handler, true);
    this.cleanups.push(() => document.removeEventListener('submit', handler, true));
  }

  private trackOutboundLinks(): void {
    const handler = (e: MouseEvent) => {
      const anchor = (e.target as HTMLElement)?.closest?.('a') as HTMLAnchorElement | null;
      if (!anchor?.href) return;

      try {
        const url = new URL(anchor.href);
        if (url.hostname !== location.hostname) {
          this.track('$outbound_link', { url: anchor.href, text: (anchor.textContent ?? '').trim().slice(0, 200) });
        }
      } catch {
        // invalid URL
      }
    };
    document.addEventListener('click', handler, true);
    this.cleanups.push(() => document.removeEventListener('click', handler, true));
  }

  private trackTimeOnPage(): void {
    const startTime = Date.now();
    let totalVisibleMs = 0;
    let lastVisibleAt = document.hidden ? 0 : Date.now();

    const intervalId = setInterval(() => {
      if (!document.hidden && lastVisibleAt > 0) {
        totalVisibleMs += Date.now() - lastVisibleAt;
        lastVisibleAt = Date.now();
      }
    }, 5000);

    const visibilityHandler = () => {
      if (document.hidden) {
        if (lastVisibleAt > 0) {
          totalVisibleMs += Date.now() - lastVisibleAt;
          lastVisibleAt = 0;
        }
        this.track('$time_on_page', {
          totalSeconds: Math.round(totalVisibleMs / 1000),
          wallSeconds: Math.round((Date.now() - startTime) / 1000),
          url: location.href,
        });
      } else {
        lastVisibleAt = Date.now();
      }
    };
    document.addEventListener('visibilitychange', visibilityHandler);
    this.cleanups.push(() => {
      clearInterval(intervalId);
      document.removeEventListener('visibilitychange', visibilityHandler);
    });
  }

  private trackErrors(): void {
    const errorHandler = (e: ErrorEvent) => {
      this.track('$error', {
        message: e.message,
        filename: e.filename,
        lineno: e.lineno,
        colno: e.colno,
      });
    };
    const rejectionHandler = (e: PromiseRejectionEvent) => {
      this.track('$unhandled_rejection', {
        reason: String(e.reason),
      });
    };
    window.addEventListener('error', errorHandler);
    window.addEventListener('unhandledrejection', rejectionHandler);
    this.cleanups.push(() => {
      window.removeEventListener('error', errorHandler);
      window.removeEventListener('unhandledrejection', rejectionHandler);
    });
  }
}
