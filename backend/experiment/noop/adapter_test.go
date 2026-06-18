package noop_test

import (
	"context"
	"testing"

	"github.com/EthanShen10086/voxera-kit/experiment"
	"github.com/EthanShen10086/voxera-kit/experiment/noop"
)

func TestNoopExperiment(t *testing.T) {
	a := noop.NewAdapter()
	ctx := context.Background()
	if err := a.Create(ctx, experiment.Config{Key: "exp"}); err != nil {
		t.Fatal(err)
	}
	cfg, err := a.Get(ctx, "exp")
	if err != nil || cfg != nil {
		t.Fatalf("Get: %+v err=%v", cfg, err)
	}
	list, err := a.List(ctx, experiment.StatusDraft)
	if err != nil || len(list) != 0 {
		t.Fatalf("List: %v err=%v", list, err)
	}
	if err := a.Start(ctx, "exp"); err != nil {
		t.Fatal(err)
	}
	if err := a.Stop(ctx, "exp"); err != nil {
		t.Fatal(err)
	}
	assign, err := a.Assign(ctx, "exp", "user")
	if err != nil || assign.VariantKey != "control" {
		t.Fatalf("Assign: %+v err=%v", assign, err)
	}
	prev, err := a.GetAssignment(ctx, "exp", "user")
	if err != nil || prev != nil {
		t.Fatalf("GetAssignment: %+v err=%v", prev, err)
	}
	if err := a.RecordMetric(ctx, "exp", "user", "conv", 1); err != nil {
		t.Fatal(err)
	}
	res, err := a.GetResults(ctx, "exp")
	if err != nil || res.ExperimentKey != "exp" {
		t.Fatalf("GetResults: %+v err=%v", res, err)
	}
}
