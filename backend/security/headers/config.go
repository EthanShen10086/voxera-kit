// Package headers provides configuration types and presets for HTTP security
// response headers such as HSTS, CSP, and X-Frame-Options.
package headers

import "fmt"

// Config holds the complete set of security headers to apply to HTTP responses.
type Config struct {
	HSTS                HSTSConfig
	CSP                 string
	XFrameOptions       string
	XContentTypeOptions string
	ReferrerPolicy      string
	PermissionsPolicy   string
}

// HSTSConfig holds the configuration for the Strict-Transport-Security header.
type HSTSConfig struct {
	MaxAge            int
	IncludeSubDomains bool
	Preload           bool
}

// String formats the HSTS configuration as a valid header value.
func (h HSTSConfig) String() string {
	val := fmt.Sprintf("max-age=%d", h.MaxAge)
	if h.IncludeSubDomains {
		val += "; includeSubDomains"
	}
	if h.Preload {
		val += "; preload"
	}
	return val
}

// DefaultStrict returns a Config with strict security headers suitable for
// production deployments.
func DefaultStrict() Config {
	return Config{
		HSTS: HSTSConfig{
			MaxAge:            63072000,
			IncludeSubDomains: true,
			Preload:           true,
		},
		CSP:                 "default-src 'self'; script-src 'self'; style-src 'self'; img-src 'self'; font-src 'self'; object-src 'none'; frame-ancestors 'none'; base-uri 'self'; form-action 'self'",
		XFrameOptions:       "DENY",
		XContentTypeOptions: "nosniff",
		ReferrerPolicy:      "strict-origin-when-cross-origin",
		PermissionsPolicy:   "camera=(), microphone=(), geolocation=(), payment=()",
	}
}

// DefaultPermissive returns a Config with relaxed security headers suitable for
// development environments.
func DefaultPermissive() Config {
	return Config{
		HSTS: HSTSConfig{
			MaxAge:            0,
			IncludeSubDomains: false,
			Preload:           false,
		},
		CSP:                 "",
		XFrameOptions:       "SAMEORIGIN",
		XContentTypeOptions: "nosniff",
		ReferrerPolicy:      "no-referrer-when-downgrade",
		PermissionsPolicy:   "",
	}
}
