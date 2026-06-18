// Package google provides a Google Cloud Translation-backed implementation
// of [translation.Translator] using the REST v2 API.
package google

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/EthanShen10086/voxera-kit/translation"
)

const defaultEndpoint = "https://translation.googleapis.com"

type translateResponse struct {
	Data struct {
		Translations []struct {
			TranslatedText         string `json:"translatedText"`
			DetectedSourceLanguage string `json:"detectedSourceLanguage"`
		} `json:"translations"`
	} `json:"data"`
}

type detectResponse struct {
	Data struct {
		Detections [][]struct {
			Language string  `json:"language"`
			Score    float64 `json:"confidence"`
		} `json:"detections"`
	} `json:"data"`
}

// Adapter implements [translation.Translator] using Google Cloud Translation.
type Adapter struct {
	cfg    translation.Config
	client *http.Client
}

// New creates an [Adapter] with the supplied configuration.
func New(cfg translation.Config) *Adapter {
	if cfg.Endpoint == "" {
		cfg.Endpoint = defaultEndpoint
	}
	return &Adapter{
		cfg:    cfg,
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

// Translate converts text using Google Cloud Translation.
func (a *Adapter) Translate(ctx context.Context, text string, opts *translation.TranslateOptions) (*translation.TranslateResult, error) {
	if opts == nil || opts.TargetLang == "" {
		return nil, fmt.Errorf("google translate: TargetLang is required")
	}

	form := url.Values{}
	form.Set("q", text)
	form.Set("target", opts.TargetLang)
	if opts.SourceLang != "" {
		form.Set("source", opts.SourceLang)
	}

	var resp translateResponse
	if err := a.postForm(ctx, "/language/translate/v2", form, &resp); err != nil {
		return nil, err
	}
	if len(resp.Data.Translations) == 0 {
		return nil, fmt.Errorf("google translate: empty response")
	}
	tr := resp.Data.Translations[0]
	return &translation.TranslateResult{
		Text:               tr.TranslatedText,
		DetectedSourceLang: tr.DetectedSourceLanguage,
	}, nil
}

// TranslateBatch translates multiple items via Google Cloud Translation.
func (a *Adapter) TranslateBatch(ctx context.Context, items []translation.BatchItem) ([]translation.TranslateResult, error) {
	results := make([]translation.TranslateResult, len(items))
	for i, item := range items {
		res, err := a.Translate(ctx, item.Text, &translation.TranslateOptions{
			SourceLang: item.SourceLang,
			TargetLang: item.TargetLang,
		})
		if err != nil {
			return nil, err
		}
		results[i] = *res
	}
	return results, nil
}

// DetectLanguage identifies the language of text using Google.
func (a *Adapter) DetectLanguage(ctx context.Context, text string) (string, float64, error) {
	form := url.Values{}
	form.Set("q", text)

	var resp detectResponse
	if err := a.postForm(ctx, "/language/translate/v2/detect", form, &resp); err != nil {
		return "", 0, err
	}
	if len(resp.Data.Detections) == 0 || len(resp.Data.Detections[0]) == 0 {
		return "", 0, fmt.Errorf("google detect: empty response")
	}
	d := resp.Data.Detections[0][0]
	return d.Language, d.Score, nil
}

// SupportedLanguages returns the languages supported by Google Cloud Translation.
func (a *Adapter) SupportedLanguages(_ context.Context) ([]string, error) {
	return []string{
		"en", "zh", "zh-TW", "ja", "ko", "fr", "de", "es", "pt", "ru", "ar",
		"it", "nl", "pl", "tr", "vi", "th", "id", "hi",
	}, nil
}

// Close releases resources held by the Google adapter.
func (a *Adapter) Close() error { return nil }

func (a *Adapter) postForm(ctx context.Context, path string, form url.Values, out any) error {
	if a.cfg.APIKey == "" {
		return fmt.Errorf("google translate: APIKey is required")
	}
	endpoint := strings.TrimRight(a.cfg.Endpoint, "/") + path + "?key=" + url.QueryEscape(a.cfg.APIKey)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := a.client.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("google translate API (%d): %s", resp.StatusCode, string(body))
	}
	if err := json.Unmarshal(body, out); err != nil {
		return fmt.Errorf("google translate decode: %w", err)
	}
	return nil
}

var _ translation.Translator = (*Adapter)(nil)
