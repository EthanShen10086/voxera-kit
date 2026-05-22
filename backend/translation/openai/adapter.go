// Package openai provides an OpenAI-backed implementation of [translation.Translator].
//
// This adapter uses the chat completions API to perform translation,
// language detection, and batch operations via prompt engineering.
package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/EthanShen10086/voxera-kit/translation"
)

const defaultEndpoint = "https://api.openai.com"
const defaultModel = "gpt-4o-mini"

type chatRequest struct {
	Model    string        `json:"model"`
	Messages []chatMessage `json:"messages"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

// Adapter implements [translation.Translator] using OpenAI's chat completions API.
type Adapter struct {
	cfg    translation.Config
	client *http.Client
}

// New creates an [Adapter] with the supplied configuration.
func New(cfg translation.Config) *Adapter {
	if cfg.Endpoint == "" {
		cfg.Endpoint = defaultEndpoint
	}
	if cfg.Model == "" {
		cfg.Model = defaultModel
	}
	return &Adapter{
		cfg:    cfg,
		client: &http.Client{Timeout: 60 * time.Second},
	}
}

// Translate converts text using the OpenAI chat completions API.
func (a *Adapter) Translate(ctx context.Context, text string, opts *translation.TranslateOptions) (*translation.TranslateResult, error) {
	systemPrompt := "You are a professional translator. Translate the user's text accurately, preserving tone and meaning. Output ONLY the translated text with no explanation."
	if opts.Formality != "" {
		systemPrompt += fmt.Sprintf(" Use a %s tone.", opts.Formality)
	}

	userPrompt := fmt.Sprintf("Translate the following text to %s", opts.TargetLang)
	if opts.SourceLang != "" {
		userPrompt += fmt.Sprintf(" from %s", opts.SourceLang)
	}
	userPrompt += ":\n\n" + text

	content, err := a.complete(ctx, systemPrompt, userPrompt)
	if err != nil {
		return nil, fmt.Errorf("openai translate: %w", err)
	}

	return &translation.TranslateResult{
		Text:               strings.TrimSpace(content),
		DetectedSourceLang: opts.SourceLang,
	}, nil
}

// TranslateBatch translates multiple items via the OpenAI API in a single call.
func (a *Adapter) TranslateBatch(ctx context.Context, items []translation.BatchItem) ([]translation.TranslateResult, error) {
	if len(items) == 0 {
		return nil, nil
	}

	systemPrompt := "You are a professional translator. You will receive numbered text items. " +
		"Translate each item to its specified target language. " +
		"Return ONLY the translations, one per line, prefixed with the same number. " +
		"Format: 1. translated text"

	var userPrompt strings.Builder
	for i, item := range items {
		fmt.Fprintf(&userPrompt, "%d. [to %s] %s\n", i+1, item.TargetLang, item.Text)
	}

	content, err := a.complete(ctx, systemPrompt, userPrompt.String())
	if err != nil {
		return nil, fmt.Errorf("openai batch translate: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(content), "\n")
	results := make([]translation.TranslateResult, len(items))

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		var idx int
		var translated string
		// Try "1. text" format
		if _, err := fmt.Sscanf(line, "%d.", &idx); err == nil {
			dotPos := strings.Index(line, ".")
			if dotPos >= 0 && dotPos+1 < len(line) {
				translated = strings.TrimSpace(line[dotPos+1:])
			}
		}

		if idx >= 1 && idx <= len(items) {
			results[idx-1] = translation.TranslateResult{
				Text:               translated,
				DetectedSourceLang: items[idx-1].SourceLang,
			}
		}
	}

	return results, nil
}

// DetectLanguage identifies the language of text using OpenAI.
func (a *Adapter) DetectLanguage(ctx context.Context, text string) (string, float64, error) {
	systemPrompt := "You are a language detection expert. Detect the language of the given text. " +
		"Respond with ONLY the BCP-47 language code (e.g. en, zh, ja, fr, de, es, ko, pt, ru, ar). Nothing else."

	content, err := a.complete(ctx, systemPrompt, text)
	if err != nil {
		return "", 0, fmt.Errorf("openai detect language: %w", err)
	}

	lang := strings.TrimSpace(strings.ToLower(content))
	return lang, 0.9, nil
}

// SupportedLanguages returns the languages supported by the OpenAI adapter.
func (a *Adapter) SupportedLanguages(_ context.Context) ([]string, error) {
	return []string{
		"en", "zh", "ja", "ko", "fr", "de", "es", "pt", "ru", "ar",
		"it", "nl", "pl", "tr", "vi", "th", "id", "ms", "hi", "bn",
		"sv", "da", "no", "fi", "cs", "sk", "ro", "hu", "uk", "he",
	}, nil
}

// Close releases resources held by the OpenAI adapter.
func (a *Adapter) Close() error { return nil }

func (a *Adapter) complete(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	reqBody := chatRequest{
		Model: a.cfg.Model,
		Messages: []chatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
	}

	payload, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	url := a.cfg.Endpoint + "/v1/chat/completions"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	if a.cfg.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+a.cfg.APIKey)
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("openai API error (%d): %s", resp.StatusCode, string(body))
	}

	var chatResp chatResponse
	if err = json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}

	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("openai returned no choices")
	}

	return chatResp.Choices[0].Message.Content, nil
}

// compile-time interface check
var _ translation.Translator = (*Adapter)(nil)
