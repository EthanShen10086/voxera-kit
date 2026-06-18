package noop_test

import (
	"context"
	"testing"
	"time"

	"github.com/EthanShen10086/voxera-kit/aiquota"
	"github.com/EthanShen10086/voxera-kit/aiquota/noop"
)

func TestNoopQuotaStore(t *testing.T) {
	s := noop.NewStore()
	ctx := context.Background()
	if err := s.CheckQuota(ctx, "u", "gpt-4o", 100); err != nil {
		t.Fatal(err)
	}
	if err := s.RecordUsage(ctx, aiquota.UsageRecord{UserID: "u"}); err != nil {
		t.Fatal(err)
	}
	usage, err := s.GetUsage(ctx, "u")
	if err != nil || usage.UserID != "u" {
		t.Fatalf("GetUsage: %+v err=%v", usage, err)
	}
	q, err := s.GetQuota(ctx, "u")
	if err != nil || q.Tier != aiquota.TierEnterprise {
		t.Fatalf("GetQuota: %+v err=%v", q, err)
	}
	if err := s.SetTier(ctx, "u", aiquota.TierPro); err != nil {
		t.Fatal(err)
	}
	ok, err := s.IsWhitelisted(ctx, "u")
	if err != nil || !ok {
		t.Fatalf("IsWhitelisted: %v err=%v", ok, err)
	}
	if err := s.AddToWhitelist(ctx, aiquota.WhitelistEntry{UserID: "u"}); err != nil {
		t.Fatal(err)
	}
	if err := s.RemoveFromWhitelist(ctx, "u"); err != nil {
		t.Fatal(err)
	}
	list, err := s.ListWhitelist(ctx)
	if err != nil || list != nil {
		t.Fatalf("ListWhitelist: %v err=%v", list, err)
	}
	report, err := s.GetCostReport(ctx, "t", time.Now().Add(-24*time.Hour), time.Now())
	if err != nil || report.TenantID != "t" {
		t.Fatalf("GetCostReport: %+v err=%v", report, err)
	}
	release, err := s.AcquireConcurrency(ctx, "u")
	if err != nil {
		t.Fatal(err)
	}
	release()
}
