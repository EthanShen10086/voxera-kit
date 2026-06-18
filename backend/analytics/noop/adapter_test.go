package noop_test

import (
	"context"
	"testing"

	"github.com/EthanShen10086/voxera-kit/analytics"
	"github.com/EthanShen10086/voxera-kit/analytics/noop"
)

func TestNoopAnalytics(t *testing.T) {
	a := noop.New()
	ctx := context.Background()
	if err := a.Track(ctx, analytics.Event{Name: "click"}); err != nil {
		t.Fatal(err)
	}
	if err := a.TrackBatch(ctx, []analytics.Event{{Name: "a"}}); err != nil {
		t.Fatal(err)
	}
	if err := a.Identify(ctx, analytics.UserProfile{UserID: "u"}); err != nil {
		t.Fatal(err)
	}
	if err := a.Alias(ctx, "prev", "u"); err != nil {
		t.Fatal(err)
	}
	if _, err := a.QueryFunnel(ctx, analytics.FunnelQuery{}); err != nil {
		t.Fatal(err)
	}
	if _, err := a.QueryRetention(ctx, analytics.RetentionQuery{}); err != nil {
		t.Fatal(err)
	}
	if _, err := a.QueryPath(ctx, analytics.PathQuery{}); err != nil {
		t.Fatal(err)
	}
	if _, err := a.QueryEvents(ctx, analytics.EventQuery{}); err != nil {
		t.Fatal(err)
	}
	if _, err := a.QueryUserProfile(ctx, "u"); err != nil {
		t.Fatal(err)
	}
}
