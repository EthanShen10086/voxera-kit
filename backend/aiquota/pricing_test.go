package aiquota_test

import (
	"testing"

	"github.com/EthanShen10086/voxera-kit/aiquota"
)

func TestDefaultPricingAndCost(t *testing.T) {
	prices := aiquota.DefaultPricing()
	if len(prices) == 0 {
		t.Fatal("expected pricing table")
	}
	cost := aiquota.CalculateCost("gpt-4o", 1_000_000, 1_000_000)
	if cost != 1250 {
		t.Fatalf("CalculateCost = %d", cost)
	}
	if aiquota.CalculateCost("unknown-model", 100, 100) != 0 {
		t.Fatal("expected zero cost for unknown model")
	}
}

func TestDefaultQuotas(t *testing.T) {
	quotas := aiquota.DefaultQuotas()
	free, ok := quotas[aiquota.TierFree]
	if !ok || free.DailyTokens != 10000 {
		t.Fatalf("free tier: %+v", free)
	}
	ent := quotas[aiquota.TierEnterprise]
	if len(ent.AllowedModels) == 0 {
		t.Fatal("enterprise should allow models")
	}
}
