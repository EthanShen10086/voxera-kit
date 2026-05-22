package middleware

import (
	"net/http"
	"runtime/debug"

	"github.com/EthanShen10086/voxera-kit/observability/logger"
)

// Recovery returns a [Func] that catches panics raised during request
// handling, logs the stack trace, and responds with 500 Internal Server Error
// instead of crashing the process.
func Recovery(log logger.Logger) Func {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					log.Error("panic recovered",
						logger.Field{Key: "error", Value: rec},
						logger.Field{Key: "stack", Value: string(debug.Stack())},
						logger.Field{Key: "method", Value: r.Method},
						logger.Field{Key: "path", Value: r.URL.Path},
						logger.Field{Key: "request_id", Value: RequestIDFromContext(r.Context())},
					)
					http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
