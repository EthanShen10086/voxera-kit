import type { SpanContext, TracingConfig } from "./types.js";

function generateId(length: number): string {
  const chars = "0123456789abcdef";
  let id = "";
  for (let i = 0; i < length; i++) {
    id += chars[Math.floor(Math.random() * chars.length)];
  }
  return id;
}

/** Configuration for exporting spans to an OTLP-compatible collector. */
export interface ExportConfig {
  /** Full URL of the OTLP/HTTP collector (e.g. `https://otel.example.com/v1/traces`). */
  collectorUrl: string;
  /** Number of completed spans to buffer before flushing. Defaults to 20. */
  batchSize?: number;
  /** Interval in ms between automatic flushes. Defaults to 5000. */
  flushInterval?: number;
  /** Additional HTTP headers sent with every export request. */
  headers?: Record<string, string>;
}

/** OTLP JSON attribute value representation. */
interface OtlpAttributeValue {
  stringValue?: string;
  intValue?: number;
  boolValue?: boolean;
}

/** OTLP JSON attribute key/value pair. */
interface OtlpAttribute {
  key: string;
  value: OtlpAttributeValue;
}

/** OTLP JSON span representation. */
interface OtlpSpan {
  traceId: string;
  spanId: string;
  parentSpanId?: string;
  name: string;
  kind: number;
  startTimeUnixNano: string;
  endTimeUnixNano: string;
  attributes: OtlpAttribute[];
}

/**
 * Distributed tracing client that generates W3C Trace Context headers and
 * optionally exports completed spans to an OTLP/HTTP collector in batched
 * JSON format.
 */
export class TracingClient {
  private readonly config: TracingConfig;
  private readonly activeSpans = new Map<string, SpanContext>();

  private readonly exportConfig?: ExportConfig;
  private readonly spanBuffer: SpanContext[] = [];
  private flushTimer: ReturnType<typeof setInterval> | null = null;

  constructor(config: TracingConfig, exportConfig?: ExportConfig) {
    this.config = {
      sampleRate: 1.0,
      propagateContextHeaders: true,
      ...config,
    };

    if (exportConfig) {
      this.exportConfig = {
        batchSize: 20,
        flushInterval: 5000,
        ...exportConfig,
      };

      if (this.exportConfig.flushInterval! > 0) {
        this.flushTimer = setInterval(() => {
          void this.flush();
        }, this.exportConfig.flushInterval);
      }
    }
  }

  /**
   * Start a new span. Optionally link it to a parent span by providing
   * {@link options.parentSpanId}.
   */
  startSpan(
    name: string,
    options?: {
      parentSpanId?: string;
      attributes?: Record<string, string | number | boolean>;
    },
  ): SpanContext {
    const span: SpanContext = {
      traceId: options?.parentSpanId
        ? (this.activeSpans.get(options.parentSpanId)?.traceId ?? generateId(32))
        : generateId(32),
      spanId: generateId(16),
      parentSpanId: options?.parentSpanId,
      name,
      startTime: Date.now(),
      attributes: options?.attributes ?? {},
    };

    this.activeSpans.set(span.spanId, span);
    return span;
  }

  /**
   * End the span identified by {@link spanId} and enqueue it for export.
   * Returns the finalized span or `undefined` if it was not found.
   */
  endSpan(spanId: string): SpanContext | undefined {
    const span = this.activeSpans.get(spanId);
    if (!span) return undefined;

    this.activeSpans.delete(spanId);
    span.attributes["duration_ms"] = Date.now() - span.startTime;

    if (this.shouldSample()) {
      this.exportSpan(span);
    }

    return span;
  }

  /** Merge additional attributes into an active span. */
  setAttributes(
    spanId: string,
    attributes: Record<string, string | number | boolean>,
  ): void {
    const span = this.activeSpans.get(spanId);
    if (span) {
      Object.assign(span.attributes, attributes);
    }
  }

  /**
   * Build propagation headers (W3C Trace Context format) for outgoing HTTP
   * requests so downstream services can continue the trace.
   */
  getContextHeaders(spanId: string): Record<string, string> {
    const span = this.activeSpans.get(spanId);
    if (!span || !this.config.propagateContextHeaders) return {};

    return {
      traceparent: `00-${span.traceId}-${span.spanId}-01`,
    };
  }

  /** Return the most recently started active span, or `null`. */
  activeSpan(): SpanContext | null {
    let latest: SpanContext | null = null;
    for (const span of this.activeSpans.values()) {
      if (!latest || span.startTime > latest.startTime) {
        latest = span;
      }
    }
    return latest;
  }

  /** Flush all buffered spans to the collector immediately. */
  async flush(): Promise<void> {
    if (this.spanBuffer.length === 0 || !this.exportConfig) return;

    const batch = this.spanBuffer.splice(0, this.spanBuffer.length);
    await this.sendBatch(batch);
  }

  /** Stop the periodic flush timer and flush remaining spans. */
  async dispose(): Promise<void> {
    if (this.flushTimer !== null) {
      clearInterval(this.flushTimer);
      this.flushTimer = null;
    }
    await this.flush();
  }

  // ---------------------------------------------------------------------------
  // Private
  // ---------------------------------------------------------------------------

  private shouldSample(): boolean {
    return Math.random() < (this.config.sampleRate ?? 1.0);
  }

  private exportSpan(span: SpanContext): void {
    if (!this.exportConfig) return;

    this.spanBuffer.push(span);

    if (this.spanBuffer.length >= (this.exportConfig.batchSize ?? 20)) {
      void this.flush();
    }
  }

  private async sendBatch(spans: SpanContext[]): Promise<void> {
    if (!this.exportConfig) return;

    const otlpSpans: OtlpSpan[] = spans.map((s) => this.toOtlpSpan(s));

    const payload = {
      resourceSpans: [
        {
          resource: {
            attributes: [
              {
                key: "service.name",
                value: { stringValue: this.config.serviceName },
              },
            ],
          },
          scopeSpans: [
            {
              spans: otlpSpans,
            },
          ],
        },
      ],
    };

    try {
      const headers: Record<string, string> = {
        "Content-Type": "application/json",
        ...this.exportConfig.headers,
      };

      await fetch(this.exportConfig.collectorUrl, {
        method: "POST",
        headers,
        body: JSON.stringify(payload),
        keepalive: true,
      });
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : String(err);
      console.warn(`[TracingClient] Export failed: ${message}`);
    }
  }

  private toOtlpSpan(span: SpanContext): OtlpSpan {
    const durationMs =
      typeof span.attributes["duration_ms"] === "number"
        ? span.attributes["duration_ms"]
        : 0;

    const startNano = BigInt(span.startTime) * BigInt(1_000_000);
    const endNano = startNano + BigInt(Math.round(durationMs)) * BigInt(1_000_000);

    const attributes: OtlpAttribute[] = [];
    for (const [key, val] of Object.entries(span.attributes)) {
      if (key === "duration_ms") continue;
      attributes.push(this.toOtlpAttribute(key, val));
    }

    const result: OtlpSpan = {
      traceId: span.traceId,
      spanId: span.spanId,
      name: span.name,
      kind: 1, // SPAN_KIND_INTERNAL
      startTimeUnixNano: startNano.toString(),
      endTimeUnixNano: endNano.toString(),
      attributes,
    };

    if (span.parentSpanId) {
      result.parentSpanId = span.parentSpanId;
    }

    return result;
  }

  private toOtlpAttribute(
    key: string,
    value: string | number | boolean,
  ): OtlpAttribute {
    if (typeof value === "string") {
      return { key, value: { stringValue: value } };
    }
    if (typeof value === "boolean") {
      return { key, value: { boolValue: value } };
    }
    return { key, value: { intValue: value } };
  }
}
