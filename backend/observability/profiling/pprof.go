// Package profiling provides helpers for registering Go runtime profiling
// endpoints on an HTTP mux.
package profiling

import (
	"net/http"
	"net/http/pprof"
)

// RegisterPprof registers the standard /debug/pprof/* routes on the given
// [http.ServeMux] when enabled is true. When enabled is false the call is a
// no-op, making it safe to wire unconditionally and gate via configuration.
func RegisterPprof(mux *http.ServeMux, enabled bool) {
	if !enabled {
		return
	}
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
}
