package llm

import "unicode"

// Pricing per 1M tokens in USD cents (multiply by 100 from dollar prices).
// Updated as of 2025 public pricing pages.
var pricingTable = map[string][2]int64{
	// OpenAI
	"gpt-4o":         {250, 1000},
	"gpt-4o-mini":    {15, 60},
	"gpt-4-turbo":    {1000, 3000},
	"gpt-4":          {3000, 6000},
	"gpt-3.5-turbo":  {50, 150},
	// DeepSeek
	"deepseek-chat":     {14, 28},
	"deepseek-coder":    {14, 28},
	"deepseek-reasoner": {55, 219},
	// Qwen
	"qwen-turbo": {30, 60},
	"qwen-plus":  {80, 200},
	"qwen-max":   {200, 600},
	// Claude
	"claude-sonnet-4-20250514":    {300, 1500},
	"claude-3-5-haiku-20241022":   {80, 400},
	// Hunyuan
	"hunyuan-pro":      {300, 900},
	"hunyuan-standard": {45, 80},
	"hunyuan-lite":     {0, 0},
}

// EstimateTokens returns an approximate token count for text.
// It uses a simple heuristic: roughly 4 characters per token for Latin scripts
// and roughly 2 characters per token for CJK characters.
func EstimateTokens(text string) int {
	if len(text) == 0 {
		return 0
	}
	var latinChars, cjkChars int
	for _, r := range text {
		if isCJK(r) {
			cjkChars++
		} else {
			latinChars++
		}
	}
	tokens := latinChars/4 + cjkChars/2
	if tokens == 0 && len(text) > 0 {
		tokens = 1
	}
	return tokens
}

// EstimateCost calculates the estimated cost in USD cents for a given model
// and token counts. Returns 0 if the model is not in the pricing table.
func EstimateCost(model string, inputTokens, outputTokens int) int64 {
	prices, ok := pricingTable[model]
	if !ok {
		return 0
	}
	inputCost := int64(inputTokens) * prices[0] / 1_000_000
	outputCost := int64(outputTokens) * prices[1] / 1_000_000
	return inputCost + outputCost
}

// isCJK reports whether r is a CJK unified ideograph or common fullwidth character.
func isCJK(r rune) bool {
	return unicode.Is(unicode.Han, r) ||
		unicode.Is(unicode.Katakana, r) ||
		unicode.Is(unicode.Hiragana, r) ||
		unicode.Is(unicode.Hangul, r)
}
