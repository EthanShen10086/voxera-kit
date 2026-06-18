package selfhosted_test

import (
	"context"
	"testing"

	"github.com/EthanShen10086/voxera-kit/ad"
	"github.com/EthanShen10086/voxera-kit/ad/selfhosted"
)

func TestSelfHostedAdapter(t *testing.T) {
	a := selfhosted.NewAdapter([]ad.Ad{{ID: "ad1", SlotType: ad.SlotBanner}})
	if a.Name() != "self_hosted" || !a.Available(context.Background()) {
		t.Fatal("adapter not ready")
	}
	got, err := a.FetchAd(context.Background(), ad.Request{})
	if err != nil || got == nil || got.ID != "ad1" {
		t.Fatalf("FetchAd: %+v err=%v", got, err)
	}
	if err := a.ReportImpression(context.Background(), "ad1", "u"); err != nil {
		t.Fatal(err)
	}
	if err := a.ReportClick(context.Background(), "ad1", "u"); err != nil {
		t.Fatal(err)
	}
	if a.Impressions() != 1 || a.Clicks() != 1 {
		t.Fatalf("metrics = %d/%d", a.Impressions(), a.Clicks())
	}
}

func TestSelfHostedEmptyInventory(t *testing.T) {
	a := selfhosted.NewAdapter(nil)
	if a.Available(context.Background()) {
		t.Fatal("expected unavailable")
	}
	got, err := a.FetchAd(context.Background(), ad.Request{})
	if err != nil || got != nil {
		t.Fatalf("FetchAd: %+v err=%v", got, err)
	}
}
