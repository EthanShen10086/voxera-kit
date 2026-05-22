// Package aiquota manages AI usage quotas, metering, billing,
// whitelist/VIP access for expensive LLM calls.
package aiquota

import (
	"context"
	"errors"
	"time"
)

// ErrQuotaExceeded indicates the user has exhausted their AI usage allowance.
var ErrQuotaExceeded = errors.New("aiquota: usage quota exceeded")

// ErrModelNotAllowed indicates the requested model is not available for the user's tier.
var ErrModelNotAllowed = errors.New("aiquota: model not allowed for this tier")

// ErrConcurrencyLimit indicates too many concurrent AI requests for this user.
var ErrConcurrencyLimit = errors.New("aiquota: concurrent request limit reached")

// Tier defines the AI access level for a user or tenant.
type Tier string

const (
	// TierFree is the basic free access tier.
	TierFree Tier = "free"
	// TierPro is the paid professional tier.
	TierPro Tier = "pro"
	// TierEnterprise is the enterprise tier with higher limits.
	TierEnterprise Tier = "enterprise"
	// TierVIP is the VIP tier with unlimited access.
	TierVIP Tier = "vip"
	// TierAdmin is the administrative tier with full access.
	TierAdmin Tier = "admin"
)

// OverQuotaPolicy controls behavior when a user exceeds their quota.
type OverQuotaPolicy string

const (
	// PolicyReject denies the request immediately.
	PolicyReject OverQuotaPolicy = "reject"
	// PolicyDegrade falls back to a cheaper model.
	PolicyDegrade OverQuotaPolicy = "degrade"
	// PolicyQueue places the request in a waiting queue.
	PolicyQueue OverQuotaPolicy = "queue"
	// PolicyNotify allows the request but sends a notification.
	PolicyNotify OverQuotaPolicy = "notify"
)

// Quota defines the AI usage limits for a tier.
type Quota struct {
	Tier            Tier
	DailyTokens     int64
	MonthlyTokens   int64
	DailyRequests   int64
	MonthlyRequests int64
	MaxTokensPerReq int
	AllowedModels   []string
	ConcurrentLimit int
	Priority        int
	OverQuota       OverQuotaPolicy
	DegradeModel    string
}

// Usage tracks current consumption for a user.
type Usage struct {
	UserID           string
	TenantID         string
	Tier             Tier
	DailyTokens      int64
	MonthlyTokens    int64
	DailyRequests    int64
	MonthlyRequests  int64
	TotalCostCents   int64
	LastResetDaily   time.Time
	LastResetMonthly time.Time
}

// UsageRecord captures the details of a single AI call for metering.
type UsageRecord struct {
	RequestID    string
	UserID       string
	TenantID     string
	Model        string
	InputTokens  int
	OutputTokens int
	CostCents    int64
	Latency      time.Duration
	Success      bool
	Timestamp    time.Time
}

// CostReport aggregates cost data for a tenant over a period.
type CostReport struct {
	TenantID       string
	Period         string
	TotalRequests  int64
	TotalTokens    int64
	TotalCostCents int64
	ByModel        map[string]ModelCost
	ByUser         map[string]int64
}

// ModelCost holds per-model cost breakdown.
type ModelCost struct {
	Requests int64
	Tokens   int64
	Cost     int64
}

// WhitelistEntry records an admin-granted unlimited access override.
type WhitelistEntry struct {
	UserID    string
	Reason    string
	GrantedBy string
	GrantedAt time.Time
}

// Manager handles quota checking, metering, and enforcement.
type Manager interface {
	// CheckQuota verifies the user has sufficient quota for the request.
	CheckQuota(ctx context.Context, userID string, model string, estimatedTokens int) error
	// RecordUsage persists a completed AI call for metering and billing.
	RecordUsage(ctx context.Context, record UsageRecord) error
	// GetUsage retrieves current usage counters for a user.
	GetUsage(ctx context.Context, userID string) (*Usage, error)
	// GetQuota returns the quota configuration for a user.
	GetQuota(ctx context.Context, userID string) (*Quota, error)
	// SetTier changes the access tier for a user.
	SetTier(ctx context.Context, userID string, tier Tier) error
	// IsWhitelisted checks whether a user has unlimited access override.
	IsWhitelisted(ctx context.Context, userID string) (bool, error)
	// AddToWhitelist grants unlimited access to a user.
	AddToWhitelist(ctx context.Context, entry WhitelistEntry) error
	// RemoveFromWhitelist revokes unlimited access from a user.
	RemoveFromWhitelist(ctx context.Context, userID string) error
	// ListWhitelist returns all current whitelist entries.
	ListWhitelist(ctx context.Context) ([]WhitelistEntry, error)
	// GetCostReport generates an aggregated cost report for a tenant.
	GetCostReport(ctx context.Context, tenantID string, from, to time.Time) (*CostReport, error)
	// AcquireConcurrency reserves a concurrency slot and returns a release function.
	AcquireConcurrency(ctx context.Context, userID string) (release func(), err error)
}
