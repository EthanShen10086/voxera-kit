package noop_test

import (
	"context"
	"testing"

	"github.com/EthanShen10086/voxera-kit/ad"
	"github.com/EthanShen10086/voxera-kit/ad/noop"
)

func TestNoopAdProvider(t *testing.T) {
	a := noop.NewAdapter()
	if a.Name() != "noop" {
		t.Fatalf("Name = %q", a.Name())
	}
	got, err := a.FetchAd(context.Background(), ad.Request{})
	if err != nil || got != nil {
		t.Fatalf("FetchAd: %+v err=%v", got, err)
	}
	if a.Available(context.Background()) {
		t.Fatal("expected unavailable")
	}
	if err := a.ReportImpression(context.Background(), "ad", "user"); err != nil {
		t.Fatal(err)
	}
	if err := a.ReportClick(context.Background(), "ad", "user"); err != nil {
		t.Fatal(err)
	}
}
