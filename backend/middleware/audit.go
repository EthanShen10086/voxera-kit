package middleware

import (
	"bytes"
	"io"
	"net/http"
	"time"

	"github.com/EthanShen10086/voxera-kit/audit"
	"github.com/EthanShen10086/voxera-kit/observability/logger"
)

const maxAuditBodySize = 1 << 16 // 64 KiB

// isMutating reports whether the HTTP method changes server state.
func isMutating(method string) bool {
	switch method {
	case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		return true
	}
	return false
}

// Audit returns a [Func] that writes an audit entry for every mutating
// request (POST, PUT, PATCH, DELETE). The entry captures actor, action,
// resource, client IP, response status, and a size-limited copy of the
// request body.
func Audit(writer audit.Writer, log logger.Logger) Func {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !isMutating(r.Method) {
				next.ServeHTTP(w, r)
				return
			}

			var body []byte
			if r.Body != nil {
				lr := io.LimitReader(r.Body, maxAuditBodySize)
				var err error
				body, err = io.ReadAll(lr)
				if err != nil {
					log.Error("audit: failed to read request body",
						logger.Field{Key: "error", Value: err.Error()},
					)
				}
				r.Body = io.NopCloser(bytes.NewReader(body))
			}

			sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(sw, r)

			entry := audit.Entry{
				ActorID:        UserIDFromContext(r.Context()),
				TenantID:       TenantIDFromContext(r.Context()),
				Action:         r.Method,
				ResourceType:   r.URL.Path,
				IPAddress:      r.RemoteAddr,
				UserAgent:      r.UserAgent(),
				RequestBody:    body,
				ResponseStatus: sw.status,
				Timestamp:      time.Now(),
			}

			if err := writer.Write(r.Context(), entry); err != nil {
				log.Error("audit: failed to write entry",
					logger.Field{Key: "error", Value: err.Error()},
				)
			}
		})
	}
}
