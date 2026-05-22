// Package noop provides a no-op implementation of the aiquota.Manager interface
// that permits all requests without enforcement.
package noop

import (
	"context"
	"time"

	"github.com/EthanShen10086/voxera-kit/aiquota"
)

// Store is a no-op implementation that allows unlimited access.
type Store struct{}

// NewStore creates a new no-op quota store.
func NewStore() *Store {
	return &Store{}
}

// CheckQuota always returns nil, allowing all requests.
func (s *Store) CheckQuota(_ context.Context, _ string, _ string, _ int) error {
	return nil
}

// RecordUsage is a no-op that discards the usage record.
func (s *Store) RecordUsage(_ context.Context, _ aiquota.UsageRecord) error {
	return nil
}

// GetUsage returns an empty usage record.
func (s *Store) GetUsage(_ context.Context, userID string) (*aiquota.Usage, error) {
	return &aiquota.Usage{UserID: userID}, nil
}

// GetQuota returns a permissive unlimited quota.
func (s *Store) GetQuota(_ context.Context, _ string) (*aiquota.Quota, error) {
	return &aiquota.Quota{
		Tier:      aiquota.TierAdmin,
		OverQuota: aiquota.PolicyNotify,
	}, nil
}

// SetTier is a no-op.
func (s *Store) SetTier(_ context.Context, _ string, _ aiquota.Tier) error {
	return nil
}

// IsWhitelisted always returns true.
func (s *Store) IsWhitelisted(_ context.Context, _ string) (bool, error) {
	return true, nil
}

// AddToWhitelist is a no-op.
func (s *Store) AddToWhitelist(_ context.Context, _ aiquota.WhitelistEntry) error {
	return nil
}

// RemoveFromWhitelist is a no-op.
func (s *Store) RemoveFromWhitelist(_ context.Context, _ string) error {
	return nil
}

// ListWhitelist returns an empty list.
func (s *Store) ListWhitelist(_ context.Context) ([]aiquota.WhitelistEntry, error) {
	return nil, nil
}

// GetCostReport returns an empty cost report.
func (s *Store) GetCostReport(_ context.Context, tenantID string, from, to time.Time) (*aiquota.CostReport, error) {
	return &aiquota.CostReport{
		TenantID: tenantID,
		Period:   from.Format(time.DateOnly) + " to " + to.Format(time.DateOnly),
		ByModel:  make(map[string]aiquota.ModelCost),
		ByUser:   make(map[string]int64),
	}, nil
}

// AcquireConcurrency always succeeds and returns a no-op release function.
func (s *Store) AcquireConcurrency(_ context.Context, _ string) (func(), error) {
	return func() {}, nil
}
