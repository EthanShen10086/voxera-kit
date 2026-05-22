// Package memory provides an in-memory implementation of the aiquota.Manager interface.
package memory

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/EthanShen10086/voxera-kit/aiquota"
)

// Store is an in-memory implementation of aiquota.Manager.
type Store struct {
	mu          sync.RWMutex
	userTiers   map[string]aiquota.Tier
	usage       map[string]*aiquota.Usage
	whitelist   map[string]aiquota.WhitelistEntry
	records     []aiquota.UsageRecord
	concurrency map[string]*int32
}

// NewStore creates a new in-memory quota store.
func NewStore() *Store {
	return &Store{
		userTiers:   make(map[string]aiquota.Tier),
		usage:       make(map[string]*aiquota.Usage),
		whitelist:   make(map[string]aiquota.WhitelistEntry),
		records:     make([]aiquota.UsageRecord, 0),
		concurrency: make(map[string]*int32),
	}
}

// CheckQuota verifies the user has sufficient quota for the request.
func (s *Store) CheckQuota(_ context.Context, userID string, model string, estimatedTokens int) error {
	s.mu.RLock()
	_, whitelisted := s.whitelist[userID]
	s.mu.RUnlock()

	if whitelisted {
		return nil
	}

	quota := s.getQuotaForUser(userID)

	if !isModelAllowed(quota, model) {
		return aiquota.ErrModelNotAllowed
	}

	u := s.getOrCreateUsage(userID)

	s.mu.RLock()
	defer s.mu.RUnlock()

	s.maybeResetCounters(u)

	if quota.DailyTokens > 0 && u.DailyTokens+int64(estimatedTokens) > quota.DailyTokens {
		return aiquota.ErrQuotaExceeded
	}
	if quota.MonthlyTokens > 0 && u.MonthlyTokens+int64(estimatedTokens) > quota.MonthlyTokens {
		return aiquota.ErrQuotaExceeded
	}
	if quota.DailyRequests > 0 && u.DailyRequests >= quota.DailyRequests {
		return aiquota.ErrQuotaExceeded
	}

	return nil
}

// RecordUsage persists a completed AI call for metering and billing.
func (s *Store) RecordUsage(_ context.Context, record aiquota.UsageRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.records = append(s.records, record)

	u := s.getOrCreateUsageLocked(record.UserID)
	s.maybeResetCounters(u)

	tokens := int64(record.InputTokens + record.OutputTokens)
	u.DailyTokens += tokens
	u.MonthlyTokens += tokens
	u.DailyRequests++
	u.MonthlyRequests++
	u.TotalCostCents += record.CostCents

	return nil
}

// GetUsage retrieves current usage counters for a user.
func (s *Store) GetUsage(_ context.Context, userID string) (*aiquota.Usage, error) {
	u := s.getOrCreateUsage(userID)

	s.mu.RLock()
	defer s.mu.RUnlock()

	s.maybeResetCounters(u)

	cpy := *u
	return &cpy, nil
}

// GetQuota returns the quota configuration for a user.
func (s *Store) GetQuota(_ context.Context, userID string) (*aiquota.Quota, error) {
	q := s.getQuotaForUser(userID)
	return &q, nil
}

// SetTier changes the access tier for a user.
func (s *Store) SetTier(_ context.Context, userID string, tier aiquota.Tier) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.userTiers[userID] = tier

	if u, ok := s.usage[userID]; ok {
		u.Tier = tier
	}
	return nil
}

// IsWhitelisted checks whether a user has unlimited access override.
func (s *Store) IsWhitelisted(_ context.Context, userID string) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	_, ok := s.whitelist[userID]
	return ok, nil
}

// AddToWhitelist grants unlimited access to a user.
func (s *Store) AddToWhitelist(_ context.Context, entry aiquota.WhitelistEntry) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.whitelist[entry.UserID] = entry
	return nil
}

// RemoveFromWhitelist revokes unlimited access from a user.
func (s *Store) RemoveFromWhitelist(_ context.Context, userID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.whitelist, userID)
	return nil
}

// ListWhitelist returns all current whitelist entries.
func (s *Store) ListWhitelist(_ context.Context) ([]aiquota.WhitelistEntry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entries := make([]aiquota.WhitelistEntry, 0, len(s.whitelist))
	for _, e := range s.whitelist {
		entries = append(entries, e)
	}
	return entries, nil
}

// GetCostReport generates an aggregated cost report for a tenant.
func (s *Store) GetCostReport(_ context.Context, tenantID string, from, to time.Time) (*aiquota.CostReport, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	report := &aiquota.CostReport{
		TenantID: tenantID,
		Period:   from.Format(time.DateOnly) + " to " + to.Format(time.DateOnly),
		ByModel:  make(map[string]aiquota.ModelCost),
		ByUser:   make(map[string]int64),
	}

	for _, r := range s.records {
		if r.TenantID != tenantID {
			continue
		}
		if r.Timestamp.Before(from) || r.Timestamp.After(to) {
			continue
		}

		tokens := int64(r.InputTokens + r.OutputTokens)
		report.TotalRequests++
		report.TotalTokens += tokens
		report.TotalCostCents += r.CostCents

		mc := report.ByModel[r.Model]
		mc.Requests++
		mc.Tokens += tokens
		mc.Cost += r.CostCents
		report.ByModel[r.Model] = mc

		report.ByUser[r.UserID] += r.CostCents
	}

	return report, nil
}

// AcquireConcurrency reserves a concurrency slot and returns a release function.
func (s *Store) AcquireConcurrency(_ context.Context, userID string) (func(), error) {
	quota := s.getQuotaForUser(userID)

	s.mu.Lock()
	counter, ok := s.concurrency[userID]
	if !ok {
		var v int32
		counter = &v
		s.concurrency[userID] = counter
	}
	s.mu.Unlock()

	current := atomic.AddInt32(counter, 1)
	if quota.ConcurrentLimit > 0 && int(current) > quota.ConcurrentLimit {
		atomic.AddInt32(counter, -1)
		return nil, aiquota.ErrConcurrencyLimit
	}

	release := func() {
		atomic.AddInt32(counter, -1)
	}
	return release, nil
}

func (s *Store) getQuotaForUser(userID string) aiquota.Quota {
	s.mu.RLock()
	tier, ok := s.userTiers[userID]
	s.mu.RUnlock()

	if !ok {
		tier = aiquota.TierFree
	}

	quotas := aiquota.DefaultQuotas()
	if q, exists := quotas[tier]; exists {
		return q
	}
	return quotas[aiquota.TierFree]
}

func (s *Store) getOrCreateUsage(userID string) *aiquota.Usage {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.getOrCreateUsageLocked(userID)
}

func (s *Store) getOrCreateUsageLocked(userID string) *aiquota.Usage {
	u, ok := s.usage[userID]
	if !ok {
		now := time.Now()
		tier := s.userTiers[userID]
		if tier == "" {
			tier = aiquota.TierFree
		}
		u = &aiquota.Usage{
			UserID:           userID,
			Tier:             tier,
			LastResetDaily:   now,
			LastResetMonthly: now,
		}
		s.usage[userID] = u
	}
	return u
}

func (s *Store) maybeResetCounters(u *aiquota.Usage) {
	now := time.Now()

	y1, m1, d1 := u.LastResetDaily.Date()
	y2, m2, d2 := now.Date()
	if y1 != y2 || m1 != m2 || d1 != d2 {
		u.DailyTokens = 0
		u.DailyRequests = 0
		u.LastResetDaily = now
	}

	if u.LastResetMonthly.Year() != now.Year() || u.LastResetMonthly.Month() != now.Month() {
		u.MonthlyTokens = 0
		u.MonthlyRequests = 0
		u.LastResetMonthly = now
	}
}

func isModelAllowed(q aiquota.Quota, model string) bool {
	if len(q.AllowedModels) == 0 {
		return true
	}
	for _, m := range q.AllowedModels {
		if m == model {
			return true
		}
	}
	return false
}
