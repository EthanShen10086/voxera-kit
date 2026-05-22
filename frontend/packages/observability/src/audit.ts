/** A single auditable user action. */
export interface AuditAction {
  action: string;
  resource?: string;
  resourceId?: string;
  metadata?: Record<string, unknown>;
  timestamp: number;
}

/** Configuration for the {@link AuditClient}. */
export interface AuditClientConfig {
  /** URL to POST audit action batches to. */
  endpoint: string;
  /** Number of buffered actions that trigger an automatic flush. Defaults to 25. */
  batchSize?: number;
  /** Interval in ms between automatic flushes. Defaults to 10 000. */
  flushInterval?: number;
  /** User ID attached to every action. Can be set later via {@link AuditClient.identify}. */
  userId?: string;
  /** Session ID attached to every action. Can be set later via {@link AuditClient.identify}. */
  sessionId?: string;
}

/**
 * Buffers user-facing audit actions and flushes them in batches to a
 * remote endpoint. Automatically flushes when the page is hidden
 * (`visibilitychange`) and on a configurable interval.
 */
export class AuditClient {
  private readonly config: Required<
    Pick<AuditClientConfig, "endpoint" | "batchSize" | "flushInterval">
  >;

  private userId?: string;
  private sessionId?: string;

  private readonly buffer: AuditAction[] = [];
  private flushTimer: ReturnType<typeof setInterval> | null = null;
  private boundVisibilityHandler: (() => void) | null = null;

  constructor(config: AuditClientConfig) {
    this.config = {
      endpoint: config.endpoint,
      batchSize: config.batchSize ?? 25,
      flushInterval: config.flushInterval ?? 10_000,
    };

    this.userId = config.userId;
    this.sessionId = config.sessionId;

    if (this.config.flushInterval > 0) {
      this.flushTimer = setInterval(() => {
        void this.flush();
      }, this.config.flushInterval);
    }

    if (
      typeof globalThis !== "undefined" &&
      typeof (globalThis as unknown as { document?: unknown }).document !== "undefined"
    ) {
      this.boundVisibilityHandler = () => {
        if (
          (globalThis as unknown as { document: { visibilityState: string } }).document
            .visibilityState === "hidden"
        ) {
          void this.flush();
        }
      };

      (globalThis as unknown as Window).document.addEventListener(
        "visibilitychange",
        this.boundVisibilityHandler,
      );
    }
  }

  /**
   * Buffer an audit action for later flushing. Triggers an immediate flush
   * when the buffer reaches the configured batch size.
   */
  track(
    action: string,
    resource?: string,
    resourceId?: string,
    metadata?: Record<string, unknown>,
  ): void {
    this.buffer.push({
      action,
      resource,
      resourceId,
      metadata,
      timestamp: Date.now(),
    });

    if (this.buffer.length >= this.config.batchSize) {
      void this.flush();
    }
  }

  /**
   * Associate a user and optional session with all subsequent (and buffered)
   * audit actions.
   */
  identify(userId: string, sessionId?: string): void {
    this.userId = userId;
    this.sessionId = sessionId;
  }

  /** Flush all buffered actions to the remote endpoint. */
  async flush(): Promise<void> {
    if (this.buffer.length === 0) return;

    const batch = this.buffer.splice(0, this.buffer.length);

    try {
      await fetch(this.config.endpoint, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          userId: this.userId,
          sessionId: this.sessionId,
          actions: batch,
        }),
        keepalive: true,
      });
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : String(err);
      console.warn(`[AuditClient] Flush failed: ${message}`);
    }
  }

  /** Stop the periodic timer and remove the visibility listener. */
  dispose(): void {
    if (this.flushTimer !== null) {
      clearInterval(this.flushTimer);
      this.flushTimer = null;
    }

    if (
      this.boundVisibilityHandler &&
      typeof globalThis !== "undefined" &&
      typeof (globalThis as unknown as { document?: unknown }).document !== "undefined"
    ) {
      (globalThis as unknown as Window).document.removeEventListener(
        "visibilitychange",
        this.boundVisibilityHandler,
      );
    }
  }
}
