// Package framework defines the server abstraction ports for HTTP and RPC,
// allowing applications to swap between Go-Kratos, CloudWeGo, or other
// frameworks without changing business logic.
package framework

import (
	"context"
	"time"
)

// HandlerFunc is a transport-agnostic request handler.
type HandlerFunc func(ctx context.Context, req any) (any, error)

// Middleware wraps a [HandlerFunc] to add cross-cutting concerns
// (logging, auth, rate-limiting, etc.).
type Middleware func(next HandlerFunc) HandlerFunc

// Route describes a single registered HTTP endpoint.
type Route struct {
	Method      string
	Path        string
	Handler     HandlerFunc
	Middlewares []Middleware
}

// ServerConfig holds the network and TLS settings for any server type.
type ServerConfig struct {
	Host         string
	Port         string
	TLSCert      string
	TLSKey       string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

// Server is the minimal lifecycle interface every server adapter must satisfy.
type Server interface {
	// Start binds the server and begins accepting connections.
	// It blocks until ctx is canceled or an unrecoverable error occurs.
	Start(ctx context.Context) error

	// Stop performs a graceful shutdown, waiting for in-flight requests to finish.
	Stop(ctx context.Context) error

	// Use appends global middlewares that run on every request.
	Use(middlewares ...Middleware)
}

// HTTPServer extends [Server] with HTTP-specific routing capabilities.
type HTTPServer interface {
	Server

	// Handle registers a handler for the given HTTP method and path.
	Handle(method, path string, handler HandlerFunc)

	// Group returns a sub-router prefixed with the given path segment.
	Group(prefix string) HTTPServer

	// Routes returns all registered routes for introspection.
	Routes() []Route
}

// RPCServer extends [Server] with gRPC-style service registration.
type RPCServer interface {
	Server

	// RegisterService registers a protobuf service descriptor and its implementation.
	RegisterService(desc any, impl any)

	// UseUnary appends unary interceptors to the RPC pipeline.
	UseUnary(interceptors ...any)
}
