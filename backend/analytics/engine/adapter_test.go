package engine_test

import (
	"context"
	"testing"
	"time"

	"github.com/EthanShen10086/voxera-kit/analytics"
	"github.com/EthanShen10086/voxera-kit/analytics/engine"
)

func baseTime() time.Time {
	return time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
}

func TestTrackIdentifyAlias(t *testing.T) {
	ctx := context.Background()
	eng := engine.New()
	ts := baseTime()

	if err := eng.Track(ctx, analytics.Event{
		Name: "page_view", UserID: "u1", TenantID: "t1", SessionID: "s1",
		Timestamp: ts,
		Ctx:       analytics.EventContext{UTMSource: "google", Referrer: "https://ref"},
	}); err != nil {
		t.Fatal(err)
	}
	if err := eng.TrackBatch(ctx, []analytics.Event{
		{Name: "click", UserID: "u1", TenantID: "t1", SessionID: "s2", Timestamp: ts.Add(time.Minute)},
	}); err != nil {
		t.Fatal(err)
	}
	if err := eng.Identify(ctx, analytics.UserProfile{
		UserID: "u1", TenantID: "t1", Properties: map[string]any{"plan": "pro"},
	}); err != nil {
		t.Fatal(err)
	}
	if err := eng.Alias(ctx, "anon-1", "u1"); err != nil {
		t.Fatal(err)
	}

	profile, err := eng.QueryUserProfile(ctx, "u1")
	if err != nil || profile.EventCount < 2 || profile.Properties["plan"] != "pro" {
		t.Fatalf("profile: %+v err=%v", profile, err)
	}
	if profile.SessionCount != 2 {
		t.Fatalf("sessions = %d", profile.SessionCount)
	}
}

func TestQueryFunnel(t *testing.T) {
	ctx := context.Background()
	eng := engine.New()
	ts := baseTime()
	from := ts.Add(-time.Hour)
	to := ts.Add(time.Hour)

	for _, ev := range []analytics.Event{
		{Name: "signup", UserID: "u1", TenantID: "t1", Timestamp: ts},
		{Name: "purchase", UserID: "u1", TenantID: "t1", Timestamp: ts.Add(10 * time.Second)},
		{Name: "signup", UserID: "u2", TenantID: "t1", Timestamp: ts},
	} {
		_ = eng.Track(ctx, ev)
	}

	result, err := eng.QueryFunnel(ctx, analytics.FunnelQuery{
		TenantID: "t1",
		From:     from, To: to,
		WindowSec: 3600,
		Steps: []analytics.FunnelStep{
			{Name: "signup"},
			{Name: "purchase"},
		},
	})
	if err != nil || len(result.Steps) != 2 {
		t.Fatalf("funnel: %+v err=%v", result, err)
	}
	if result.Steps[0].EnteredCount != 2 || result.Steps[1].EnteredCount != 1 {
		t.Fatalf("counts: step0=%d step1=%d", result.Steps[0].EnteredCount, result.Steps[1].EnteredCount)
	}
}

func TestQueryRetentionPathAndEvents(t *testing.T) {
	ctx := context.Background()
	eng := engine.New()
	ts := time.Date(2026, 6, 10, 10, 0, 0, 0, time.UTC)

	_ = eng.Track(ctx, analytics.Event{Name: "signup", UserID: "u1", TenantID: "t1", Timestamp: ts})
	_ = eng.Track(ctx, analytics.Event{Name: "login", UserID: "u1", TenantID: "t1", Timestamp: ts.Add(24 * time.Hour)})

	retention, err := eng.QueryRetention(ctx, analytics.RetentionQuery{
		TenantID: "t1", CohortEvent: "signup", ReturnEvent: "login",
		From: ts.Add(-time.Hour), To: ts.Add(72 * time.Hour),
		Granularity: analytics.GranularityDay, Periods: 2,
	})
	if err != nil || len(retention.Cohorts) == 0 {
		t.Fatalf("retention: %+v err=%v", retention, err)
	}

	_ = eng.Track(ctx, analytics.Event{Name: "home", UserID: "u2", TenantID: "t1", SessionID: "sess", Timestamp: ts})
	_ = eng.Track(ctx, analytics.Event{Name: "checkout", UserID: "u2", TenantID: "t1", SessionID: "sess", Timestamp: ts.Add(time.Second)})

	path, err := eng.QueryPath(ctx, analytics.PathQuery{
		TenantID: "t1", From: ts.Add(-time.Hour), To: ts.Add(time.Hour),
		StartEvent: "home", MaxSteps: 3,
	})
	if err != nil || len(path.Paths) == 0 {
		t.Fatalf("path: %+v err=%v", path, err)
	}

	events, err := eng.QueryEvents(ctx, analytics.EventQuery{
		TenantID: "t1", UserID: "u1", From: ts.Add(-time.Hour), To: ts.Add(72*time.Hour), Limit: 10,
	})
	if err != nil || len(events) == 0 {
		t.Fatalf("events: %v err=%v", events, err)
	}
}

func TestQueryFunnelEmptySteps(t *testing.T) {
	eng := engine.New()
	result, err := eng.QueryFunnel(context.Background(), analytics.FunnelQuery{})
	if err != nil || result == nil {
		t.Fatalf("empty funnel: %+v err=%v", result, err)
	}
}
