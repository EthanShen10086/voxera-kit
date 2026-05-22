// Package memory provides in-memory implementations of share.ShareRepository
// and share.ShareGenerator.
package memory

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"sync"
	"time"

	"github.com/EthanShen10086/voxera-kit/share"
)

// Repository implements share.ShareRepository using in-memory maps.
type Repository struct {
	mu      sync.RWMutex
	byID    map[string]*share.Link
	byToken map[string]*share.Link
	byUser  map[string][]*share.Link
}

// NewRepository creates a new in-memory share repository.
func NewRepository() *Repository {
	return &Repository{
		byID:    make(map[string]*share.Link),
		byToken: make(map[string]*share.Link),
		byUser:  make(map[string][]*share.Link),
	}
}

// Save persists a share link in memory.
func (r *Repository) Save(_ context.Context, link *share.Link) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.byID[link.ID] = link
	r.byToken[link.Token] = link
	r.byUser[link.CreatedBy] = append(r.byUser[link.CreatedBy], link)
	return nil
}

// FindByToken looks up a share link by its token.
func (r *Repository) FindByToken(_ context.Context, token string) (*share.Link, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	link, ok := r.byToken[token]
	if !ok {
		return nil, errors.New("share link not found")
	}
	return link, nil
}

// FindByID looks up a share link by its ID.
func (r *Repository) FindByID(_ context.Context, id string) (*share.Link, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	link, ok := r.byID[id]
	if !ok {
		return nil, errors.New("share link not found")
	}
	return link, nil
}

// ListByUser returns share links created by the given user with pagination.
func (r *Repository) ListByUser(_ context.Context, userID string, limit, offset int) ([]*share.Link, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	all := r.byUser[userID]
	if offset >= len(all) {
		return nil, nil
	}
	end := offset + limit
	if end > len(all) {
		end = len(all)
	}
	return all[offset:end], nil
}

// IncrementUseCount atomically increments the use counter for a share link.
func (r *Repository) IncrementUseCount(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	link, ok := r.byID[id]
	if !ok {
		return errors.New("share link not found")
	}
	link.UseCount++
	return nil
}

// Revoke invalidates a share link by setting its expiration to the past.
func (r *Repository) Revoke(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	link, ok := r.byID[id]
	if !ok {
		return errors.New("share link not found")
	}
	link.ExpiresAt = time.Now().Add(-1 * time.Second)
	return nil
}

// Generator implements share.ShareGenerator using random hex tokens.
type Generator struct{}

// NewGenerator creates a new share link generator.
func NewGenerator() *Generator { return &Generator{} }

// Generate creates a new share link from the given request.
func (g *Generator) Generate(_ context.Context, req share.CreateShareRequest) (*share.Link, error) {
	id := generateHex(16)
	token := generateHex(24)
	now := time.Now().UTC()
	return &share.Link{
		ID:           id,
		Token:        token,
		ResourceType: req.ResourceType,
		ResourceID:   req.ResourceID,
		CreatedBy:    req.CreatedBy,
		Permissions:  req.Permissions,
		ExpiresAt:    req.ExpiresAt,
		MaxUses:      req.MaxUses,
		CreatedAt:    now,
		Metadata:     req.Metadata,
	}, nil
}

func generateHex(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
