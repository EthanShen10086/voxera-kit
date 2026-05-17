package share

import (
	"context"
	"time"
)

type Permission string

const (
	PermissionView     Permission = "view"
	PermissionDownload Permission = "download"
	PermissionEdit     Permission = "edit"
)

type ShareLink struct {
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

func (s *ShareLink) IsExpired() bool {
	if s.ExpiresAt.IsZero() {
		return false
	}
	return time.Now().After(s.ExpiresAt)
}

func (s *ShareLink) IsExhausted() bool {
	if s.MaxUses <= 0 {
		return false
	}
	return s.UseCount >= s.MaxUses
}

func (s *ShareLink) IsValid() bool {
	return !s.IsExpired() && !s.IsExhausted()
}

type CreateShareRequest struct {
	ResourceType string
	ResourceID   string
	CreatedBy    string
	Permissions  []Permission
	ExpiresAt    time.Time
	MaxUses      int
	Metadata     map[string]string
}

type ShareGenerator interface {
	Generate(ctx context.Context, req CreateShareRequest) (*ShareLink, error)
}

type ShareRepository interface {
	Save(ctx context.Context, link *ShareLink) error
	FindByToken(ctx context.Context, token string) (*ShareLink, error)
	FindByID(ctx context.Context, id string) (*ShareLink, error)
	ListByUser(ctx context.Context, userID string, limit, offset int) ([]*ShareLink, error)
	IncrementUseCount(ctx context.Context, id string) error
	Revoke(ctx context.Context, id string) error
}

type ShareService interface {
	CreateShare(ctx context.Context, req CreateShareRequest) (*ShareLink, error)
	ResolveShare(ctx context.Context, token string) (*ShareLink, error)
	RevokeShare(ctx context.Context, id string) error
	ListShares(ctx context.Context, userID string, limit, offset int) ([]*ShareLink, error)
}
