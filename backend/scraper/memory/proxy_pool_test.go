package memory_test

import (
	"context"
	"testing"

	"github.com/EthanShen10086/voxera-kit/scraper"
	"github.com/EthanShen10086/voxera-kit/scraper/memory"
)

func TestProxyPoolRoundRobin(t *testing.T) {
	p := memory.NewProxyPool()
	_ = p.Add(&scraper.Proxy{URL: "http://a"})
	_ = p.Add(&scraper.Proxy{URL: "http://b"})
	ctx := context.Background()
	first, err := p.Next(ctx)
	if err != nil || first.URL != "http://b" {
		t.Fatalf("first = %+v err=%v", first, err)
	}
	second, err := p.Next(ctx)
	if err != nil || second.URL != "http://a" {
		t.Fatalf("second = %+v err=%v", second, err)
	}
	if p.Size() != 2 {
		t.Fatalf("size = %d", p.Size())
	}
}

func TestProxyPoolRemove(t *testing.T) {
	p := memory.NewProxyPool()
	_ = p.Add(&scraper.Proxy{URL: "http://x"})
	if err := p.Remove("http://x"); err != nil {
		t.Fatal(err)
	}
	if _, err := p.Next(context.Background()); err == nil {
		t.Fatal("expected empty pool error")
	}
	if err := p.Remove("missing"); err == nil {
		t.Fatal("expected remove error")
	}
}

func TestProxyPoolHealthCheck(t *testing.T) {
	p := memory.NewProxyPool()
	if err := p.HealthCheck(context.Background()); err != nil {
		t.Fatal(err)
	}
}
