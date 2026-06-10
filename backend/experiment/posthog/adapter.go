// Package posthog provides a PostHog-backed implementation of the experiment
// Manager interface using PostHog's Feature Flags and Experiments API.
package posthog

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/EthanShen10086/voxera-kit/experiment"
)

// Adapter communicates with the PostHog Experiments API to manage A/B tests,
// variant assignments, and metric event capture.
type Adapter struct {
	apiKey    string
	endpoint  string
	projectID string
	client    *http.Client
}

// NewAdapter creates a new PostHog experiment adapter.
func NewAdapter(apiKey, endpoint, projectID string) *Adapter {
	return &Adapter{
		apiKey:    apiKey,
		endpoint:  endpoint,
		projectID: projectID,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Create registers a new experiment via the PostHog Experiments API.
func (a *Adapter) Create(ctx context.Context, cfg experiment.Config) error {
	payload := map[string]any{
		"name":             cfg.Name,
		"description":      cfg.Description,
		"feature_flag_key": cfg.Key,
		"parameters": map[string]any{
			"feature_flag_variants": buildVariants(cfg.Variants),
		},
	}

	url := fmt.Sprintf("%s/api/projects/%s/experiments", a.endpoint, a.projectID)

	return a.doPost(ctx, url, payload)
}

// Get retrieves an experiment by key. PostHog identifies experiments by ID
// rather than key, so this performs a list and filters client-side.
func (a *Adapter) Get(ctx context.Context, key string) (*experiment.Config, error) {
	all, err := a.List(ctx, "")
	if err != nil {
		return nil, err
	}

	for _, cfg := range all {
		if cfg.Key == key {
			return &cfg, nil
		}
	}

	return nil, fmt.Errorf("experiment %q not found in PostHog", key)
}

// List returns experiments from PostHog, optionally filtered by status.
func (a *Adapter) List(ctx context.Context, status experiment.Status) ([]experiment.Config, error) {
	url := fmt.Sprintf("%s/api/projects/%s/experiments", a.endpoint, a.projectID)

	body, err := a.doGet(ctx, url)
	if err != nil {
		return nil, err
	}

	var resp struct {
		Results []struct {
			ID          int    `json:"id"`
			Name        string `json:"name"`
			Description string `json:"description"`
			StartDate   string `json:"start_date"`
			EndDate     string `json:"end_date"`
			FeatureFlag struct {
				Key string `json:"key"`
			} `json:"feature_flag"`
		} `json:"results"`
	}

	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("decoding PostHog response: %w", err)
	}

	var configs []experiment.Config
	for _, r := range resp.Results {
		cfg := experiment.Config{
			ID:          fmt.Sprintf("%d", r.ID),
			Key:         r.FeatureFlag.Key,
			Name:        r.Name,
			Description: r.Description,
			Status:      deriveStatus(r.StartDate, r.EndDate),
		}

		if status == "" || cfg.Status == status {
			configs = append(configs, cfg)
		}
	}

	return configs, nil
}

// Start is a no-op; PostHog experiments are started via the dashboard or by
// setting a start date.
func (a *Adapter) Start(_ context.Context, _ string) error {
	return nil
}

// Stop is a no-op; PostHog experiments are stopped via the dashboard.
func (a *Adapter) Stop(_ context.Context, _ string) error {
	return nil
}

// Assign retrieves the variant assignment for a user from PostHog's feature
// flag evaluation endpoint. PostHog handles assignment automatically.
func (a *Adapter) Assign(ctx context.Context, key string, userID string) (*experiment.Assignment, error) {
	url := fmt.Sprintf("%s/decide/?v=3", a.endpoint)

	payload := map[string]any{
		"api_key":     a.apiKey,
		"distinct_id": userID,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshaling decide request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("creating decide request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("PostHog decide request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading decide response: %w", err)
	}

	var result struct {
		FeatureFlags map[string]any `json:"featureFlags"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("decoding decide response: %w", err)
	}

	variant, ok := result.FeatureFlags[key]
	if !ok {
		return nil, fmt.Errorf("experiment %q not found in PostHog flags", key)
	}

	variantKey := "control"
	if s, ok := variant.(string); ok {
		variantKey = s
	}

	return &experiment.Assignment{
		ExperimentKey: key,
		UserID:        userID,
		VariantKey:    variantKey,
		AssignedAt:    time.Now(),
	}, nil
}

// GetAssignment retrieves the current assignment by delegating to Assign, since
// PostHog assignments are deterministic.
func (a *Adapter) GetAssignment(ctx context.Context, key string, userID string) (*experiment.Assignment, error) {
	return a.Assign(ctx, key, userID)
}

// RecordMetric captures a metric event via the PostHog /capture endpoint.
func (a *Adapter) RecordMetric(ctx context.Context, key string, userID string, metricKey string, value float64) error {
	url := fmt.Sprintf("%s/capture", a.endpoint)

	payload := map[string]any{
		"api_key":     a.apiKey,
		"distinct_id": userID,
		"event":       metricKey,
		"properties": map[string]any{
			"experiment_key": key,
			"metric_key":     metricKey,
			"value":          value,
			"$lib":           "voxera-kit",
		},
	}

	return a.doPost(ctx, url, payload)
}

// GetResults fetches experiment results from the PostHog Experiments API. Full
// statistical analysis is performed server-side by PostHog.
func (a *Adapter) GetResults(ctx context.Context, key string) (*experiment.Result, error) {
	cfg, err := a.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/api/projects/%s/experiments/%s/results", a.endpoint, a.projectID, cfg.ID)

	body, err := a.doGet(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("fetching experiment results: %w", err)
	}

	var resp struct {
		Insight []any `json:"insight"`
	}

	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("decoding results response: %w", err)
	}

	return &experiment.Result{
		ExperimentKey: key,
		Status:        cfg.Status,
		AnalyzedAt:    time.Now(),
	}, nil
}

func (a *Adapter) doPost(ctx context.Context, url string, payload any) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshaling request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+a.apiKey)

	resp, err := a.client.Do(req)
	if err != nil {
		return fmt.Errorf("PostHog API request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= http.StatusBadRequest {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("PostHog API error (status %d): %s", resp.StatusCode, body)
	}

	return nil
}

func (a *Adapter) doGet(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+a.apiKey)

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("PostHog API request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode >= http.StatusBadRequest {
		return nil, fmt.Errorf("PostHog API error (status %d): %s", resp.StatusCode, body)
	}

	return body, nil
}

func buildVariants(variants []experiment.Variant) []map[string]any {
	result := make([]map[string]any, 0, len(variants))
	for _, v := range variants {
		result = append(result, map[string]any{
			"key":                v.Key,
			"name":               v.Name,
			"rollout_percentage": v.Weight,
		})
	}
	return result
}

func deriveStatus(startDate, endDate string) experiment.Status {
	if startDate == "" {
		return experiment.StatusDraft
	}
	if endDate != "" {
		return experiment.StatusComplete
	}
	return experiment.StatusRunning
}
