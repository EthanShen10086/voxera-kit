package middleware

import (
	"net/http"

	"github.com/EthanShen10086/voxera-kit/loadshed"
)

// LoadShed returns a [Func] that rejects requests with 503 Service
// Unavailable when the [loadshed.Shedder] reports that the system is
// overloaded. A Retry-After header hints the client to back off.
func LoadShed(shedder loadshed.Shedder) Func {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, err := shedder.Allow()
			if err != nil {
				w.Header().Set("Retry-After", "1")
				http.Error(w, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
				return
			}

			sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(sw, r)
			token.Done(sw.status < 500)
		})
	}
}
