package middleware

import (
	"net/http"
	"time"

	"github.com/EthanShen10086/voxera-kit/observability/logger"
)

// Logging returns a [Func] that emits a structured log line for every
// completed request. It records method, path, status code, duration in
// milliseconds, client IP, request ID, and trace ID.
func Logging(log logger.Logger) Func {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}

			next.ServeHTTP(sw, r)

			log.Info("http request",
				logger.Field{Key: "method", Value: r.Method},
				logger.Field{Key: "path", Value: r.URL.Path},
				logger.Field{Key: "status", Value: sw.status},
				logger.Field{Key: "duration_ms", Value: time.Since(start).Milliseconds()},
				logger.Field{Key: "client_ip", Value: r.RemoteAddr},
				logger.Field{Key: "request_id", Value: RequestIDFromContext(r.Context())},
				logger.Field{Key: "trace_id", Value: TraceIDFromContext(r.Context())},
			)
		})
	}
}

// statusWriter wraps [http.ResponseWriter] to capture the status code.
type statusWriter struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

// WriteHeader captures the status code before delegating to the inner writer.
func (sw *statusWriter) WriteHeader(code int) {
	if !sw.wroteHeader {
		sw.status = code
		sw.wroteHeader = true
	}
	sw.ResponseWriter.WriteHeader(code)
}

// Write delegates to the inner writer and ensures the status is captured.
func (sw *statusWriter) Write(b []byte) (int, error) {
	if !sw.wroteHeader {
		sw.wroteHeader = true
	}
	return sw.ResponseWriter.Write(b)
}
