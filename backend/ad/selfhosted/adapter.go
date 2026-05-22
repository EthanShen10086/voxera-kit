// Package selfhosted provides a self-hosted ad inventory adapter for the ad module.
package selfhosted

import (
	"context"
	"crypto/rand"
	"math/big"
	"sync/atomic"

	"github.com/EthanShen10086/voxera-kit/ad"
)

// Adapter implements ad.Provider using a local in-memory ad inventory.
type Adapter struct {
	inventory   []ad.Ad
	impressions atomic.Int64
	clicks      atomic.Int64
}

// NewAdapter creates a self-hosted adapter with the given ad inventory.
func NewAdapter(inventory []ad.Ad) *Adapter {
	return &Adapter{
		inventory: inventory,
	}
}

// Name returns the provider identifier.
func (a *Adapter) Name() string {
	return "self_hosted"
}

// FetchAd selects a random ad from the inventory using cryptographic randomness.
func (a *Adapter) FetchAd(_ context.Context, _ ad.Request) (*ad.Ad, error) {
	if len(a.inventory) == 0 {
		return nil, nil
	}
	n, err := rand.Int(rand.Reader, big.NewInt(int64(len(a.inventory))))
	if err != nil {
		return nil, err
	}
	result := a.inventory[n.Int64()]
	return &result, nil
}

// ReportImpression increments the impression counter for tracking.
func (a *Adapter) ReportImpression(_ context.Context, _ string, _ string) error {
	a.impressions.Add(1)
	return nil
}

// ReportClick increments the click counter for tracking.
func (a *Adapter) ReportClick(_ context.Context, _ string, _ string) error {
	a.clicks.Add(1)
	return nil
}

// Available reports whether the adapter has any inventory to serve.
func (a *Adapter) Available(_ context.Context) bool {
	return len(a.inventory) > 0
}

// Impressions returns the total number of recorded impressions.
func (a *Adapter) Impressions() int64 {
	return a.impressions.Load()
}

// Clicks returns the total number of recorded clicks.
func (a *Adapter) Clicks() int64 {
	return a.clicks.Load()
}
