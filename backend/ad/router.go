package ad

import (
	"context"
	"sort"
	"time"
)

type providerEntry struct {
	provider Provider
	priority int
}

// DefaultRouter selects the best available ad provider by priority order.
type DefaultRouter struct {
	cfg       Config
	tracker   EventTracker
	providers []providerEntry
}

// NewRouter creates a DefaultRouter with the given configuration and event tracker.
func NewRouter(cfg Config, tracker EventTracker) *DefaultRouter {
	return &DefaultRouter{
		cfg:     cfg,
		tracker: tracker,
	}
}

// RegisterProvider adds a provider at the given priority (lower number = higher priority).
func (r *DefaultRouter) RegisterProvider(p Provider, priority int) {
	r.providers = append(r.providers, providerEntry{provider: p, priority: priority})
	sort.Slice(r.providers, func(i, j int) bool {
		return r.providers[i].priority < r.providers[j].priority
	})
}

// Fetch selects an ad from the highest-priority available provider.
// It returns nil for paid users or minors when policy is "hide".
func (r *DefaultRouter) Fetch(ctx context.Context, req Request) (*Ad, error) {
	if req.IsPaidUser {
		return nil, nil
	}
	if req.IsMinor && r.cfg.MinorPolicy == "hide" {
		return nil, nil
	}

	for _, entry := range r.providers {
		if !entry.provider.Available(ctx) {
			continue
		}
		result, err := entry.provider.FetchAd(ctx, req)
		if err != nil {
			continue
		}
		if result != nil && r.tracker != nil {
			_ = r.tracker.TrackImpression(ctx, result, req.UserID)
		}
		return result, nil
	}

	return r.fallbackAd(), nil
}

func (r *DefaultRouter) fallbackAd() *Ad {
	if r.cfg.FallbackHTML == "" {
		return nil
	}
	return &Ad{
		ID:           "fallback",
		SlotType:     SlotBanner,
		ProviderName: "fallback",
		Content: Content{
			Type:    "html",
			Payload: r.cfg.FallbackHTML,
		},
		ExpiresAt: time.Now().Add(time.Hour),
	}
}
