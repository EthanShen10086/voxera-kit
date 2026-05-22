// Package share defines the port interfaces for shareable link generation,
// resolution, and lifecycle management.
package share

import (
	"context"
	"time"
)

// Permission represents an access level granted by a share link.
type Permission string

// Share permission constants.
const (
	PermissionView     Permission = "view"
	PermissionDownload Permission = "download"
	PermissionEdit     Permission = "edit"
)

// Link represents a shareable link to a resource.
type Link struct {
	ID           string
	Token        string
	ResourceType string
	ResourceID   string
	CreatedBy    string
	Permissions  []Permission
	ExpiresAt    time.Time
	MaxUses      int
	UseCount     int
	CreatedAt    time.Time
	Metadata     map[string]string
}

// IsExpired reports whether the share link has passed its expiration time.
func (s *Link) IsExpired() bool {
	if s.ExpiresAt.IsZero() {
		return false
	}
	return time.Now().After(s.ExpiresAt)
}

// IsExhausted reports whether the share link has reached its maximum use count.
func (s *Link) IsExhausted() bool {
	if s.MaxUses <= 0 {
		return false
	}
	return s.UseCount >= s.MaxUses
}

// IsValid reports whether the share link is neither expired nor exhausted.
func (s *Link) IsValid() bool {
	return !s.IsExpired() && !s.IsExhausted()
}

// CreateShareRequest holds the parameters for creating a new share link.
type CreateShareRequest struct {
	ResourceType string
	ResourceID   string
	CreatedBy    string
	Permissions  []Permission
	ExpiresAt    time.Time
	MaxUses      int
	Metadata     map[string]string
}

// Generator creates share links from requests.
type Generator interface {
	Generate(ctx context.Context, req CreateShareRequest) (*Link, error)
}

// Repository persists and queries share links.
type Repository interface {
	Save(ctx context.Context, link *Link) error
	FindByToken(ctx context.Context, token string) (*Link, error)
	FindByID(ctx context.Context, id string) (*Link, error)
	ListByUser(ctx context.Context, userID string, limit, offset int) ([]*Link, error)
	IncrementUseCount(ctx context.Context, id string) error
	Revoke(ctx context.Context, id string) error
}

// Service orchestrates share link creation, resolution, and revocation.
type Service interface {
	CreateShare(ctx context.Context, req CreateShareRequest) (*Link, error)
	ResolveShare(ctx context.Context, token string) (*Link, error)
	RevokeShare(ctx context.Context, id string) error
	ListShares(ctx context.Context, userID string, limit, offset int) ([]*Link, error)
}
