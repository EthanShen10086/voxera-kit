package middleware

import (
	"context"
	"encoding/json"
	"net/http"
)

// HealthChecker probes a single dependency for readiness.
type HealthChecker interface {
	// Check returns nil when the dependency is healthy and an error otherwise.
	Check(ctx context.Context) error
}

// healthResponse is the JSON body returned by the health endpoints.
type healthResponse struct {
	Status string            `json:"status"`
	Checks map[string]string `json:"checks,omitempty"`
}

// HealthCheck returns a [Func] that intercepts liveness and readiness probes.
//
//   - GET /health always returns 200 with {"status":"ok"} (liveness).
//   - GET /ready runs every registered [HealthChecker]; if all pass it returns
//     200 with {"status":"ok","checks":{…}}, otherwise 503 with
//     {"status":"degraded","checks":{…}}.
//
// All other paths are forwarded to the wrapped handler unchanged.
func HealthCheck(checks map[string]HealthChecker) Func {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch {
			case r.Method == http.MethodGet && r.URL.Path == "/health":
				writeLiveness(w)
			case r.Method == http.MethodGet && r.URL.Path == "/ready":
				writeReadiness(w, r.Context(), checks)
			default:
				next.ServeHTTP(w, r)
			}
		})
	}
}

func writeLiveness(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(healthResponse{Status: "ok"})
}

func writeReadiness(w http.ResponseWriter, ctx context.Context, checks map[string]HealthChecker) {
	results := make(map[string]string, len(checks))
	healthy := true
	for name, c := range checks {
		if err := c.Check(ctx); err != nil {
			results[name] = err.Error()
			healthy = false
		} else {
			results[name] = "ok"
		}
	}

	resp := healthResponse{Checks: results}
	status := http.StatusOK
	if healthy {
		resp.Status = "ok"
	} else {
		resp.Status = "degraded"
		status = http.StatusServiceUnavailable
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(resp)
}
