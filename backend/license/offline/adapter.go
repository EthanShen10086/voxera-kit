// Package offline provides an offline license validator that uses RSA
// signature verification without requiring network access.
package offline

import (
	"context"
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"time"

	"github.com/EthanShen10086/voxera-kit/license"
)

// Adapter implements license.Manager using offline RSA signature verification.
type Adapter struct {
	publicKey *rsa.PublicKey
}

// NewAdapter creates an offline license adapter by parsing the provided PEM-encoded
// RSA public key.
func NewAdapter(publicKeyPEM []byte) (*Adapter, error) {
	block, _ := pem.Decode(publicKeyPEM)
	if block == nil {
		return nil, errors.New("offline: failed to decode PEM block")
	}
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("offline: failed to parse public key: %w", err)
	}
	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("offline: key is not an RSA public key")
	}
	return &Adapter{publicKey: rsaPub}, nil
}

// licensePayload is the JSON structure embedded in the license key (before signing).
type licensePayload struct {
	ID        string   `json:"id"`
	TenantID  string   `json:"tenant_id"`
	Type      string   `json:"type"`
	Features  []string `json:"features"`
	MaxUsers  int      `json:"max_users"`
	IssuedAt  int64    `json:"issued_at"`
	ExpiresAt int64    `json:"expires_at"`
	Signature string   `json:"signature"`
}

// Validate decodes and verifies the license key, returning the license if valid.
func (a *Adapter) Validate(_ context.Context, key string) (*license.License, error) {
	data, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		return nil, license.ErrInvalidLicense
	}

	var payload licensePayload
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, license.ErrInvalidLicense
	}

	if err := a.verifySignature(payload); err != nil {
		return nil, license.ErrInvalidLicense
	}

	lic := &license.License{
		ID:        payload.ID,
		TenantID:  payload.TenantID,
		Type:      license.Type(payload.Type),
		Features:  payload.Features,
		MaxUsers:  payload.MaxUsers,
		IssuedAt:  time.Unix(payload.IssuedAt, 0),
		ExpiresAt: time.Unix(payload.ExpiresAt, 0),
		Signature: payload.Signature,
	}

	if time.Now().After(lic.ExpiresAt) {
		return nil, license.ErrExpired
	}

	return lic, nil
}

// Refresh is not supported in offline mode and always returns an error.
func (a *Adapter) Refresh(_ context.Context, _ string) (*license.License, error) {
	return nil, errors.New("offline: refresh is not supported in offline mode")
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
	_, err := a.Validate(ctx, key)
	if err != nil {
		if errors.Is(err, license.ErrExpired) {
			return true, nil
		}
		return false, err
	}
	return false, nil
}

func (a *Adapter) verifySignature(payload licensePayload) error {
	sig, err := base64.StdEncoding.DecodeString(payload.Signature)
	if err != nil {
		return license.ErrInvalidLicense
	}

	signable := licensePayload{
		ID:        payload.ID,
		TenantID:  payload.TenantID,
		Type:      payload.Type,
		Features:  payload.Features,
		MaxUsers:  payload.MaxUsers,
		IssuedAt:  payload.IssuedAt,
		ExpiresAt: payload.ExpiresAt,
	}
	data, err := json.Marshal(signable)
	if err != nil {
		return license.ErrInvalidLicense
	}

	hash := sha256.Sum256(data)
	if err := rsa.VerifyPKCS1v15(a.publicKey, crypto.SHA256, hash[:], sig); err != nil {
		return license.ErrInvalidLicense
	}
	return nil
}
