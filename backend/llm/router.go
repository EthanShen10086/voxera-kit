package llm

import (
	"context"
	"fmt"
	"sort"
	"sync"
)

// entry pairs a provider with its routing priority (lower = higher priority).
type entry struct {
	provider Provider
	priority int
}

// Router dispatches chat and embedding requests to registered providers,
// supporting model-based routing and priority-based fallback.
type Router struct {
	mu      sync.RWMutex
	entries []entry
}

// NewRouter creates an empty router with no registered providers.
func NewRouter() *Router {
	return &Router{}
}

// Register adds a provider with the given priority. Lower priority values
// are tried first. If two providers share a priority, insertion order wins.
func (r *Router) Register(p Provider, priority int) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.entries = append(r.entries, entry{provider: p, priority: priority})
	sort.SliceStable(r.entries, func(i, j int) bool {
		return r.entries[i].priority < r.entries[j].priority
	})
}

// Route selects a provider for the request and executes a chat completion.
//
// Selection order:
//  1. If req.Model is set, pick the first available provider that lists
//     that model.
//  2. Otherwise, iterate providers by priority and use the first available one.
//  3. If the chosen provider fails, fall through to the next candidate.
func (r *Router) Route(ctx context.Context, req Request) (*Response, error) {
	providers, err := r.candidates(ctx, req.Model)
	if err != nil {
		return nil, err
	}
	var lastErr error
	for _, p := range providers {
		resp, chatErr := p.Chat(ctx, req)
		if chatErr == nil {
			return resp, nil
		}
		lastErr = chatErr
	}
	return nil, fmt.Errorf("all providers failed: %w", lastErr)
}

// RouteStream selects a provider and starts a streaming chat completion.
// Fallback logic mirrors Route.
func (r *Router) RouteStream(ctx context.Context, req Request) (<-chan StreamChunk, error) {
	providers, err := r.candidates(ctx, req.Model)
	if err != nil {
		return nil, err
	}
	var lastErr error
	for _, p := range providers {
		ch, streamErr := p.ChatStream(ctx, req)
		if streamErr == nil {
			return ch, nil
		}
		lastErr = streamErr
	}
	return nil, fmt.Errorf("all providers failed: %w", lastErr)
}

// ListAllModels aggregates model info from every registered provider.
func (r *Router) ListAllModels(ctx context.Context) []ModelInfo {
	r.mu.RLock()
	entries := make([]entry, len(r.entries))
	copy(entries, r.entries)
	r.mu.RUnlock()

	var all []ModelInfo
	for _, e := range entries {
		models, err := e.provider.ListModels(ctx)
		if err != nil {
			continue
		}
		all = append(all, models...)
	}
	return all
}

// candidates returns providers eligible for the given model, ordered by
// priority. When model is empty, all available providers are returned.
func (r *Router) candidates(ctx context.Context, model string) ([]Provider, error) {
	r.mu.RLock()
	entries := make([]entry, len(r.entries))
	copy(entries, r.entries)
	r.mu.RUnlock()

	if len(entries) == 0 {
		return nil, fmt.Errorf("no providers registered")
	}

	var out []Provider
	for _, e := range entries {
		if !e.provider.Available(ctx) {
			continue
		}
		if model != "" {
			if !r.providerHasModel(ctx, e.provider, model) {
				continue
			}
		}
		out = append(out, e.provider)
	}

	if len(out) == 0 {
		if model != "" {
			return nil, fmt.Errorf("no available provider supports model %q", model)
		}
		return nil, fmt.Errorf("no available providers")
	}
	return out, nil
}

// providerHasModel checks whether p advertises the given model ID.
func (r *Router) providerHasModel(ctx context.Context, p Provider, model string) bool {
	models, err := p.ListModels(ctx)
	if err != nil {
		return false
	}
	for _, m := range models {
		if m.ID == model {
			return true
		}
	}
	return false
}
