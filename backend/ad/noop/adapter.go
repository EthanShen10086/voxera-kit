// Package noop provides a no-operation ad adapter for disabled advertising
// or paid user scenarios.
package noop

import (
	"context"

	"github.com/EthanShen10086/voxera-kit/ad"
)

// Adapter implements ad.Provider as a no-op for paid users or disabled advertising.
type Adapter struct{}

// NewAdapter creates a no-op ad adapter.
func NewAdapter() *Adapter {
	return &Adapter{}
}

// Name returns the provider identifier.
func (a *Adapter) Name() string {
	return "noop"
}

// FetchAd always returns nil as this adapter serves no ads.
func (a *Adapter) FetchAd(_ context.Context, _ ad.Request) (*ad.Ad, error) {
	return nil, nil
}

// ReportImpression is a no-op.
func (a *Adapter) ReportImpression(_ context.Context, _ string, _ string) error {
	return nil
}

// ReportClick is a no-op.
func (a *Adapter) ReportClick(_ context.Context, _ string, _ string) error {
	return nil
}

// Available always returns false as this adapter intentionally serves no ads.
func (a *Adapter) Available(_ context.Context) bool {
	return false
}
