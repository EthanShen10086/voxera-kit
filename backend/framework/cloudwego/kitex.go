// Package cloudwego provides CloudWeGo-backed implementations of the
// [framework.HTTPServer] and [framework.RPCServer] ports.
package cloudwego

import (
	"context"

	"github.com/EthanShen10086/voxera-kit/framework"
)

// KitexAdapter implements [framework.RPCServer] using CloudWeGo Kitex.
//
// Intended dependency: github.com/cloudwego/kitex
type KitexAdapter struct {
	cfg          framework.ServerConfig
	middlewares  []framework.Middleware
	interceptors []any
}

// NewKitexAdapter creates a Kitex-based RPC server.
func NewKitexAdapter(cfg framework.ServerConfig) *KitexAdapter {
	return &KitexAdapter{cfg: cfg}
}

// Start begins accepting RPC connections via Kitex.
func (k *KitexAdapter) Start(ctx context.Context) error {
	// TODO: build Kitex server, register services, start
	return nil
}

// Stop performs a graceful shutdown of the Kitex server.
func (k *KitexAdapter) Stop(ctx context.Context) error {
	// TODO: graceful shutdown of Kitex server
	return nil
}

// Use appends global middlewares to the Kitex server.
func (k *KitexAdapter) Use(middlewares ...framework.Middleware) {
	k.middlewares = append(k.middlewares, middlewares...)
}

// RegisterService registers a Thrift/Protobuf service with the Kitex server.
func (k *KitexAdapter) RegisterService(desc any, impl any) {
	// TODO: register Thrift/Protobuf service with the Kitex server
}

// UseUnary appends unary interceptors to the Kitex RPC pipeline.
func (k *KitexAdapter) UseUnary(interceptors ...any) {
	k.interceptors = append(k.interceptors, interceptors...)
}

var _ framework.RPCServer = (*KitexAdapter)(nil)
