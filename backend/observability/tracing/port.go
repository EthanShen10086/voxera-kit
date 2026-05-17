// Package tracing defines the distributed tracing port used to propagate
// request context across service boundaries.
package tracing

import "context"

// SpanKind describes the relationship of a span to its parent.
type SpanKind int

const (
	// SpanKindInternal is the default span kind.
	SpanKindInternal SpanKind = iota
	// SpanKindServer marks an inbound request span.
	SpanKindServer
	// SpanKindClient marks an outbound request span.
	SpanKindClient
	// SpanKindProducer marks a message-producing span.
	SpanKindProducer
	// SpanKindConsumer marks a message-consuming span.
	SpanKindConsumer
)

// SpanOption configures optional span behaviour at creation time.
type SpanOption struct {
	Kind       SpanKind
	Attributes map[string]any
}

// SpanContext carries the identifiers needed to correlate spans across processes.
type SpanContext struct {
	TraceID string
	SpanID  string
}

// Span represents a single unit of work inside a trace.
type Span interface {
	// End completes the span, recording its duration.
	End()

	// SetAttributes attaches a key-value pair to the span.
	SetAttributes(key string, value any)

	// RecordError marks the span as errored and records the error.
	RecordError(err error)

	// SpanContext returns the identifiers for this span.
	SpanContext() SpanContext
}

// Tracer is the primary port for creating and managing spans.
type Tracer interface {
	// Start begins a new span, returning a child context and the span handle.
	Start(ctx context.Context, name string, opts ...SpanOption) (context.Context, Span)

	// Shutdown flushes pending spans and releases resources.
	Shutdown(ctx context.Context) error
}
