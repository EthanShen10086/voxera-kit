package middleware

import (
	"context"
	"net/http"
	"time"
)

// Timeout returns a [Func] that enforces a per-request deadline. If the
// handler does not complete within duration d the client receives a
// 503 Service Unavailable response.
func Timeout(d time.Duration) Func {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), d)
			defer cancel()

			done := make(chan struct{})
			sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}

			go func() {
				next.ServeHTTP(sw, r.WithContext(ctx))
				close(done)
			}()

			select {
			case <-done:
				// Handler completed in time.
			case <-ctx.Done():
				http.Error(w, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
			}
		})
	}
}
