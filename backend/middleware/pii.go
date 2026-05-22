package middleware

import (
	"bytes"
	"net/http"

	"github.com/EthanShen10086/voxera-kit/pii"
)

// PIIRedact returns a [Func] that scrubs personally identifiable information
// from error response bodies (4xx and 5xx). Non-error responses pass through
// unmodified.
func PIIRedact(redactor pii.Redactor) Func {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rec := &bufferedResponseWriter{
				header: w.Header(),
				buf:    &bytes.Buffer{},
				status: http.StatusOK,
			}

			next.ServeHTTP(rec, r)

			if rec.status >= 400 {
				redacted := redactor.Redact(rec.buf.String())
				w.WriteHeader(rec.status)
				_, _ = w.Write([]byte(redacted))
				return
			}

			w.WriteHeader(rec.status)
			_, _ = w.Write(rec.buf.Bytes())
		})
	}
}

// bufferedResponseWriter captures the response so that error bodies can be
// post-processed.
type bufferedResponseWriter struct {
	header http.Header
	buf    *bytes.Buffer
	status int
}

// Header returns the response header map.
func (b *bufferedResponseWriter) Header() http.Header { return b.header }

// WriteHeader captures the status code.
func (b *bufferedResponseWriter) WriteHeader(code int) { b.status = code }

// Write appends data to the internal buffer.
func (b *bufferedResponseWriter) Write(p []byte) (int, error) { return b.buf.Write(p) }
