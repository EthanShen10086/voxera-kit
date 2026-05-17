import type { IPerformanceMonitor, WebVitalsMetrics } from "./types.js";

const isBrowser = typeof globalThis !== "undefined" && typeof (globalThis as any).window !== "undefined";

export class PerformanceMonitor implements IPerformanceMonitor {
  private readonly marks = new Map<string, number>();
  private readonly metrics = new Map<string, number>();

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

    const navEntries = performance.getEntriesByType("navigation") as PerformanceNavigationTiming[];
    if (navEntries.length > 0) {
      const nav = navEntries[0]!;
      result.ttfb = nav.responseStart - nav.requestStart;
    }

    return result;
  }

  markStart(name: string): void {
    const now = isBrowser && typeof performance !== "undefined" ? performance.now() : Date.now();
    this.marks.set(name, now);

    if (isBrowser && typeof performance !== "undefined") {
      performance.mark(`${name}:start`);
    }
  }

  markEnd(name: string): number {
    const start = this.marks.get(name);
    const now = isBrowser && typeof performance !== "undefined" ? performance.now() : Date.now();

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

  measure<T>(name: string, fn: () => T): T {
    this.markStart(name);
    try {
      return fn();
    } finally {
      this.markEnd(name);
    }
  }

  getMetrics(): Record<string, number> {
    return Object.fromEntries(this.metrics);
  }
}
