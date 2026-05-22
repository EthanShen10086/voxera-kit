package aiquota

// ModelPrice holds the pricing for a specific model.
type ModelPrice struct {
	Model       string
	InputPer1M  int64 // USD cents per 1M input tokens
	OutputPer1M int64 // USD cents per 1M output tokens
}

// DefaultPricing returns the built-in model pricing table.
func DefaultPricing() []ModelPrice {
	return []ModelPrice{
		{Model: "gpt-4o", InputPer1M: 250, OutputPer1M: 1000},
		{Model: "gpt-4o-mini", InputPer1M: 15, OutputPer1M: 60},
		{Model: "gpt-4", InputPer1M: 3000, OutputPer1M: 6000},
		{Model: "deepseek-chat", InputPer1M: 14, OutputPer1M: 28},
		{Model: "deepseek-coder", InputPer1M: 14, OutputPer1M: 28},
		{Model: "qwen-turbo", InputPer1M: 30, OutputPer1M: 60},
		{Model: "qwen-plus", InputPer1M: 80, OutputPer1M: 200},
		{Model: "qwen-max", InputPer1M: 200, OutputPer1M: 600},
		{Model: "claude-sonnet-4-20250514", InputPer1M: 300, OutputPer1M: 1500},
		{Model: "claude-3-5-haiku-20241022", InputPer1M: 80, OutputPer1M: 400},
		{Model: "hunyuan-pro", InputPer1M: 100, OutputPer1M: 300},
		{Model: "hunyuan-standard", InputPer1M: 45, OutputPer1M: 80},
		{Model: "hunyuan-lite", InputPer1M: 0, OutputPer1M: 0},
	}
}

// CalculateCost computes the cost in cents for a given model and token usage.
func CalculateCost(model string, inputTokens, outputTokens int) int64 {
	for _, p := range DefaultPricing() {
		if p.Model == model {
			inputCost := p.InputPer1M * int64(inputTokens) / 1_000_000
			outputCost := p.OutputPer1M * int64(outputTokens) / 1_000_000
			return inputCost + outputCost
		}
	}
	return 0
}

// DefaultQuotas returns the standard quota configurations for each tier.
// All tiers have finite quotas. Unlimited access is ONLY granted via the
// configurable whitelist (admin-managed), never automatically by tier.
func DefaultQuotas() map[Tier]Quota {
	allModels := make([]string, 0, len(DefaultPricing()))
	for _, p := range DefaultPricing() {
		allModels = append(allModels, p.Model)
	}

	return map[Tier]Quota{
		TierFree: {
			Tier:            TierFree,
			DailyTokens:     10000,
			MonthlyTokens:   100000,
			DailyRequests:   20,
			MonthlyRequests: 0,
			AllowedModels:   []string{"deepseek-chat", "hunyuan-lite"},
			ConcurrentLimit: 1,
			Priority:        0,
			OverQuota:       PolicyReject,
		},
		TierPro: {
			Tier:            TierPro,
			DailyTokens:     500000,
			MonthlyTokens:   5000000,
			DailyRequests:   500,
			MonthlyRequests: 0,
			AllowedModels:   []string{"gpt-4o-mini", "deepseek-chat", "deepseek-coder", "qwen-turbo", "qwen-plus", "hunyuan-standard"},
			ConcurrentLimit: 5,
			Priority:        5,
			OverQuota:       PolicyDegrade,
			DegradeModel:    "deepseek-chat",
		},
		TierEnterprise: {
			Tier:            TierEnterprise,
			DailyTokens:     5000000,
			MonthlyTokens:   50000000,
			DailyRequests:   5000,
			MonthlyRequests: 0,
			AllowedModels:   allModels,
			ConcurrentLimit: 20,
			Priority:        10,
			OverQuota:       PolicyNotify,
		},
	}
}
