import { LogLevel } from "./types.js";
import type { Breadcrumb, IErrorTracker } from "./types.js";

const MAX_BREADCRUMBS = 100;

function generateEventId(): string {
  const chars = "0123456789abcdef";
  let id = "";
  for (let i = 0; i < 32; i++) {
    id += chars[Math.floor(Math.random() * chars.length)];
  }
  return id;
}

/**
 * Build a simple fingerprint from an error for deduplication purposes.
 */
function fingerprint(error: Error): string {
  return `${error.name}:${error.message}`;
}

/** User context attached to reported error events. */
export interface UserContext {
  id: string;
  email?: string;
  username?: string;
}

/** Serializable error event sent to the reporting endpoint. */
export interface ErrorEvent {
  type: "exception" | "message";
  message: string;
  stack?: string;
  breadcrumbs: Breadcrumb[];
  user?: UserContext;
  tags: Record<string, string>;
  release?: string;
  environment?: string;
  timestamp: number;
}

/** Configuration for remote error reporting via HTTP. */
export interface ReportConfig {
  /** URL to POST error events to. */
  endpoint: string;
  /** Application release/version identifier. */
  release?: string;
  /** Deployment environment (e.g. `"production"`, `"staging"`). */
  environment?: string;
  /** Fraction (0–1) of errors to report. Defaults to 1. */
  sampleRate?: number;
  /** Maximum breadcrumbs to include in each event. Defaults to {@link MAX_BREADCRUMBS}. */
  maxBreadcrumbs?: number;
  /** Transform or filter an error event before sending. Return `null` to drop. */
  beforeSend?: (event: ErrorEvent) => ErrorEvent | null;
}

/** Options accepted by the {@link ErrorTracker} constructor. */
export interface ErrorTrackerOptions {
  /** Called when an exception or message is captured. */
  onCapture?: (entry: CapturedEntry) => void;
  /** Maximum number of breadcrumbs to retain. */
  maxBreadcrumbs?: number;
  /** Install global error handlers automatically. */
  installGlobalHandlers?: boolean;
  /** Enable remote error reporting when provided. */
  reportConfig?: ReportConfig;
}

/** Internal representation of a captured error or message. */
export interface CapturedEntry {
  eventId: string;
  type: "exception" | "message";
  message: string;
  level: LogLevel;
  timestamp: number;
  error?: Error;
  context?: Record<string, unknown>;
  breadcrumbs: Breadcrumb[];
  user?: UserContext;
  fingerprint: string;
}

/**
 * Captures exceptions and messages, records breadcrumbs, and optionally
 * reports errors to a remote endpoint with batching, sampling, and
 * pre-send filtering.
 */
export class ErrorTracker implements IErrorTracker {
  private readonly breadcrumbs: Breadcrumb[] = [];
  private readonly maxBreadcrumbs: number;
  private readonly seenFingerprints = new Set<string>();
  private user?: UserContext;
  private tags: Record<string, string> = {};
  private readonly onCapture?: (entry: CapturedEntry) => void;
  private readonly reportConfig?: ReportConfig;

  private readonly reportBuffer: ErrorEvent[] = [];
  private reportTimer: ReturnType<typeof setTimeout> | null = null;

  private boundOnError?: (event: ErrorEvent) => void;
  private boundOnRejection?: (event: PromiseRejectionEvent) => void;

  constructor(options: ErrorTrackerOptions = {}) {
    this.maxBreadcrumbs = options.maxBreadcrumbs ?? MAX_BREADCRUMBS;
    this.onCapture = options.onCapture;
    this.reportConfig = options.reportConfig;

    if (options.installGlobalHandlers !== false) {
      this.installGlobalHandlers();
    }
  }

  /**
   * Capture an exception, record it locally, and enqueue it for remote
   * reporting (if configured). Returns the event ID.
   */
  captureException(error: Error, context?: Record<string, unknown>): string {
    const fp = fingerprint(error);
    const eventId = generateEventId();

    const entry: CapturedEntry = {
      eventId,
      type: "exception",
      message: error.message,
      level: LogLevel.Error,
      timestamp: Date.now(),
      error,
      context,
      breadcrumbs: [...this.breadcrumbs],
      user: this.user,
      fingerprint: fp,
    };

    this.seenFingerprints.add(fp);
    this.onCapture?.(entry);
    this.enqueueReport(entry);

    return eventId;
  }

  /**
   * Capture a plain message at the given severity level and enqueue it for
   * remote reporting (if configured). Returns the event ID.
   */
  captureMessage(message: string, level: LogLevel = LogLevel.Info): string {
    const eventId = generateEventId();

    const entry: CapturedEntry = {
      eventId,
      type: "message",
      message,
      level,
      timestamp: Date.now(),
      breadcrumbs: [...this.breadcrumbs],
      user: this.user,
      fingerprint: `message:${message}`,
    };

    this.onCapture?.(entry);
    this.enqueueReport(entry);
    return eventId;
  }

  /** Set user context attached to all subsequent error events. */
  setUser(user: UserContext): void {
    this.user = user;
  }

  /** Set a tag that will be attached to all subsequent error events. */
  setTag(key: string, value: string): void {
    this.tags[key] = value;
  }

  /** Add a breadcrumb to the trail. Oldest breadcrumbs are evicted when the limit is reached. */
  addBreadcrumb(breadcrumb: Breadcrumb): void {
    this.breadcrumbs.push(breadcrumb);
    if (this.breadcrumbs.length > this.maxBreadcrumbs) {
      this.breadcrumbs.shift();
    }
  }

  /** Flush all buffered error events to the remote endpoint immediately. */
  async flush(): Promise<void> {
    if (this.reportTimer !== null) {
      clearTimeout(this.reportTimer);
      this.reportTimer = null;
    }

    if (this.reportBuffer.length === 0 || !this.reportConfig) return;

    const batch = this.reportBuffer.splice(0, this.reportBuffer.length);
    await this.sendBatch(batch);
  }

  /** Uninstall global handlers and flush remaining events. */
  dispose(): void {
    if (
      typeof globalThis === "undefined" ||
      typeof (globalThis as unknown as { window?: unknown }).window === "undefined"
    ) {
      return;
    }

    const win = globalThis as unknown as Window;
    if (this.boundOnError) {
      win.removeEventListener(
        "error",
        this.boundOnError as unknown as EventListenerOrEventListenerObject,
      );
    }
    if (this.boundOnRejection) {
      win.removeEventListener(
        "unhandledrejection",
        this.boundOnRejection as unknown as EventListenerOrEventListenerObject,
      );
    }

    void this.flush();
  }

  // ---------------------------------------------------------------------------
  // Private — reporting
  // ---------------------------------------------------------------------------

  private enqueueReport(entry: CapturedEntry): void {
    if (!this.reportConfig) return;

    const sampleRate = this.reportConfig.sampleRate ?? 1;
    if (Math.random() >= sampleRate) return;

    const maxCrumbs = this.reportConfig.maxBreadcrumbs ?? this.maxBreadcrumbs;
    const crumbs = entry.breadcrumbs.slice(-maxCrumbs);

    let event: ErrorEvent | null = {
      type: entry.type,
      message: entry.message,
      stack: entry.error?.stack,
      breadcrumbs: crumbs,
      user: entry.user,
      tags: { ...this.tags },
      release: this.reportConfig.release,
      environment: this.reportConfig.environment,
      timestamp: entry.timestamp,
    };

    if (this.reportConfig.beforeSend) {
      event = this.reportConfig.beforeSend(event);
      if (!event) return;
    }

    this.reportBuffer.push(event);
    this.scheduleFlush();
  }

  private scheduleFlush(): void {
    if (this.reportTimer !== null) return;
    this.reportTimer = setTimeout(() => {
      this.reportTimer = null;
      void this.flush();
    }, 1000);
  }

  private async sendBatch(events: ErrorEvent[]): Promise<void> {
    if (!this.reportConfig) return;

    try {
      await fetch(this.reportConfig.endpoint, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ errors: events }),
        keepalive: true,
      });
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : String(err);
      console.warn(`[ErrorTracker] Report failed: ${message}`);
    }
  }

  // ---------------------------------------------------------------------------
  // Private — global handlers
  // ---------------------------------------------------------------------------

  private installGlobalHandlers(): void {
    if (
      typeof globalThis === "undefined" ||
      typeof (globalThis as unknown as { window?: unknown }).window === "undefined"
    ) {
      return;
    }

    const win = globalThis as unknown as Window;

    this.boundOnError = ((event: globalThis.ErrorEvent) => {
      if (event.error instanceof Error) {
        this.captureException(event.error);
      } else {
        this.captureException(new Error(event.message ?? "Unknown error"));
      }
    }) as unknown as (event: ErrorEvent) => void;

    this.boundOnRejection = (event: PromiseRejectionEvent) => {
      const reason: unknown = event.reason;
      if (reason instanceof Error) {
        this.captureException(reason);
      } else {
        this.captureException(new Error(String(reason)));
      }
    };

    win.addEventListener(
      "error",
      this.boundOnError as unknown as EventListenerOrEventListenerObject,
    );
    win.addEventListener(
      "unhandledrejection",
      this.boundOnRejection as unknown as EventListenerOrEventListenerObject,
    );
  }
}
