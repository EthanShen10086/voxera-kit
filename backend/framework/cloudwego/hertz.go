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

func (h *HertzAdapter) Start(ctx context.Context) error {
	// TODO: build hertz server, register routes, start
	return nil
}

func (h *HertzAdapter) Stop(ctx context.Context) error {
	// TODO: graceful shutdown of hertz server
	return nil
}

func (h *HertzAdapter) Use(middlewares ...framework.Middleware) {
	h.middlewares = append(h.middlewares, middlewares...)
}

func (h *HertzAdapter) Handle(method, path string, handler framework.HandlerFunc) {
	h.routes = append(h.routes, framework.Route{
		Method:  method,
		Path:    path,
		Handler: handler,
	})
}

func (h *HertzAdapter) Group(prefix string) framework.HTTPServer {
	// TODO: return a sub-router scoped to prefix
	return h
}

func (h *HertzAdapter) Routes() []framework.Route {
	return h.routes
}

var _ framework.HTTPServer = (*HertzAdapter)(nil)
