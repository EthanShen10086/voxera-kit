import type { SpanContext, TracingConfig } from "./types.js";

function generateId(length: number): string {
  const chars = "0123456789abcdef";
  let id = "";
  for (let i = 0; i < length; i++) {
    id += chars[Math.floor(Math.random() * chars.length)];
  }
  return id;
}

export class TracingClient {
  private readonly config: TracingConfig;
  private readonly activeSpans = new Map<string, SpanContext>();

  constructor(config: TracingConfig) {
    this.config = {
      sampleRate: 1.0,
      propagateContextHeaders: true,
      ...config,
    };
  }

  startSpan(name: string, options?: { parentSpanId?: string; attributes?: Record<string, string | number | boolean> }): SpanContext {
    const span: SpanContext = {
      traceId: options?.parentSpanId
        ? this.activeSpans.get(options.parentSpanId)?.traceId ?? generateId(32)
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

  setAttributes(spanId: string, attributes: Record<string, string | number | boolean>): void {
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

  private shouldSample(): boolean {
    return Math.random() < (this.config.sampleRate ?? 1.0);
  }

  activeSpan(): SpanContext | null {
    let latest: SpanContext | null = null;
    for (const span of this.activeSpans.values()) {
      if (!latest || span.startTime > latest.startTime) {
        latest = span;
      }
    }
    return latest;
  }

  /** Stub – in production this would send spans to the configured endpoint. */
  private exportSpan(_span: SpanContext): void {
    // No-op: integrate with an OTLP exporter in production.
  }
}
