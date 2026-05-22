// Package kratos provides Go-Kratos-backed implementations of the
// [framework.HTTPServer] and [framework.RPCServer] ports.
package kratos

import (
	"context"

	"github.com/EthanShen10086/voxera-kit/framework"
)

// GRPCAdapter implements [framework.RPCServer] using Go-Kratos gRPC transport.
type GRPCAdapter struct {
	cfg          framework.ServerConfig
	middlewares  []framework.Middleware
	interceptors []any
}

// NewGRPCAdapter creates a Kratos-based gRPC server.
func NewGRPCAdapter(cfg framework.ServerConfig) *GRPCAdapter {
	return &GRPCAdapter{cfg: cfg}
}

// Start begins accepting gRPC connections via Kratos.
func (g *GRPCAdapter) Start(ctx context.Context) error {
	// TODO: build kratos grpc.Server, register services, start
	return nil
}

// Stop performs a graceful shutdown of the Kratos gRPC server.
func (g *GRPCAdapter) Stop(ctx context.Context) error {
	// TODO: graceful shutdown of kratos grpc.Server
	return nil
}

// Use appends global middlewares to the gRPC server.
func (g *GRPCAdapter) Use(middlewares ...framework.Middleware) {
	g.middlewares = append(g.middlewares, middlewares...)
}

// RegisterService registers a protobuf service with the gRPC server.
func (g *GRPCAdapter) RegisterService(desc any, impl any) {
	// TODO: register protobuf service with the gRPC server
}

// UseUnary appends unary interceptors to the gRPC pipeline.
func (g *GRPCAdapter) UseUnary(interceptors ...any) {
	g.interceptors = append(g.interceptors, interceptors...)
}

var _ framework.RPCServer = (*GRPCAdapter)(nil)
