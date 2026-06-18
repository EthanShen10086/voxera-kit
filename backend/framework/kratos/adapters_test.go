package kratos_test

import (
	"context"
	"testing"

	"github.com/EthanShen10086/voxera-kit/framework"
	"github.com/EthanShen10086/voxera-kit/framework/kratos"
)

func TestHTTPAdapterLifecycle(t *testing.T) {
	cfg := framework.ServerConfig{Host: "127.0.0.1", Port: "8081"}
	h := kratos.NewHTTPAdapter(cfg)
	h.Use(func(next framework.HandlerFunc) framework.HandlerFunc { return next })
	h.Handle("POST", "/hook", func(ctx context.Context, _ any) (any, error) { return nil, nil })
	if len(h.Routes()) != 1 {
		t.Fatalf("routes = %#v", h.Routes())
	}
	if err := h.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := h.Stop(context.Background()); err != nil {
		t.Fatal(err)
	}
	_ = h.Group("/v1")
}

func TestGRPCAdapterLifecycle(t *testing.T) {
	cfg := framework.ServerConfig{Host: "127.0.0.1", Port: "9091"}
	g := kratos.NewGRPCAdapter(cfg)
	g.Use(func(next framework.HandlerFunc) framework.HandlerFunc { return next })
	g.RegisterService(nil, nil)
	g.UseUnary("unary")
	if err := g.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := g.Stop(context.Background()); err != nil {
		t.Fatal(err)
	}
}
