// Package online provides a license validator that communicates with a
// remote license server over HTTP.
package online

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/EthanShen10086/voxera-kit/license"
)

// Adapter implements license.Manager by delegating to a remote license server.
type Adapter struct {
	serverURL  string
	httpClient *http.Client
}

// NewAdapter creates an online license adapter targeting the given license server URL.
func NewAdapter(serverURL string) *Adapter {
	return &Adapter{
		serverURL: serverURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

type validateRequest struct {
	Key string `json:"key"`
}

type validateResponse struct {
	ID        string   `json:"id"`
	TenantID  string   `json:"tenant_id"`
	Type      string   `json:"type"`
	Features  []string `json:"features"`
	MaxUsers  int      `json:"max_users"`
	IssuedAt  int64    `json:"issued_at"`
	ExpiresAt int64    `json:"expires_at"`
	Signature string   `json:"signature"`
	Valid     bool     `json:"valid"`
	Error     string   `json:"error"`
}

// Validate sends the license key to the remote server for verification.
func (a *Adapter) Validate(ctx context.Context, key string) (*license.License, error) {
	resp, err := a.postJSON(ctx, a.serverURL+"/validate", validateRequest{Key: key})
	if err != nil {
		return nil, fmt.Errorf("online: validation request failed: %w", err)
	}
	if !resp.Valid {
		if resp.Error != "" {
			return nil, errors.New("online: " + resp.Error)
		}
		return nil, license.ErrInvalidLicense
	}
	return a.toLicense(resp), nil
}

// Refresh requests a renewed license from the remote server.
func (a *Adapter) Refresh(ctx context.Context, key string) (*license.License, error) {
	resp, err := a.postJSON(ctx, a.serverURL+"/refresh", validateRequest{Key: key})
	if err != nil {
		return nil, fmt.Errorf("online: refresh request failed: %w", err)
	}
	if !resp.Valid {
		if resp.Error != "" {
			return nil, errors.New("online: " + resp.Error)
		}
		return nil, license.ErrInvalidLicense
	}
	return a.toLicense(resp), nil
}

// Features validates the key and returns the licensed feature list.
func (a *Adapter) Features(ctx context.Context, key string) ([]string, error) {
	lic, err := a.Validate(ctx, key)
	if err != nil {
		return nil, err
	}
	return lic.Features, nil
}

// IsExpired validates the key and checks whether the license has expired.
func (a *Adapter) IsExpired(ctx context.Context, key string) (bool, error) {
	lic, err := a.Validate(ctx, key)
	if err != nil {
		return false, err
	}
	return time.Now().After(lic.ExpiresAt), nil
}

func (a *Adapter) postJSON(ctx context.Context, url string, body any) (*validateResponse, error) {
	data, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("online: server returned status %d", resp.StatusCode)
	}

	var result validateResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("online: failed to decode response: %w", err)
	}
	return &result, nil
}

func (a *Adapter) toLicense(resp *validateResponse) *license.License {
	return &license.License{
		ID:        resp.ID,
		TenantID:  resp.TenantID,
		Type:      license.Type(resp.Type),
		Features:  resp.Features,
		MaxUsers:  resp.MaxUsers,
		IssuedAt:  time.Unix(resp.IssuedAt, 0),
		ExpiresAt: time.Unix(resp.ExpiresAt, 0),
		Signature: resp.Signature,
	}
}
