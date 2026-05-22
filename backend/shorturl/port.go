// Package shorturl defines the port interface for short URL generation and resolution.
// It abstracts away the underlying storage and encoding implementation, allowing
// different backends to be used interchangeably.
package shorturl

import (
	"context"
	"time"
)

// ShortURL represents a shortened URL record.
type ShortURL struct {
	Code        string
	OriginalURL string
	CreatedAt   time.Time
	ExpiresAt   *time.Time
	ClickCount  int64
	CreatedBy   string
	Metadata    map[string]string
}

// GenerateOption is a functional option for configuring short URL generation.
type GenerateOption func(*generateOptions)

type generateOptions struct {
	expiry     time.Duration
	customCode string
	creator    string
	metadata   map[string]string
}

// WithExpiry sets the expiration duration for the short URL.
func WithExpiry(d time.Duration) GenerateOption {
	return func(o *generateOptions) {
		o.expiry = d
	}
}

// WithCustomCode sets a custom short code instead of generating one.
func WithCustomCode(code string) GenerateOption {
	return func(o *generateOptions) {
		o.customCode = code
	}
}

// WithCreator associates a creator ID with the short URL.
func WithCreator(id string) GenerateOption {
	return func(o *generateOptions) {
		o.creator = id
	}
}

// WithMetadata attaches arbitrary key-value metadata to the short URL.
func WithMetadata(m map[string]string) GenerateOption {
	return func(o *generateOptions) {
		o.metadata = m
	}
}

// GenerateParams holds resolved generation parameters after applying options.
type GenerateParams struct {
	Expiry     time.Duration
	CustomCode string
	Creator    string
	Metadata   map[string]string
}

// ResolveOptions processes a list of GenerateOption and returns the resolved parameters.
func ResolveOptions(opts []GenerateOption) GenerateParams {
	var o generateOptions
	for _, opt := range opts {
		opt(&o)
	}
	return GenerateParams{
		Expiry:     o.expiry,
		CustomCode: o.customCode,
		Creator:    o.creator,
		Metadata:   o.metadata,
	}
}

// Generator is the interface for short URL operations.
// Implementations must be safe for concurrent use.
type Generator interface {
	// Generate creates a new short URL for the given original URL.
	Generate(ctx context.Context, originalURL string, opts ...GenerateOption) (*ShortURL, error)
	// Resolve looks up the original URL by its short code.
	Resolve(ctx context.Context, code string) (*ShortURL, error)
	// Delete removes a short URL by its code.
	Delete(ctx context.Context, code string) error
	// IncrementClick atomically increments the click counter for the given code.
	IncrementClick(ctx context.Context, code string) error
	// ListByCreator returns short URLs created by the given user with pagination.
	ListByCreator(ctx context.Context, creatorID string, offset, limit int) ([]*ShortURL, error)
}

// Config holds configuration parameters for a short URL backend.
type Config struct {
	// BaseURL is the public-facing base URL prefix (e.g., "https://s.example.com").
	BaseURL string
	// CodeLength is the number of characters in the generated short code.
	CodeLength int
	// DefaultExpiry is the default expiration duration for short URLs.
	DefaultExpiry time.Duration
	// AllowCustomCode controls whether callers may specify custom short codes.
	AllowCustomCode bool
}
