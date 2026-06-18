// Package deepl provides a DeepL-backed implementation of [translation.Translator].
package deepl

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

const (
	defaultEndpoint     = "https://api.deepl.com"
	freeEndpoint        = "https://api-free.deepl.com"
	defaultFreeEndpoint = freeEndpoint
)

type translateResponse struct {
	Translations []struct {
		DetectedSourceLanguage string `json:"detected_source_language"`
		Text                   string `json:"text"`
	} `json:"translations"`
}

// Adapter implements [translation.Translator] using the DeepL API.
type Adapter struct {
	cfg    translation.Config
	client *http.Client
}

// New creates an [Adapter] with the supplied configuration.
func New(cfg translation.Config) *Adapter {
	if cfg.Endpoint == "" {
		cfg.Endpoint = defaultFreeEndpoint
	}
	return &Adapter{
		cfg:    cfg,
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

// Translate converts text using the DeepL API.
func (a *Adapter) Translate(ctx context.Context, text string, opts *translation.TranslateOptions) (*translation.TranslateResult, error) {
	if opts == nil || opts.TargetLang == "" {
		return nil, fmt.Errorf("deepl translate: TargetLang is required")
	}

	form := url.Values{}
	form.Set("text", text)
	form.Set("target_lang", strings.ToUpper(opts.TargetLang))
	if opts.SourceLang != "" {
		form.Set("source_lang", strings.ToUpper(opts.SourceLang))
	}
	if opts.Formality != "" {
		form.Set("formality", opts.Formality)
	}

	var resp translateResponse
	if err := a.postForm(ctx, "/v2/translate", form, &resp); err != nil {
		return nil, err
	}
	if len(resp.Translations) == 0 {
		return nil, fmt.Errorf("deepl translate: empty response")
	}
	tr := resp.Translations[0]
	return &translation.TranslateResult{
		Text:               tr.Text,
		DetectedSourceLang: strings.ToLower(tr.DetectedSourceLanguage),
	}, nil
}

// TranslateBatch translates multiple items via the DeepL API.
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

// DetectLanguage identifies the language of text using DeepL.
func (a *Adapter) DetectLanguage(ctx context.Context, text string) (string, float64, error) {
	res, err := a.Translate(ctx, text, &translation.TranslateOptions{TargetLang: "EN"})
	if err != nil {
		return "", 0, err
	}
	if res.DetectedSourceLang == "" {
		return "", 0, fmt.Errorf("deepl detect: no detected language")
	}
	return res.DetectedSourceLang, 0.95, nil
}

// SupportedLanguages returns the languages supported by DeepL.
func (a *Adapter) SupportedLanguages(_ context.Context) ([]string, error) {
	return []string{
		"bg", "cs", "da", "de", "el", "en", "es", "et", "fi", "fr", "hu", "id",
		"it", "ja", "ko", "lt", "lv", "nb", "nl", "pl", "pt", "ro", "ru", "sk",
		"sl", "sv", "tr", "uk", "zh",
	}, nil
}

// Close releases resources held by the DeepL adapter.
func (a *Adapter) Close() error { return nil }

func (a *Adapter) postForm(ctx context.Context, path string, form url.Values, out any) error {
	if a.cfg.APIKey == "" {
		return fmt.Errorf("deepl: APIKey is required")
	}
	endpoint := strings.TrimRight(a.cfg.Endpoint, "/") + path
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", "DeepL-Auth-Key "+a.cfg.APIKey)

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
		return fmt.Errorf("deepl API (%d): %s", resp.StatusCode, string(body))
	}
	if err := json.Unmarshal(body, out); err != nil {
		return fmt.Errorf("deepl decode: %w", err)
	}
	return nil
}

var _ translation.Translator = (*Adapter)(nil)
