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

export interface ErrorTrackerOptions {
  /** Called when an exception or message is captured. Stub for real reporting (e.g. Sentry). */
  onCapture?: (entry: CapturedEntry) => void;
  /** Maximum number of breadcrumbs to retain. */
  maxBreadcrumbs?: number;
  /** Install global error handlers automatically. */
  installGlobalHandlers?: boolean;
}

export interface CapturedEntry {
  eventId: string;
  type: "exception" | "message";
  message: string;
  level: LogLevel;
  timestamp: number;
  error?: Error;
  context?: Record<string, unknown>;
  breadcrumbs: Breadcrumb[];
  user?: { id: string; email?: string };
  fingerprint: string;
}

export class ErrorTracker implements IErrorTracker {
  private readonly breadcrumbs: Breadcrumb[] = [];
  private readonly maxBreadcrumbs: number;
  private readonly seenFingerprints = new Set<string>();
  private user?: { id: string; email?: string };
  private readonly onCapture?: (entry: CapturedEntry) => void;

  private boundOnError?: (event: ErrorEvent) => void;
  private boundOnRejection?: (event: PromiseRejectionEvent) => void;

  constructor(options: ErrorTrackerOptions = {}) {
    this.maxBreadcrumbs = options.maxBreadcrumbs ?? MAX_BREADCRUMBS;
    this.onCapture = options.onCapture;

    if (options.installGlobalHandlers !== false) {
      this.installGlobalHandlers();
    }
  }

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

    return eventId;
  }

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
    return eventId;
  }

  setUser(user: { id: string; email?: string }): void {
    this.user = user;
  }

  addBreadcrumb(breadcrumb: Breadcrumb): void {
    this.breadcrumbs.push(breadcrumb);
    if (this.breadcrumbs.length > this.maxBreadcrumbs) {
      this.breadcrumbs.shift();
    }
  }

  /** Uninstall global handlers (useful in tests or cleanup). */
  dispose(): void {
    if (typeof globalThis === "undefined" || typeof (globalThis as any).window === "undefined") {
      return;
    }

    const win = globalThis as unknown as Window;
    if (this.boundOnError) {
      win.removeEventListener("error", this.boundOnError as any);
    }
    if (this.boundOnRejection) {
      win.removeEventListener("unhandledrejection", this.boundOnRejection as any);
    }
  }

  private installGlobalHandlers(): void {
    if (typeof globalThis === "undefined" || typeof (globalThis as any).window === "undefined") {
      return;
    }

    const win = globalThis as unknown as Window;

    this.boundOnError = (event: ErrorEvent) => {
      if (event.error instanceof Error) {
        this.captureException(event.error);
      } else {
        this.captureException(new Error(event.message ?? "Unknown error"));
      }
    };

    this.boundOnRejection = (event: PromiseRejectionEvent) => {
      const reason = event.reason;
      if (reason instanceof Error) {
        this.captureException(reason);
      } else {
        this.captureException(new Error(String(reason)));
      }
    };

    win.addEventListener("error", this.boundOnError as any);
    win.addEventListener("unhandledrejection", this.boundOnRejection as any);
  }
}
