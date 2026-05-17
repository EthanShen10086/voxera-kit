// Package kratos provides Go-Kratos-backed implementations of the
// [framework.HTTPServer] and [framework.RPCServer] ports.
package kratos

import (
	"context"

	"github.com/EthanShen10086/voxera-kit/framework"
)

// HTTPAdapter implements [framework.HTTPServer] using Go-Kratos HTTP transport.
type HTTPAdapter struct {
	cfg         framework.ServerConfig
	middlewares []framework.Middleware
	routes      []framework.Route
}

// NewHTTPAdapter creates a Kratos-based HTTP server.
func NewHTTPAdapter(cfg framework.ServerConfig) *HTTPAdapter {
	return &HTTPAdapter{cfg: cfg}
}

func (h *HTTPAdapter) Start(ctx context.Context) error {
	// TODO: build kratos http.Server, apply routes and middlewares, start
	return nil
}

func (h *HTTPAdapter) Stop(ctx context.Context) error {
	// TODO: graceful shutdown of kratos http.Server
	return nil
}

func (h *HTTPAdapter) Use(middlewares ...framework.Middleware) {
	h.middlewares = append(h.middlewares, middlewares...)
}

func (h *HTTPAdapter) Handle(method, path string, handler framework.HandlerFunc) {
	h.routes = append(h.routes, framework.Route{
		Method:  method,
		Path:    path,
		Handler: handler,
	})
}

func (h *HTTPAdapter) Group(prefix string) framework.HTTPServer {
	// TODO: return a sub-router scoped to prefix
	return h
}

func (h *HTTPAdapter) Routes() []framework.Route {
	return h.routes
}

var _ framework.HTTPServer = (*HTTPAdapter)(nil)
