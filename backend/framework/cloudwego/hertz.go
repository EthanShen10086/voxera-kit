// Package cloudwego provides CloudWeGo-backed implementations of the
// [framework.HTTPServer] and [framework.RPCServer] ports.
package cloudwego

import (
	"context"

	"github.com/EthanShen10086/voxera-kit/framework"
)

// HertzAdapter implements [framework.HTTPServer] using CloudWeGo Hertz.
type HertzAdapter struct {
	cfg         framework.ServerConfig
	middlewares []framework.Middleware
	routes      []framework.Route
}

// NewHertzAdapter creates a Hertz-based HTTP server.
func NewHertzAdapter(cfg framework.ServerConfig) *HertzAdapter {
	return &HertzAdapter{cfg: cfg}
}

// Start begins accepting HTTP connections via Hertz.
func (h *HertzAdapter) Start(ctx context.Context) error {
	// TODO: build hertz server, register routes, start
	return nil
}

// Stop performs a graceful shutdown of the Hertz server.
func (h *HertzAdapter) Stop(ctx context.Context) error {
	// TODO: graceful shutdown of hertz server
	return nil
}

// Use appends global middlewares to the Hertz server.
func (h *HertzAdapter) Use(middlewares ...framework.Middleware) {
	h.middlewares = append(h.middlewares, middlewares...)
}

// Handle registers a handler for the given HTTP method and path.
func (h *HertzAdapter) Handle(method, path string, handler framework.HandlerFunc) {
	h.routes = append(h.routes, framework.Route{
		Method:  method,
		Path:    path,
		Handler: handler,
	})
}

// Group returns a sub-router scoped to the given path prefix.
func (h *HertzAdapter) Group(prefix string) framework.HTTPServer {
	// TODO: return a sub-router scoped to prefix
	return h
}

// Routes returns all registered routes.
func (h *HertzAdapter) Routes() []framework.Route {
	return h.routes
}

var _ framework.HTTPServer = (*HertzAdapter)(nil)
