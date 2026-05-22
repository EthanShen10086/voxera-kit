import type { IPerformanceMonitor, WebVitalsMetrics } from "./types.js";

const isBrowser =
  typeof globalThis !== "undefined" &&
  typeof (globalThis as unknown as { window?: unknown }).window !== "undefined";

/** Configuration for continuous Web Vitals observation. */
export interface ObserveConfig {
  /** Duration threshold (ms) above which a resource is flagged as slow. Defaults to 2000. */
  resourceSlowThreshold?: number;
  /** Callback invoked whenever a Web Vital metric is recorded or updated. */
  onVital?: (name: string, value: number) => void;
}

/** A resource timing entry that exceeded the configured slow threshold. */
export interface SlowResource {
  name: string;
  duration: number;
  initiatorType: string;
  startTime: number;
}

/** Metadata about a long task observed via the Long Tasks API. */
export interface LongTaskEntry {
  startTime: number;
  duration: number;
}

/**
 * Monitors page performance using the Performance API and PerformanceObserver.
 *
 * Provides both a one-shot {@link collectWebVitals} snapshot and a continuous
 * {@link startObserving} mode that tracks LCP, CLS, INP, FID, Long Tasks,
 * and slow resources in real time.
 */
export class PerformanceMonitor implements IPerformanceMonitor {
  private readonly marks = new Map<string, number>();
  private readonly metrics = new Map<string, number>();

  private readonly observers: PerformanceObserver[] = [];
  private observing = false;

  private lcpValue = 0;
  private clsValue = 0;
  private fidValue = 0;
  private inpCandidates: number[] = [];

  private readonly longTasks: LongTaskEntry[] = [];
  private readonly slowResources: SlowResource[] = [];

  /**
   * Collect a one-shot snapshot of Web Vitals that are available from the
   * static Performance Timeline (FCP, TTFB). For continuous, real-time
   * metrics use {@link startObserving}.
   */
  async collectWebVitals(): Promise<WebVitalsMetrics> {
    const result: WebVitalsMetrics = {};

    if (!isBrowser || typeof performance === "undefined") {
      return result;
    }

    const paintEntries = performance.getEntriesByType("paint");
    for (const entry of paintEntries) {
      if (entry.name === "first-contentful-paint") {
        result.fcp = entry.startTime;
      }
    }

    const navEntries =
      performance.getEntriesByType("navigation") as PerformanceNavigationTiming[];
    if (navEntries.length > 0) {
      const nav = navEntries[0]!;
      result.ttfb = nav.responseStart - nav.requestStart;
    }

    if (this.observing) {
      if (this.lcpValue > 0) result.lcp = this.lcpValue;
      if (this.clsValue > 0) result.cls = this.clsValue;
      if (this.fidValue > 0) result.fid = this.fidValue;
      const inp = this.computeInp();
      if (inp > 0) result.inp = inp;
    }

    return result;
  }

  /**
   * Start continuous observation of Web Vitals using PerformanceObserver.
   *
   * Observers are created for LCP, CLS, FID, INP (event timing), Long Tasks,
   * and Resource Timing. Each metric update is forwarded to the optional
   * {@link ObserveConfig.onVital} callback.
   *
   * Calling this method more than once is a no-op.
   */
  startObserving(config: ObserveConfig = {}): void {
    if (this.observing) return;
    if (!isBrowser || typeof PerformanceObserver === "undefined") return;

    this.observing = true;
    const slowThreshold = config.resourceSlowThreshold ?? 2000;
    const notify = config.onVital;

    this.observeLcp(notify);
    this.observeCls(notify);
    this.observeFid(notify);
    this.observeInp(notify);
    this.observeLongTasks(notify);
    this.observeResources(slowThreshold, notify);
  }

  /** Stop all active PerformanceObservers. */
  stopObserving(): void {
    for (const obs of this.observers) {
      obs.disconnect();
    }
    this.observers.length = 0;
    this.observing = false;
  }

  /** Return all long tasks recorded since observation started. */
  getLongTasks(): ReadonlyArray<LongTaskEntry> {
    return this.longTasks;
  }

  /** Return all slow resources recorded since observation started. */
  getSlowResources(): ReadonlyArray<SlowResource> {
    return this.slowResources;
  }

  /** Record the start of a named timing region. */
  markStart(name: string): void {
    const now =
      isBrowser && typeof performance !== "undefined" ? performance.now() : Date.now();
    this.marks.set(name, now);

    if (isBrowser && typeof performance !== "undefined") {
      performance.mark(`${name}:start`);
    }
  }

  /** End a named timing region and return the measured duration in milliseconds. */
  markEnd(name: string): number {
    const start = this.marks.get(name);
    const now =
      isBrowser && typeof performance !== "undefined" ? performance.now() : Date.now();

    if (start === undefined) {
      return 0;
    }

    const duration = now - start;
    this.metrics.set(name, duration);
    this.marks.delete(name);

    if (isBrowser && typeof performance !== "undefined") {
      performance.mark(`${name}:end`);
      performance.measure(name, `${name}:start`, `${name}:end`);
    }

    return duration;
  }

  /** Measure the synchronous execution of {@link fn} and return its result. */
  measure<T>(name: string, fn: () => T): T {
    this.markStart(name);
    try {
      return fn();
    } finally {
      this.markEnd(name);
    }
  }

  /** Return all manually recorded metrics as a plain object. */
  getMetrics(): Record<string, number> {
    return Object.fromEntries(this.metrics);
  }

  // ---------------------------------------------------------------------------
  // Private observer helpers
  // ---------------------------------------------------------------------------

  private tryObserve(
    type: string,
    callback: (entries: PerformanceEntryList) => void,
    options?: { buffered?: boolean; durationThreshold?: number },
  ): void {
    try {
      const observer = new PerformanceObserver((list) => {
        callback(list.getEntries());
      });

      const observeOptions: PerformanceObserverInit = {
        type,
        buffered: options?.buffered ?? true,
      };

      if (options?.durationThreshold !== undefined) {
        (observeOptions as PerformanceObserverInit & { durationThreshold: number })
          .durationThreshold = options.durationThreshold;
      }

      observer.observe(observeOptions);
      this.observers.push(observer);
    } catch {
      // Entry type not supported in this browser — silently skip.
    }
  }

  private observeLcp(notify?: (name: string, value: number) => void): void {
    this.tryObserve("largest-contentful-paint", (entries) => {
      const last = entries[entries.length - 1];
      if (last) {
        this.lcpValue = last.startTime;
        notify?.("lcp", this.lcpValue);
      }
    });
  }

  private observeCls(notify?: (name: string, value: number) => void): void {
    this.tryObserve("layout-shift", (entries) => {
      for (const entry of entries) {
        const shift = entry as PerformanceEntry & {
          hadRecentInput: boolean;
          value: number;
        };
        if (!shift.hadRecentInput) {
          this.clsValue += shift.value;
        }
      }
      notify?.("cls", this.clsValue);
    });
  }

  private observeFid(notify?: (name: string, value: number) => void): void {
    this.tryObserve("first-input", (entries) => {
      const first = entries[0];
      if (first) {
        const fi = first as PerformanceEntry & { processingStart: number };
        this.fidValue = fi.processingStart - fi.startTime;
        notify?.("fid", this.fidValue);
      }
    });
  }

  private observeInp(notify?: (name: string, value: number) => void): void {
    this.tryObserve(
      "event",
      (entries) => {
        for (const entry of entries) {
          this.inpCandidates.push(entry.duration);
        }
        const inp = this.computeInp();
        if (inp > 0) {
          notify?.("inp", inp);
        }
      },
      { durationThreshold: 16 },
    );
  }

  private observeLongTasks(notify?: (name: string, value: number) => void): void {
    this.tryObserve("longtask", (entries) => {
      for (const entry of entries) {
        if (entry.duration > 50) {
          this.longTasks.push({
            startTime: entry.startTime,
            duration: entry.duration,
          });
          notify?.("longtask", entry.duration);
        }
      }
    });
  }

  private observeResources(
    slowThreshold: number,
    notify?: (name: string, value: number) => void,
  ): void {
    this.tryObserve("resource", (entries) => {
      for (const entry of entries) {
        if (entry.duration > slowThreshold) {
          const res = entry as PerformanceResourceTiming;
          this.slowResources.push({
            name: res.name,
            duration: res.duration,
            initiatorType: res.initiatorType,
            startTime: res.startTime,
          });
          notify?.("slow-resource", res.duration);
        }
      }
    });
  }

  /** Compute INP as the P98 of all event durations. */
  private computeInp(): number {
    if (this.inpCandidates.length === 0) return 0;
    const sorted = [...this.inpCandidates].sort((a, b) => a - b);
    const idx = Math.min(
      Math.ceil(sorted.length * 0.98) - 1,
      sorted.length - 1,
    );
    return sorted[idx]!;
  }
}
