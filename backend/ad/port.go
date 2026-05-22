// Package ad provides a pluggable advertising and monetization abstraction
// supporting multiple ad providers, slot types, and event tracking.
package ad

import (
	"context"
	"time"
)

// SlotType identifies the visual placement of an advertisement.
type SlotType string

const (
	// SlotBanner represents a standard banner ad placement.
	SlotBanner SlotType = "banner"
	// SlotInterstitial represents a full-screen interstitial ad.
	SlotInterstitial SlotType = "interstitial"
	// SlotNative represents a native in-feed ad placement.
	SlotNative SlotType = "native"
	// SlotRewarded represents a rewarded video ad placement.
	SlotRewarded SlotType = "rewarded"
	// SlotSidebar represents a sidebar ad placement.
	SlotSidebar SlotType = "sidebar"
)

// Content holds the renderable ad payload.
type Content struct {
	Type    string // "html", "image", "script", "json"
	Payload string
}

// Ad represents a single advertisement unit ready for display.
type Ad struct {
	ID            string
	SlotType      SlotType
	ProviderName  string
	Content       Content
	ClickURL      string
	ImpressionURL string
	ExpiresAt     time.Time
	Metadata      map[string]any
}

// Request describes what ad to fetch.
type Request struct {
	SlotType    SlotType
	UserID      string
	TenantID    string
	PageContext string
	Locale      string
	Tags        []string
	IsMinor     bool
	IsPaidUser  bool
}

// Provider fetches and reports on ads from an external or internal source.
type Provider interface {
	Name() string
	FetchAd(ctx context.Context, req Request) (*Ad, error)
	ReportImpression(ctx context.Context, adID string, userID string) error
	ReportClick(ctx context.Context, adID string, userID string) error
	Available(ctx context.Context) bool
}

// Router selects the appropriate Provider based on priority and availability.
type Router interface {
	Fetch(ctx context.Context, req Request) (*Ad, error)
	RegisterProvider(provider Provider, priority int)
}

// EventTracker persists ad interaction events for analytics and billing.
type EventTracker interface {
	TrackImpression(ctx context.Context, ad *Ad, userID string) error
	TrackClick(ctx context.Context, ad *Ad, userID string) error
}

// SlotConfig configures a single ad placement.
type SlotConfig struct {
	Enabled    bool
	MaxWidth   int
	MaxHeight  int
	RefreshSec int
}

// Config holds the overall advertising module configuration.
type Config struct {
	Enabled      bool
	MinorPolicy  string // "hide" or "safe_only"
	FrequencyCap int    // max impressions per user per hour
	FallbackHTML string // shown when no provider is available
	Slots        map[SlotType]SlotConfig
}
