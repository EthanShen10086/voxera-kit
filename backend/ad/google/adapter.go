// Package google provides a Google Ads adapter for the ad module.
// Integrate with Google AdSense REST API or AdMob for production use.
package google

import (
	"context"
	"errors"

	"github.com/EthanShen10086/voxera-kit/ad"
)

// Adapter implements ad.Provider using the Google Ads platform.
type Adapter struct {
	publisherID string
	apiKey      string
}

// NewAdapter creates a Google Ads adapter with the given publisher credentials.
func NewAdapter(publisherID, apiKey string) *Adapter {
	return &Adapter{
		publisherID: publisherID,
		apiKey:      apiKey,
	}
}

// Name returns the provider identifier.
func (a *Adapter) Name() string {
	return "google_ads"
}

// FetchAd retrieves an ad from Google Ads.
// TODO: integrate with Google AdSense REST API or AdMob.
func (a *Adapter) FetchAd(_ context.Context, _ ad.Request) (*ad.Ad, error) {
	return nil, errors.New("google: not yet implemented")
}

// ReportImpression reports an ad impression to Google Ads.
// TODO: integrate with Google AdSense REST API or AdMob.
func (a *Adapter) ReportImpression(_ context.Context, _ string, _ string) error {
	return nil
}

// ReportClick reports an ad click to Google Ads.
// TODO: integrate with Google AdSense REST API or AdMob.
func (a *Adapter) ReportClick(_ context.Context, _ string, _ string) error {
	return nil
}

// Available reports whether the adapter is configured and ready.
func (a *Adapter) Available(_ context.Context) bool {
	return a.publisherID != "" && a.apiKey != ""
}
