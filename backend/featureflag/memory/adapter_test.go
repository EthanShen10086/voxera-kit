package memory_test

import (
	"context"
	"testing"

	"github.com/EthanShen10086/voxera-kit/featureflag"
	"github.com/EthanShen10086/voxera-kit/featureflag/memory"
)

func TestIsEnabled_DenyAllowAndPercentage(t *testing.T) {
	ctx := context.Background()
	a := memory.NewAdapter()

	flag := featureflag.Flag{
		Key:        "new-ui",
		Enabled:    true,
		Percentage: 50,
		DenyList:   []string{"blocked"},
		AllowList:  []string{"vip"},
	}
	if err := a.SetFlag(ctx, flag); err != nil {
		t.Fatal(err)
	}

	enabled, err := a.IsEnabled(ctx, "new-ui", featureflag.EvalContext{UserID: "blocked"})
	if err != nil || enabled {
		t.Fatalf("deny list: enabled=%v err=%v", enabled, err)
	}

	enabled, err = a.IsEnabled(ctx, "new-ui", featureflag.EvalContext{UserID: "vip"})
	if err != nil || !enabled {
		t.Fatalf("allow list: enabled=%v err=%v", enabled, err)
	}

	enabled1, _ := a.IsEnabled(ctx, "new-ui", featureflag.EvalContext{UserID: "stable-user"})
	enabled2, _ := a.IsEnabled(ctx, "new-ui", featureflag.EvalContext{UserID: "stable-user"})
	if enabled1 != enabled2 {
		t.Fatal("percentage rollout should be deterministic")
	}
}

func TestIsEnabled_MissingAndDisabled(t *testing.T) {
	ctx := context.Background()
	a := memory.NewAdapter()

	enabled, err := a.IsEnabled(ctx, "missing", featureflag.EvalContext{UserID: "u1"})
	if err != nil || enabled {
		t.Fatalf("missing flag: enabled=%v", enabled)
	}

	if err := a.SetFlag(ctx, featureflag.Flag{Key: "off", Enabled: false, Percentage: 100}); err != nil {
		t.Fatal(err)
	}
	enabled, err = a.IsEnabled(ctx, "off", featureflag.EvalContext{UserID: "u1"})
	if err != nil || enabled {
		t.Fatalf("disabled flag: enabled=%v", enabled)
	}
}

func TestGetFlags(t *testing.T) {
	ctx := context.Background()
	a := memory.NewAdapter()
	_ = a.SetFlag(ctx, featureflag.Flag{Key: "a", Enabled: true})
	_ = a.SetFlag(ctx, featureflag.Flag{Key: "b", Enabled: true})
	flags, err := a.GetFlags(ctx)
	if err != nil || len(flags) != 2 {
		t.Fatalf("flags = %v err=%v", flags, err)
	}
}
