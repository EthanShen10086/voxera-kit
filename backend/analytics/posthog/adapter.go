// Package posthog implements analytics.Store by delegating to a PostHog
// instance via its HTTP API (capture + query endpoints).
package posthog

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/EthanShen10086/voxera-kit/analytics"
)

// ErrNotSupported is returned for query operations that the PostHog HTTP
// API does not natively support.
var ErrNotSupported = errors.New("posthog: operation not supported")

// Adapter implements analytics.Store by delegating to a PostHog instance.
type Adapter struct {
	apiKey   string
	endpoint string
	client   *http.Client
}

// compile-time interface check.
var _ analytics.Store = (*Adapter)(nil)

// New creates a PostHog adapter targeting the given API key and endpoint
// (e.g. "https://app.posthog.com").
func New(apiKey, endpoint string) *Adapter {
	return &Adapter{
		apiKey:   apiKey,
		endpoint: endpoint,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Track sends a single event to PostHog via the capture endpoint.
func (a *Adapter) Track(ctx context.Context, event analytics.Event) error {
	payload := capturePayload{
		APIKey:     a.apiKey,
		Event:      event.Name,
		DistinctID: event.UserID,
		Timestamp:  event.Timestamp.Format(time.RFC3339),
		Properties: mergeContextProps(event),
	}
	return a.post(ctx, "/capture", payload)
}

// TrackBatch sends multiple events to PostHog in a single batch request.
func (a *Adapter) TrackBatch(ctx context.Context, events []analytics.Event) error {
	batch := make([]capturePayload, 0, len(events))
	for _, ev := range events {
		batch = append(batch, capturePayload{
			APIKey:     a.apiKey,
			Event:      ev.Name,
			DistinctID: ev.UserID,
			Timestamp:  ev.Timestamp.Format(time.RFC3339),
			Properties: mergeContextProps(ev),
		})
	}
	return a.post(ctx, "/batch", map[string]any{
		"api_key": a.apiKey,
		"batch":   batch,
	})
}

// Identify sends user properties to PostHog using the $identify event.
func (a *Adapter) Identify(ctx context.Context, profile analytics.UserProfile) error {
	payload := capturePayload{
		APIKey:     a.apiKey,
		Event:      "$identify",
		DistinctID: profile.UserID,
		Properties: map[string]any{
			"$set": profile.Properties,
		},
	}
	return a.post(ctx, "/capture", payload)
}

// Alias creates an alias linking previousID to newID in PostHog.
func (a *Adapter) Alias(ctx context.Context, previousID, newID string) error {
	payload := capturePayload{
		APIKey:     a.apiKey,
		Event:      "$create_alias",
		DistinctID: newID,
		Properties: map[string]any{
			"alias": previousID,
		},
	}
	return a.post(ctx, "/capture", payload)
}

// QueryFunnel queries PostHog's funnel insight API.
func (a *Adapter) QueryFunnel(ctx context.Context, q analytics.FunnelQuery) (*analytics.FunnelResult, error) {
	events := make([]map[string]any, 0, len(q.Steps))
	for i, s := range q.Steps {
		events = append(events, map[string]any{
			"id":    s.Name,
			"order": i,
			"type":  "events",
		})
	}
	body := map[string]any{
		"insight":            "FUNNELS",
		"events":             events,
		"date_from":          q.From.Format("2006-01-02"),
		"date_to":            q.To.Format("2006-01-02"),
		"funnel_window_days": max(q.WindowSec/86400, 1),
	}

	resp, err := a.queryInsight(ctx, body)
	if err != nil {
		return nil, fmt.Errorf("posthog funnel query: %w", err)
	}

	return parseFunnelResponse(resp)
}

// QueryRetention queries PostHog's retention insight API.
func (a *Adapter) QueryRetention(ctx context.Context, q analytics.RetentionQuery) (*analytics.RetentionResult, error) {
	body := map[string]any{
		"insight":         "RETENTION",
		"target_entity":   map[string]any{"id": q.CohortEvent, "type": "events"},
		"return_entity":   map[string]any{"id": q.ReturnEvent, "type": "events"},
		"date_from":       q.From.Format("2006-01-02"),
		"date_to":         q.To.Format("2006-01-02"),
		"retention_type":  "retention_first_time",
		"period":          string(q.Granularity),
		"total_intervals": q.Periods,
	}

	resp, err := a.queryInsight(ctx, body)
	if err != nil {
		return nil, fmt.Errorf("posthog retention query: %w", err)
	}

	return parseRetentionResponse(resp)
}

// QueryPath returns ErrNotSupported because PostHog does not expose a
// path-analysis query API.
func (a *Adapter) QueryPath(_ context.Context, _ analytics.PathQuery) (*analytics.PathResult, error) {
	return nil, ErrNotSupported
}

// QueryEvents fetches raw events from PostHog's events API.
func (a *Adapter) QueryEvents(ctx context.Context, q analytics.EventQuery) ([]analytics.Event, error) {
	url := a.endpoint + "/api/projects/@current/events?limit=" + fmt.Sprint(q.Limit)
	if q.UserID != "" {
		url += "&person_id=" + q.UserID
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("posthog create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+a.apiKey)

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("posthog events query: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("posthog events query: status %d: %s", resp.StatusCode, string(b))
	}

	var result struct {
		Results []struct {
			Event      string         `json:"event"`
			DistinctID string         `json:"distinct_id"`
			Timestamp  string         `json:"timestamp"`
			Properties map[string]any `json:"properties"`
		} `json:"results"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("posthog decode events: %w", err)
	}

	events := make([]analytics.Event, 0, len(result.Results))
	for _, r := range result.Results {
		ts, _ := time.Parse(time.RFC3339, r.Timestamp)
		events = append(events, analytics.Event{
			Name:       r.Event,
			UserID:     r.DistinctID,
			Timestamp:  ts,
			Properties: r.Properties,
		})
	}
	return events, nil
}

// QueryUserProfile fetches a person from PostHog's persons API.
func (a *Adapter) QueryUserProfile(ctx context.Context, userID string) (*analytics.UserProfile, error) {
	url := a.endpoint + "/api/projects/@current/persons?distinct_id=" + userID

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("posthog create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+a.apiKey)

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("posthog persons query: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("posthog persons query: status %d: %s", resp.StatusCode, string(b))
	}

	var result struct {
		Results []struct {
			DistinctIDs []string       `json:"distinct_ids"`
			Properties  map[string]any `json:"properties"`
			CreatedAt   string         `json:"created_at"`
		} `json:"results"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("posthog decode persons: %w", err)
	}
	if len(result.Results) == 0 {
		return nil, fmt.Errorf("posthog: user %q not found", userID)
	}

	person := result.Results[0]
	created, _ := time.Parse(time.RFC3339, person.CreatedAt)
	return &analytics.UserProfile{
		UserID:     userID,
		FirstSeen:  created,
		Properties: person.Properties,
	}, nil
}

// ---------------------------------------------------------------------------
// Internal helpers
// ---------------------------------------------------------------------------

type capturePayload struct {
	APIKey     string         `json:"api_key"`
	Event      string         `json:"event"`
	DistinctID string         `json:"distinct_id"`
	Timestamp  string         `json:"timestamp,omitempty"`
	Properties map[string]any `json:"properties,omitempty"`
}

func (a *Adapter) post(ctx context.Context, path string, payload any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("posthog marshal: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.endpoint+path, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("posthog create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		return fmt.Errorf("posthog request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("posthog: status %d: %s", resp.StatusCode, string(b))
	}

	// Drain the body to allow connection reuse.
	_, _ = io.Copy(io.Discard, resp.Body)
	return nil
}

func (a *Adapter) queryInsight(ctx context.Context, body map[string]any) (map[string]any, error) {
	data, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("posthog marshal: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		a.endpoint+"/api/projects/@current/insights/trend", bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("posthog create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+a.apiKey)

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("status %d: %s", resp.StatusCode, string(b))
	}

	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}
	return result, nil
}

func mergeContextProps(ev analytics.Event) map[string]any {
	props := make(map[string]any, len(ev.Properties)+8)
	for k, v := range ev.Properties {
		props[k] = v
	}
	if ev.Ctx.Platform != "" {
		props["$os"] = ev.Ctx.Platform
	}
	if ev.Ctx.PageURL != "" {
		props["$current_url"] = ev.Ctx.PageURL
	}
	if ev.Ctx.Referrer != "" {
		props["$referrer"] = ev.Ctx.Referrer
	}
	if ev.Ctx.UserAgent != "" {
		props["$user_agent"] = ev.Ctx.UserAgent
	}
	if ev.Ctx.ScreenWidth > 0 {
		props["$screen_width"] = ev.Ctx.ScreenWidth
		props["$screen_height"] = ev.Ctx.ScreenHeight
	}
	if ev.SessionID != "" {
		props["$session_id"] = ev.SessionID
	}
	return props
}

func parseFunnelResponse(resp map[string]any) (*analytics.FunnelResult, error) {
	result := &analytics.FunnelResult{}

	results, ok := resp["result"].([]any)
	if !ok {
		return result, nil
	}
	for _, r := range results {
		step, ok := r.(map[string]any)
		if !ok {
			continue
		}
		sr := analytics.FunnelStepResult{}
		if name, ok := step["name"].(string); ok {
			sr.Name = name
		}
		if count, ok := step["count"].(float64); ok {
			sr.EnteredCount = int64(count)
		}
		if rate, ok := step["conversion_rate"].(float64); ok {
			sr.ConvertRate = rate / 100
		}
		result.Steps = append(result.Steps, sr)
	}
	if len(result.Steps) > 0 && result.Steps[0].EnteredCount > 0 {
		last := result.Steps[len(result.Steps)-1]
		result.OverallRate = float64(last.EnteredCount) / float64(result.Steps[0].EnteredCount)
	}
	return result, nil
}

func parseRetentionResponse(resp map[string]any) (*analytics.RetentionResult, error) {
	result := &analytics.RetentionResult{}

	results, ok := resp["result"].([]any)
	if !ok {
		return result, nil
	}
	for _, r := range results {
		row, ok := r.(map[string]any)
		if !ok {
			continue
		}
		cr := analytics.CohortRow{}
		if label, ok := row["label"].(string); ok {
			cr.Period = label
		}
		if values, ok := row["values"].([]any); ok {
			for _, v := range values {
				vm, ok := v.(map[string]any)
				if !ok {
					continue
				}
				if count, ok := vm["count"].(float64); ok {
					cr.Retained = append(cr.Retained, int64(count))
				}
			}
		}
		if people, ok := row["people_url"].(string); ok {
			_ = people // available for follow-up queries
		}
		result.Cohorts = append(result.Cohorts, cr)
	}
	return result, nil
}
