package middleware

import (
	"context"
	"fmt"
	"net/http"

	"github.com/EthanShen10086/voxera-kit/observability/tracing"
)

const traceIDHeader = "X-Trace-ID"

// Tracing returns a [Func] that starts an OpenTelemetry-style server span for
// every request. The trace ID is propagated via the X-Trace-ID response header
// and stored in the request context (retrievable via [TraceIDFromContext]).
func Tracing(tracer tracing.Tracer) Func {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			spanName := fmt.Sprintf("%s %s", r.Method, r.URL.Path)
			ctx, span := tracer.Start(r.Context(), spanName, tracing.SpanOption{
				Kind: tracing.SpanKindServer,
				Attributes: map[string]any{
					"http.method": r.Method,
					"http.url":    r.URL.String(),
				},
			})
			defer span.End()

			traceID := span.SpanContext().TraceID
			ctx = context.WithValue(ctx, CtxTraceID, traceID)
			w.Header().Set(traceIDHeader, traceID)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
