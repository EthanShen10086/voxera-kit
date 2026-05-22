// Package tracing — OpenTelemetry adapter.
//
// OTelTracer implements [Tracer] by delegating to the OpenTelemetry SDK with
// an OTLP HTTP exporter.
package tracing

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
)

// OTelConfig holds the configuration needed to bootstrap an OTLP HTTP exporter.
type OTelConfig struct {
	// ServiceName identifies the service in traces.
	ServiceName string
	// Endpoint is the OTLP collector address (e.g. "localhost:4318").
	Endpoint string
	// SampleRate controls the fraction of traces sampled (0.0–1.0).
	SampleRate float64
	// Insecure disables TLS when true.
	Insecure bool
	// Headers are sent with every OTLP export request.
	Headers map[string]string
}

// OTelTracer implements [Tracer] using the OpenTelemetry Go SDK.
type OTelTracer struct {
	provider *sdktrace.TracerProvider
	tracer   trace.Tracer
}

// NewOTelTracer creates an [OTelTracer] backed by an OTLP HTTP exporter.
//
// It configures a [sdktrace.TracerProvider] with a [sdktrace.BatchSpanProcessor],
// sets the global provider, and returns a ready-to-use tracer. Call [OTelTracer.Shutdown]
// to flush pending spans and release resources.
func NewOTelTracer(ctx context.Context, cfg OTelConfig) (*OTelTracer, error) {
	opts := []otlptracehttp.Option{
		otlptracehttp.WithEndpoint(cfg.Endpoint),
	}
	if cfg.Insecure {
		opts = append(opts, otlptracehttp.WithInsecure())
	}
	if len(cfg.Headers) > 0 {
		opts = append(opts, otlptracehttp.WithHeaders(cfg.Headers))
	}

	exporter, err := otlptracehttp.New(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("create OTLP exporter: %w", err)
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(semconv.ServiceName(cfg.ServiceName)),
	)
	if err != nil {
		return nil, fmt.Errorf("create resource: %w", err)
	}

	sampler := sdktrace.AlwaysSample()
	if cfg.SampleRate > 0 && cfg.SampleRate < 1 {
		sampler = sdktrace.TraceIDRatioBased(cfg.SampleRate)
	} else if cfg.SampleRate == 0 {
		sampler = sdktrace.NeverSample()
	}

	provider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sampler),
	)
	otel.SetTracerProvider(provider)

	return &OTelTracer{
		provider: provider,
		tracer:   provider.Tracer(cfg.ServiceName),
	}, nil
}

func toOTelSpanKind(k SpanKind) trace.SpanKind {
	switch k {
	case SpanKindServer:
		return trace.SpanKindServer
	case SpanKindClient:
		return trace.SpanKindClient
	case SpanKindProducer:
		return trace.SpanKindProducer
	case SpanKindConsumer:
		return trace.SpanKindConsumer
	default:
		return trace.SpanKindInternal
	}
}

// Start begins a new span, returning a child context and the span handle.
func (o *OTelTracer) Start(ctx context.Context, name string, opts ...SpanOption) (context.Context, Span) {
	var startOpts []trace.SpanStartOption
	for _, opt := range opts {
		startOpts = append(startOpts, trace.WithSpanKind(toOTelSpanKind(opt.Kind)))
		if len(opt.Attributes) > 0 {
			var attrs []attribute.KeyValue
			for k, v := range opt.Attributes {
				attrs = append(attrs, attribute.String(k, fmt.Sprint(v)))
			}
			startOpts = append(startOpts, trace.WithAttributes(attrs...))
		}
	}

	ctx, span := o.tracer.Start(ctx, name, startOpts...)
	return ctx, &otelSpan{span: span}
}

// Shutdown flushes pending spans and releases resources.
func (o *OTelTracer) Shutdown(ctx context.Context) error {
	return o.provider.Shutdown(ctx)
}

var _ Tracer = (*OTelTracer)(nil)

// otelSpan wraps an OpenTelemetry span to satisfy the [Span] interface.
type otelSpan struct {
	span trace.Span
}

// End completes the span, recording its duration.
func (s *otelSpan) End() {
	s.span.End()
}

// SetAttributes attaches a key-value pair to the span.
func (s *otelSpan) SetAttributes(key string, value any) {
	s.span.SetAttributes(attribute.String(key, fmt.Sprint(value)))
}

// RecordError marks the span as errored and records the error.
func (s *otelSpan) RecordError(err error) {
	s.span.RecordError(err)
}

// SpanContext returns the identifiers for this span.
func (s *otelSpan) SpanContext() SpanContext {
	sc := s.span.SpanContext()
	return SpanContext{
		TraceID: sc.TraceID().String(),
		SpanID:  sc.SpanID().String(),
	}
}

var _ Span = (*otelSpan)(nil)
