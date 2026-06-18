package memory_test

import (
	"context"
	"testing"

	"github.com/EthanShen10086/voxera-kit/registry"
	"github.com/EthanShen10086/voxera-kit/registry/memory"
)

func TestServiceRegistry(t *testing.T) {
	ctx := context.Background()
	a := memory.New()

	var notified int
	if err := a.Watch(ctx, "api", func([]*registry.ServiceInstance) { notified++ }); err != nil {
		t.Fatal(err)
	}

	inst := &registry.ServiceInstance{ID: "i1", Name: "api", Host: "127.0.0.1", Port: 8080}
	if err := a.Register(ctx, inst); err != nil {
		t.Fatal(err)
	}
	if err := a.Register(ctx, &registry.ServiceInstance{Name: "api"}); err == nil {
		t.Fatal("expected missing ID error")
	}

	found, err := a.Discover(ctx, "api")
	if err != nil || len(found) != 1 {
		t.Fatalf("Discover: %v %v", found, err)
	}
	if err := a.Heartbeat(ctx, "i1"); err != nil {
		t.Fatal(err)
	}
	if err := a.Deregister(ctx, "i1"); err != nil {
		t.Fatal(err)
	}
	if notified == 0 {
		t.Fatal("expected watch callback")
	}
}

func TestConfigCenter(t *testing.T) {
	ctx := context.Background()
	a := memory.New()

	var updates int
	if err := a.WatchConfig(ctx, "app.debug", func(*registry.ConfigValue) { updates++ }); err != nil {
		t.Fatal(err)
	}
	if err := a.Set(ctx, "app.debug", "true"); err != nil {
		t.Fatal(err)
	}
	if err := a.Set(ctx, "app.debug", "false"); err != nil {
		t.Fatal(err)
	}
	cv, err := a.Get(ctx, "app.debug")
	if err != nil || cv.Value != "false" || cv.Version != 2 {
		t.Fatalf("Get: %+v %v", cv, err)
	}
	list, err := a.List(ctx, "app.")
	if err != nil || len(list) != 1 {
		t.Fatalf("List: %v %v", list, err)
	}
	if err := a.Delete(ctx, "app.debug"); err != nil {
		t.Fatal(err)
	}
	if updates == 0 {
		t.Fatal("expected config watch callback")
	}
}

func TestConfigErrors(t *testing.T) {
	ctx := context.Background()
	a := memory.New()
	if _, err := a.Get(ctx, "missing"); err == nil {
		t.Fatal("expected not found")
	}
	if err := a.Delete(ctx, "missing"); err == nil {
		t.Fatal("expected not found")
	}
	if err := a.Heartbeat(ctx, "missing"); err == nil {
		t.Fatal("expected not found")
	}
	if err := a.Deregister(ctx, "missing"); err == nil {
		t.Fatal("expected not found")
	}
}
