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

// Start begins accepting HTTP connections via Kratos.
func (h *HTTPAdapter) Start(ctx context.Context) error {
	// TODO: build kratos http.Server, apply routes and middlewares, start
	return nil
}

// Stop performs a graceful shutdown of the Kratos HTTP server.
func (h *HTTPAdapter) Stop(ctx context.Context) error {
	// TODO: graceful shutdown of kratos http.Server
	return nil
}

// Use appends global middlewares to the HTTP server.
func (h *HTTPAdapter) Use(middlewares ...framework.Middleware) {
	h.middlewares = append(h.middlewares, middlewares...)
}

// Handle registers a handler for the given HTTP method and path.
func (h *HTTPAdapter) Handle(method, path string, handler framework.HandlerFunc) {
	h.routes = append(h.routes, framework.Route{
		Method:  method,
		Path:    path,
		Handler: handler,
	})
}

// Group returns a sub-router scoped to the given path prefix.
func (h *HTTPAdapter) Group(prefix string) framework.HTTPServer {
	// TODO: return a sub-router scoped to prefix
	return h
}

// Routes returns all registered routes.
func (h *HTTPAdapter) Routes() []framework.Route {
	return h.routes
}

var _ framework.HTTPServer = (*HTTPAdapter)(nil)
