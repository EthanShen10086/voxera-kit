package memory_test

import (
	"context"
	"testing"

	"github.com/EthanShen10086/voxera-kit/experiment"
	"github.com/EthanShen10086/voxera-kit/experiment/memory"
)

func TestExperimentLifecycle(t *testing.T) {
	ctx := context.Background()
	s := memory.NewStore()

	cfg := experiment.Config{
		Key:  "checkout",
		Name: "Checkout button",
		Variants: []experiment.Variant{
			{Key: "control", Weight: 50, IsControl: true},
			{Key: "treatment", Weight: 50},
		},
		Metrics: []experiment.MetricDef{
			{Key: "conversion", Type: experiment.MetricConversion},
			{Key: "revenue", Type: experiment.MetricRevenue},
		},
	}
	if err := s.Create(ctx, cfg); err != nil {
		t.Fatal(err)
	}
	if err := s.Create(ctx, cfg); err == nil {
		t.Fatal("expected duplicate create error")
	}

	got, err := s.Get(ctx, "checkout")
	if err != nil || got.Key != "checkout" || got.Status != experiment.StatusDraft {
		t.Fatalf("Get: %+v err=%v", got, err)
	}

	if err := s.Start(ctx, "checkout"); err != nil {
		t.Fatal(err)
	}
	if err := s.Stop(ctx, "checkout"); err != nil {
		t.Fatal(err)
	}

	list, err := s.List(ctx, experiment.StatusComplete)
	if err != nil || len(list) != 1 {
		t.Fatalf("List: %v err=%v", list, err)
	}
}

func TestAssignAndResults(t *testing.T) {
	ctx := context.Background()
	s := memory.NewStore()
	cfg := experiment.Config{
		Key: "pricing",
		Variants: []experiment.Variant{
			{Key: "control", Weight: 1, IsControl: true},
			{Key: "variant-a", Weight: 1},
		},
		Metrics: []experiment.MetricDef{
			{Key: "conversion", Type: experiment.MetricConversion},
		},
	}
	_ = s.Create(ctx, cfg)
	_ = s.Start(ctx, "pricing")

	a1, err := s.Assign(ctx, "pricing", "user-1")
	if err != nil {
		t.Fatal(err)
	}
	a2, err := s.Assign(ctx, "pricing", "user-1")
	if err != nil || a2.VariantKey != a1.VariantKey {
		t.Fatalf("assignments differ: %+v vs %+v", a1, a2)
	}

	if _, err := s.GetAssignment(ctx, "pricing", "unknown"); err != nil {
		t.Fatal(err)
	}

	_, _ = s.Assign(ctx, "pricing", "user-2")
	_ = s.RecordMetric(ctx, "pricing", "user-1", "conversion", 1)
	_ = s.RecordMetric(ctx, "pricing", "user-2", "conversion", 1)
	_ = s.RecordMetric(ctx, "pricing", "user-2", "revenue", 9.99)

	result, err := s.GetResults(ctx, "pricing")
	if err != nil || result.TotalUsers < 2 || len(result.Metrics) == 0 {
		t.Fatalf("results: %+v err=%v", result, err)
	}
}

func TestExperimentErrors(t *testing.T) {
	ctx := context.Background()
	s := memory.NewStore()
	if _, err := s.Get(ctx, "missing"); err == nil {
		t.Fatal("expected not found")
	}
	if err := s.RecordMetric(ctx, "missing", "u", "m", 1); err == nil {
		t.Fatal("expected not found")
	}
}
