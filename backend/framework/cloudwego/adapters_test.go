package cloudwego_test

import (
	"context"
	"testing"

	"github.com/EthanShen10086/voxera-kit/framework"
	"github.com/EthanShen10086/voxera-kit/framework/cloudwego"
)

func TestHertzAdapterLifecycle(t *testing.T) {
	cfg := framework.ServerConfig{Host: "127.0.0.1", Port: "8080"}
	h := cloudwego.NewHertzAdapter(cfg)
	h.Use(func(next framework.HandlerFunc) framework.HandlerFunc { return next })
	h.Handle("GET", "/ping", func(ctx context.Context, _ any) (any, error) { return nil, nil })
	routes := h.Routes()
	if len(routes) != 1 || routes[0].Path != "/ping" {
		t.Fatalf("routes = %#v", routes)
	}
	if err := h.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := h.Stop(context.Background()); err != nil {
		t.Fatal(err)
	}
	_ = h.Group("/api")
}

func TestKitexAdapterLifecycle(t *testing.T) {
	cfg := framework.ServerConfig{Host: "127.0.0.1", Port: "9090"}
	k := cloudwego.NewKitexAdapter(cfg)
	k.Use(func(next framework.HandlerFunc) framework.HandlerFunc { return next })
	k.RegisterService(nil, nil)
	k.UseUnary("interceptor")
	if err := k.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := k.Stop(context.Background()); err != nil {
		t.Fatal(err)
	}
}
