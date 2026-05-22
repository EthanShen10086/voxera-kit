// Package tracing — OpenTelemetry adapter.
//
// OTelTracer implements [Tracer] by delegating to the OpenTelemetry SDK.
// Import path for the real dependency: go.opentelemetry.io/otel
package tracing

import "context"

// OTelTracer implements [Tracer] using the OpenTelemetry Go SDK.
type OTelTracer struct {
	// tracer trace.Tracer  // uncomment when OTel SDK is in go.mod
}

// NewOTelTracer creates an [OTelTracer]. The caller is responsible for
// configuring the global TracerProvider before calling this constructor.
func NewOTelTracer() *OTelTracer {
	return &OTelTracer{}
}

// Start begins a new span, returning a child context and the span handle.
func (o *OTelTracer) Start(ctx context.Context, name string, opts ...SpanOption) (context.Context, Span) {
	// TODO: delegate to o.tracer.Start, wrap the OTel span
	return ctx, &otelSpan{}
}

// Shutdown flushes pending spans and releases resources.
func (o *OTelTracer) Shutdown(ctx context.Context) error {
	// TODO: flush and shut down the TracerProvider
	return nil
}

var _ Tracer = (*OTelTracer)(nil)

// otelSpan wraps an OpenTelemetry span to satisfy the [Span] interface.
type otelSpan struct {
	// span trace.Span  // uncomment when OTel SDK is in go.mod
}

func (s *otelSpan) End() {
	// TODO: s.span.End()
}

func (s *otelSpan) SetAttributes(key string, value any) {
	// TODO: s.span.SetAttributes(attribute.String(key, fmt.Sprint(value)))
}

func (s *otelSpan) RecordError(err error) {
	// TODO: s.span.RecordError(err)
}

func (s *otelSpan) SpanContext() SpanContext {
	// TODO: extract trace/span IDs from OTel span context
	return SpanContext{}
}

var _ Span = (*otelSpan)(nil)
