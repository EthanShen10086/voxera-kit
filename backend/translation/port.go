// Package translation defines the ports (interfaces) and domain types for
// multi-provider text translation, language detection, and batch operations.
package translation

import "context"

// TranslateOptions controls how a single translation request behaves.
type TranslateOptions struct {
	// SourceLang is the BCP-47 language code of the source text.
	// Leave empty to enable automatic detection.
	SourceLang string

	// TargetLang is the required BCP-47 language code for the output.
	TargetLang string

	// Formality adjusts the register of the translation (e.g. "formal", "informal").
	// Not every provider supports this; unsupported values are silently ignored.
	Formality string

	// GlossaryID references a provider-specific glossary for domain terminology.
	GlossaryID string
}

// TranslateResult holds the output of a single translation request.
type TranslateResult struct {
	// Text is the translated content.
	Text string

	// DetectedSourceLang is the language code the provider detected, if any.
	DetectedSourceLang string

	// Confidence expresses how certain the provider is about the detected language (0–1).
	Confidence float64
}

// BatchItem represents one text segment inside a batch translation request.
type BatchItem struct {
	Text       string
	SourceLang string
	TargetLang string
}

// Translator is the primary port that every translation adapter must satisfy.
type Translator interface {
	// Translate converts a single piece of text according to opts.
	Translate(ctx context.Context, text string, opts *TranslateOptions) (*TranslateResult, error)

	// TranslateBatch translates multiple items in one round-trip where possible.
	TranslateBatch(ctx context.Context, items []BatchItem) ([]TranslateResult, error)

	// DetectLanguage returns the most likely BCP-47 code and a confidence score.
	DetectLanguage(ctx context.Context, text string) (string, float64, error)

	// SupportedLanguages lists all target languages the provider can handle.
	SupportedLanguages(ctx context.Context) ([]string, error)

	// Close releases any resources held by the adapter.
	Close() error
}

// Config carries the credentials and tuning knobs shared by all adapters.
type Config struct {
	// APIKey is the provider credential.
	APIKey string

	// Endpoint overrides the default API base URL.
	Endpoint string

	// Model selects the translation model (relevant for LLM-based providers).
	Model string

	// MaxBatchSize caps the number of items sent in a single batch call.
	MaxBatchSize int
}
