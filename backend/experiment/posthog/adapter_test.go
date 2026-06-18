package posthog_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/EthanShen10086/voxera-kit/experiment"
	"github.com/EthanShen10086/voxera-kit/experiment/posthog"
)

func TestCreateListAssign(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && strings.Contains(r.URL.Path, "/experiments"):
			w.WriteHeader(http.StatusCreated)
		case r.Method == http.MethodGet && strings.Contains(r.URL.Path, "/experiments"):
			_ = json.NewEncoder(w).Encode(map[string]any{
				"results": []map[string]any{{
					"id": 1, "name": "Checkout", "description": "btn",
					"feature_flag": map[string]string{"key": "checkout-exp"},
				}},
			})
		case strings.Contains(r.URL.Path, "/decide"):
			_ = json.NewEncoder(w).Encode(map[string]any{
				"featureFlags": map[string]string{"checkout-exp": "variant-a"},
			})
		case strings.Contains(r.URL.Path, "/capture"):
			w.WriteHeader(http.StatusOK)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	a := posthog.NewAdapter("ph-key", srv.URL, "proj-1")
	ctx := context.Background()

	cfg := experiment.Config{
		Key: "checkout-exp", Name: "Checkout",
		Variants: []experiment.Variant{{Key: "control"}, {Key: "variant-a"}},
	}
	if err := a.Create(ctx, cfg); err != nil {
		t.Fatal(err)
	}
	got, err := a.Get(ctx, "checkout-exp")
	if err != nil || got.Key != "checkout-exp" {
		t.Fatalf("Get: %+v err=%v", got, err)
	}
	assign, err := a.Assign(ctx, "checkout-exp", "user-1")
	if err != nil || assign.VariantKey != "variant-a" {
		t.Fatalf("Assign: %+v err=%v", assign, err)
	}
	if err := a.RecordMetric(ctx, "checkout-exp", "user-1", "conversion", 1); err != nil {
		t.Fatal(err)
	}
	if err := a.Start(ctx, "checkout-exp"); err != nil {
		t.Fatal(err)
	}
	if err := a.Stop(ctx, "checkout-exp"); err != nil {
		t.Fatal(err)
	}
}
