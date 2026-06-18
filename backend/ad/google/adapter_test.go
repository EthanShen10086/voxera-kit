package google_test

import (
	"context"
	"testing"

	"github.com/EthanShen10086/voxera-kit/ad"
	"github.com/EthanShen10086/voxera-kit/ad/google"
)

func TestGoogleAdapter(t *testing.T) {
	a := google.NewAdapter("pub", "key")
	if a.Name() != "google_ads" {
		t.Fatalf("Name = %q", a.Name())
	}
	if !a.Available(context.Background()) {
		t.Fatal("expected available")
	}
	if _, err := a.FetchAd(context.Background(), ad.Request{}); err == nil {
		t.Fatal("expected not implemented error")
	}
	if err := a.ReportImpression(context.Background(), "ad", "u"); err != nil {
		t.Fatal(err)
	}
}
