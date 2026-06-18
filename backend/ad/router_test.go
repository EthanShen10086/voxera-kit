package ad_test

import (
	"context"
	"testing"

	"github.com/EthanShen10086/voxera-kit/ad"
	"github.com/EthanShen10086/voxera-kit/ad/noop"
	"github.com/EthanShen10086/voxera-kit/ad/selfhosted"
)

type stubTracker struct {
	count int
}

func (s *stubTracker) TrackImpression(_ context.Context, _ *ad.Ad, _ string) error {
	s.count++
	return nil
}

func (s *stubTracker) TrackClick(_ context.Context, _ *ad.Ad, _ string) error {
	return nil
}

func TestRouterFetchPaidUser(t *testing.T) {
	r := ad.NewRouter(ad.Config{}, nil)
	got, err := r.Fetch(context.Background(), ad.Request{IsPaidUser: true})
	if err != nil || got != nil {
		t.Fatalf("Fetch: %+v err=%v", got, err)
	}
}

func TestRouterFetchSelfHosted(t *testing.T) {
	tracker := &stubTracker{}
	r := ad.NewRouter(ad.Config{}, tracker)
	inv := selfhosted.NewAdapter([]ad.Ad{{ID: "a1", SlotType: ad.SlotBanner}})
	r.RegisterProvider(inv, 1)
	r.RegisterProvider(noop.NewAdapter(), 0)

	got, err := r.Fetch(context.Background(), ad.Request{UserID: "u1"})
	if err != nil || got == nil || got.ID != "a1" {
		t.Fatalf("Fetch: %+v err=%v", got, err)
	}
	if tracker.count != 1 {
		t.Fatalf("impressions = %d", tracker.count)
	}
}

func TestRouterFallback(t *testing.T) {
	r := ad.NewRouter(ad.Config{FallbackHTML: "<b>ad</b>"}, nil)
	got, err := r.Fetch(context.Background(), ad.Request{})
	if err != nil || got == nil || got.ID != "fallback" {
		t.Fatalf("Fetch: %+v err=%v", got, err)
	}
}

func TestRouterMinorHide(t *testing.T) {
	r := ad.NewRouter(ad.Config{MinorPolicy: "hide"}, nil)
	got, err := r.Fetch(context.Background(), ad.Request{IsMinor: true})
	if err != nil || got != nil {
		t.Fatalf("Fetch: %+v err=%v", got, err)
	}
}
