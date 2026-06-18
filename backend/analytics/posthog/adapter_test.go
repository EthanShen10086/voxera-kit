package posthog_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/EthanShen10086/voxera-kit/analytics"
	"github.com/EthanShen10086/voxera-kit/analytics/posthog"
)

func TestTrackBatchIdentifyAlias(t *testing.T) {
	var paths []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.URL.Path)
		if r.Method != http.MethodPost {
			t.Errorf("method = %s", r.Method)
		}
		_, _ = io.Copy(io.Discard, r.Body)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	a := posthog.New("ph-key", srv.URL)
	ctx := context.Background()
	ts := time.Now()

	if err := a.Track(ctx, analytics.Event{Name: "click", UserID: "u1", Timestamp: ts}); err != nil {
		t.Fatal(err)
	}
	if err := a.TrackBatch(ctx, []analytics.Event{{Name: "view", UserID: "u2", Timestamp: ts}}); err != nil {
		t.Fatal(err)
	}
	if err := a.Identify(ctx, analytics.UserProfile{UserID: "u1", Properties: map[string]any{"tier": "pro"}}); err != nil {
		t.Fatal(err)
	}
	if err := a.Alias(ctx, "anon", "u1"); err != nil {
		t.Fatal(err)
	}

	if len(paths) < 4 {
		t.Fatalf("paths = %v", paths)
	}
}

func TestQueryFunnelAndRetention(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/insights/trend") {
			t.Fatalf("path = %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"result": []map[string]any{
				{"name": "signup", "count": float64(100), "conversion_rate": float64(100)},
				{"name": "purchase", "count": float64(40), "conversion_rate": float64(40)},
			},
		})
	}))
	defer srv.Close()

	a := posthog.New("ph-key", srv.URL)
	from := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	to := from.Add(7 * 24 * time.Hour)

	funnel, err := a.QueryFunnel(context.Background(), analytics.FunnelQuery{
		From: from, To: to,
		Steps: []analytics.FunnelStep{{Name: "signup"}, {Name: "purchase"}},
	})
	if err != nil || len(funnel.Steps) != 2 {
		t.Fatalf("funnel: %+v err=%v", funnel, err)
	}

	retention, err := a.QueryRetention(context.Background(), analytics.RetentionQuery{
		From: from, To: to, CohortEvent: "signup", ReturnEvent: "login",
		Granularity: analytics.GranularityWeek, Periods: 4,
	})
	if err != nil || retention == nil {
		t.Fatalf("retention: %+v err=%v", retention, err)
	}
}

func TestQueryPathNotSupported(t *testing.T) {
	a := posthog.New("k", "http://localhost")
	_, err := a.QueryPath(context.Background(), analytics.PathQuery{})
	if err == nil || !strings.Contains(err.Error(), "not supported") {
		t.Fatalf("err = %v", err)
	}
}

func TestQueryEventsAndProfile(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.Contains(r.URL.Path, "/events"):
			_ = json.NewEncoder(w).Encode(map[string]any{
				"results": []map[string]any{{
					"event": "click", "distinct_id": "u1",
					"timestamp": "2026-06-01T00:00:00Z",
				}},
			})
		case strings.Contains(r.URL.Path, "/persons"):
			_ = json.NewEncoder(w).Encode(map[string]any{
				"results": []map[string]any{{
					"distinct_ids": []string{"u1"},
					"properties":   map[string]any{"email": "a@b.com"},
					"created_at":   "2026-06-01T00:00:00Z",
				}},
			})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	a := posthog.New("ph-key", srv.URL)
	events, err := a.QueryEvents(context.Background(), analytics.EventQuery{UserID: "u1", Limit: 10})
	if err != nil || len(events) != 1 || events[0].Name != "click" {
		t.Fatalf("events: %+v err=%v", events, err)
	}
	profile, err := a.QueryUserProfile(context.Background(), "u1")
	if err != nil || profile.UserID != "u1" {
		t.Fatalf("profile: %+v err=%v", profile, err)
	}
}
