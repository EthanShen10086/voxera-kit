package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/EthanShen10086/voxera-kit/observability/metrics"
)

// Metrics returns a [Func] that records RED (Rate, Errors, Duration) metrics
// for every HTTP request using the supplied [metrics.Recorder].
//
// Recorded metrics:
//   - http_requests_total — counter incremented once per request
//   - http_request_duration_seconds — histogram of request latency
//   - http_requests_in_flight — gauge of concurrently active requests
//
// All metrics are tagged with method, path, and status_code.
func Metrics(recorder metrics.Recorder) Func {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tags := map[string]string{
				"method": r.Method,
				"path":   r.URL.Path,
			}

			recorder.Gauge("http_requests_in_flight", 1, tags)
			start := time.Now()
			sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}

			next.ServeHTTP(sw, r)

			elapsed := time.Since(start).Seconds()
			tags["status_code"] = strconv.Itoa(sw.status)

			recorder.Counter("http_requests_total", 1, tags)
			recorder.Histogram("http_request_duration_seconds", elapsed, tags)
			recorder.Gauge("http_requests_in_flight", -1, map[string]string{
				"method": r.Method,
				"path":   r.URL.Path,
			})
		})
	}
}
