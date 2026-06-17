// Minimal gateway wiring example — mirrors MsgGuard assembly pattern.
// Build: go run ./examples/minimal-gateway (from repo root with go.work)
package main

import (
	"net/http"

	"github.com/EthanShen10086/voxera-kit/circuitbreaker"
	cbmem "github.com/EthanShen10086/voxera-kit/circuitbreaker/memory"
	"github.com/EthanShen10086/voxera-kit/middleware"
)

func main() {
	cb := cbmem.New(circuitbreaker.Config{MaxFailures: 5})
	_ = cb // use in handlers

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	h := middleware.Chain(mux, middleware.RequestID())
	http.ListenAndServe(":8080", h)
}
