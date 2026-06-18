package memory_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/EthanShen10086/voxera-kit/aiquota"
	"github.com/EthanShen10086/voxera-kit/aiquota/memory"
)

func TestCheckQuotaAndRecordUsage(t *testing.T) {
	ctx := context.Background()
	s := memory.NewStore()

	if err := s.CheckQuota(ctx, "u1", "deepseek-chat", 100); err != nil {
		t.Fatalf("CheckQuota: %v", err)
	}
	if err := s.CheckQuota(ctx, "u1", "gpt-4o", 100); !errors.Is(err, aiquota.ErrModelNotAllowed) {
		t.Fatalf("model check: %v", err)
	}

	if err := s.RecordUsage(ctx, aiquota.UsageRecord{
		UserID: "u1", TenantID: "t1", Model: "deepseek-chat",
		InputTokens: 100, OutputTokens: 50, CostCents: 10,
		Timestamp: time.Now(),
	}); err != nil {
		t.Fatal(err)
	}
	usage, err := s.GetUsage(ctx, "u1")
	if err != nil || usage.DailyTokens != 150 {
		t.Fatalf("usage: %+v err=%v", usage, err)
	}
}

func TestWhitelistAndTier(t *testing.T) {
	ctx := context.Background()
	s := memory.NewStore()

	_ = s.AddToWhitelist(ctx, aiquota.WhitelistEntry{UserID: "vip", Reason: "test"})
	ok, err := s.IsWhitelisted(ctx, "vip")
	if err != nil || !ok {
		t.Fatalf("whitelist: ok=%v err=%v", ok, err)
	}
	if err := s.CheckQuota(ctx, "vip", "gpt-4o", 1_000_000); err != nil {
		t.Fatalf("whitelisted should bypass quota: %v", err)
	}
	if err := s.RemoveFromWhitelist(ctx, "vip"); err != nil {
		t.Fatal(err)
	}
	ok, _ = s.IsWhitelisted(ctx, "vip")
	if ok {
		t.Fatal("expected removed from whitelist")
	}

	if err := s.SetTier(ctx, "pro-user", aiquota.TierPro); err != nil {
		t.Fatal(err)
	}
	q, err := s.GetQuota(ctx, "pro-user")
	if err != nil || q.Tier != aiquota.TierPro {
		t.Fatalf("GetQuota: %+v err=%v", q, err)
	}
	if err := s.CheckQuota(ctx, "pro-user", "gpt-4o-mini", 1000); err != nil {
		t.Fatalf("pro tier model: %v", err)
	}
}

func TestConcurrencyAndCostReport(t *testing.T) {
	ctx := context.Background()
	s := memory.NewStore()

	release, err := s.AcquireConcurrency(ctx, "u1")
	if err != nil {
		t.Fatal(err)
	}
	release()

	_, err = s.AcquireConcurrency(ctx, "u1")
	if err != nil {
		t.Fatal(err)
	}
	_, err = s.AcquireConcurrency(ctx, "u1")
	if !errors.Is(err, aiquota.ErrConcurrencyLimit) {
		t.Fatalf("concurrency limit: %v", err)
	}

	now := time.Now()
	_ = s.RecordUsage(ctx, aiquota.UsageRecord{
		UserID: "u1", TenantID: "tenant-a", Model: "deepseek-chat",
		InputTokens: 10, OutputTokens: 5, CostCents: 3, Timestamp: now,
	})
	report, err := s.GetCostReport(ctx, "tenant-a", now.Add(-time.Hour), now.Add(time.Hour))
	if err != nil || report.TotalRequests != 1 || report.TotalCostCents != 3 {
		t.Fatalf("report: %+v err=%v", report, err)
	}

	entries, err := s.ListWhitelist(ctx)
	if err != nil || len(entries) != 0 {
		t.Fatalf("whitelist list: %v", entries)
	}
}
