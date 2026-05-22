package middleware

import (
	"net/http"

	"github.com/EthanShen10086/voxera-kit/security/headers"
)

// SecurityHeaders returns a [Func] that sets standard HTTP security headers
// on every response according to the supplied [headers.Config].
func SecurityHeaders(cfg headers.Config) Func {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h := w.Header()
			if v := cfg.HSTS.String(); cfg.HSTS.MaxAge > 0 {
				h.Set("Strict-Transport-Security", v)
			}
			if cfg.CSP != "" {
				h.Set("Content-Security-Policy", cfg.CSP)
			}
			if cfg.XFrameOptions != "" {
				h.Set("X-Frame-Options", cfg.XFrameOptions)
			}
			if cfg.XContentTypeOptions != "" {
				h.Set("X-Content-Type-Options", cfg.XContentTypeOptions)
			}
			if cfg.ReferrerPolicy != "" {
				h.Set("Referrer-Policy", cfg.ReferrerPolicy)
			}
			if cfg.PermissionsPolicy != "" {
				h.Set("Permissions-Policy", cfg.PermissionsPolicy)
			}
			next.ServeHTTP(w, r)
		})
	}
}
