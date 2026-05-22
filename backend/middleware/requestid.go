package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
)

const requestIDHeader = "X-Request-ID"

// generateID produces a 16-byte random hex string (32 characters).
func generateID() string {
	b := make([]byte, 16)
	// crypto/rand.Read always returns len(b) bytes on supported platforms;
	// a failure here indicates a broken OS entropy source, so we panic.
	if _, err := rand.Read(b); err != nil {
		panic("middleware: crypto/rand failed: " + err.Error())
	}
	return hex.EncodeToString(b)
}

// RequestID returns a [Func] that ensures every request carries a unique
// correlation ID. If the incoming request already has an X-Request-ID header
// that value is reused; otherwise a cryptographically random 16-byte hex ID
// is generated. The ID is stored in the request context (retrievable via
// [RequestIDFromContext]) and set on the response header.
func RequestID() Func {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id := r.Header.Get(requestIDHeader)
			if id == "" {
				id = generateID()
			}
			ctx := context.WithValue(r.Context(), CtxRequestID, id)
			w.Header().Set(requestIDHeader, id)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
