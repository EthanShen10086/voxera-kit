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

type Repository struct {
	mu      sync.RWMutex
	byID    map[string]*share.ShareLink
	byToken map[string]*share.ShareLink
	byUser  map[string][]*share.ShareLink
}

func NewRepository() *Repository {
	return &Repository{
		byID:    make(map[string]*share.ShareLink),
		byToken: make(map[string]*share.ShareLink),
		byUser:  make(map[string][]*share.ShareLink),
	}
}

func (r *Repository) Save(_ context.Context, link *share.ShareLink) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.byID[link.ID] = link
	r.byToken[link.Token] = link
	r.byUser[link.CreatedBy] = append(r.byUser[link.CreatedBy], link)
	return nil
}

func (r *Repository) FindByToken(_ context.Context, token string) (*share.ShareLink, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	link, ok := r.byToken[token]
	if !ok {
		return nil, errors.New("share link not found")
	}
	return link, nil
}

func (r *Repository) FindByID(_ context.Context, id string) (*share.ShareLink, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	link, ok := r.byID[id]
	if !ok {
		return nil, errors.New("share link not found")
	}
	return link, nil
}

func (r *Repository) ListByUser(_ context.Context, userID string, limit, offset int) ([]*share.ShareLink, error) {
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

type Generator struct{}

func NewGenerator() *Generator { return &Generator{} }

func (g *Generator) Generate(_ context.Context, req share.CreateShareRequest) (*share.ShareLink, error) {
	id := generateHex(16)
	token := generateHex(24)
	now := time.Now().UTC()
	return &share.ShareLink{
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
