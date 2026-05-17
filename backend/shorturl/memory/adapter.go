// Package memory provides an in-memory implementation of the shorturl.ShortURLGenerator
// interface using map-based storage and base62 encoding.
package memory

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/EthanShen10086/voxera-kit/shorturl"
)

const base62Alphabet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

// Adapter is an in-memory short URL generator and resolver.
type Adapter struct {
	mu     sync.RWMutex
	urls   map[string]*shorturl.ShortURL
	config shorturl.ShortURLConfig
	rng    *rand.Rand
}

// New creates a new in-memory short URL adapter with the given configuration.
func New(config shorturl.ShortURLConfig) *Adapter {
	codeLen := config.CodeLength
	if codeLen <= 0 {
		codeLen = 6
	}
	config.CodeLength = codeLen
	return &Adapter{
		urls:   make(map[string]*shorturl.ShortURL),
		config: config,
		rng:    rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (a *Adapter) Generate(_ context.Context, originalURL string, opts ...shorturl.GenerateOption) (*shorturl.ShortURL, error) {
	params := shorturl.ResolveOptions(opts)

	code := params.CustomCode
	if code == "" {
		code = a.generateCode()
	} else if !a.config.AllowCustomCode {
		return nil, fmt.Errorf("shorturl: custom codes are not allowed")
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	if _, exists := a.urls[code]; exists {
		return nil, fmt.Errorf("shorturl: code %q already exists", code)
	}

	now := time.Now()
	su := &shorturl.ShortURL{
		Code:        code,
		OriginalURL: originalURL,
		CreatedAt:   now,
		CreatedBy:   params.Creator,
		Metadata:    params.Metadata,
	}

	if params.Expiry > 0 {
		exp := now.Add(params.Expiry)
		su.ExpiresAt = &exp
	} else if a.config.DefaultExpiry > 0 {
		exp := now.Add(a.config.DefaultExpiry)
		su.ExpiresAt = &exp
	}

	a.urls[code] = su
	return su, nil
}

func (a *Adapter) Resolve(_ context.Context, code string) (*shorturl.ShortURL, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	su, exists := a.urls[code]
	if !exists {
		return nil, fmt.Errorf("shorturl: code %q not found", code)
	}

	if su.ExpiresAt != nil && time.Now().After(*su.ExpiresAt) {
		return nil, fmt.Errorf("shorturl: code %q has expired", code)
	}

	return su, nil
}

func (a *Adapter) Delete(_ context.Context, code string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if _, exists := a.urls[code]; !exists {
		return fmt.Errorf("shorturl: code %q not found", code)
	}
	delete(a.urls, code)
	return nil
}

func (a *Adapter) IncrementClick(_ context.Context, code string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	su, exists := a.urls[code]
	if !exists {
		return fmt.Errorf("shorturl: code %q not found", code)
	}
	su.ClickCount++
	return nil
}

func (a *Adapter) ListByCreator(_ context.Context, creatorID string, offset, limit int) ([]*shorturl.ShortURL, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	var results []*shorturl.ShortURL
	for _, su := range a.urls {
		if su.CreatedBy == creatorID {
			results = append(results, su)
		}
	}

	if offset >= len(results) {
		return nil, nil
	}
	end := offset + limit
	if end > len(results) {
		end = len(results)
	}
	return results[offset:end], nil
}

func (a *Adapter) generateCode() string {
	b := make([]byte, a.config.CodeLength)
	for i := range b {
		b[i] = base62Alphabet[a.rng.Intn(len(base62Alphabet))]
	}
	return string(b)
}
