package offline_test

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"testing"
	"time"

	"github.com/EthanShen10086/voxera-kit/license"
	"github.com/EthanShen10086/voxera-kit/license/offline"
)

func TestNewAdapter_InvalidPEM(t *testing.T) {
	_, err := offline.NewAdapter([]byte("not pem"))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestValidateAndRefresh(t *testing.T) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	pubDER, err := x509.MarshalPKIXPublicKey(&priv.PublicKey)
	if err != nil {
		t.Fatal(err)
	}
	pubPEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubDER})
	a, err := offline.NewAdapter(pubPEM)
	if err != nil {
		t.Fatal(err)
	}

	type payload struct {
		ID        string   `json:"id"`
		TenantID  string   `json:"tenant_id"`
		Type      string   `json:"type"`
		Features  []string `json:"features"`
		MaxUsers  int      `json:"max_users"`
		IssuedAt  int64    `json:"issued_at"`
		ExpiresAt int64    `json:"expires_at"`
		Signature string   `json:"signature"`
	}
	issued := time.Now().Unix()
	expires := time.Now().Add(time.Hour).Unix()
	signable := payload{
		ID: "lic-1", TenantID: "t1", Type: "pro",
		Features: []string{"feature-a"}, MaxUsers: 10,
		IssuedAt: issued, ExpiresAt: expires,
	}
	data, err := json.Marshal(signable)
	if err != nil {
		t.Fatal(err)
	}
	hash := sha256.Sum256(data)
	sig, err := rsa.SignPKCS1v15(rand.Reader, priv, crypto.SHA256, hash[:])
	if err != nil {
		t.Fatal(err)
	}
	signable.Signature = base64.StdEncoding.EncodeToString(sig)
	wire, err := json.Marshal(signable)
	if err != nil {
		t.Fatal(err)
	}
	key := base64.StdEncoding.EncodeToString(wire)

	lic, err := a.Validate(context.Background(), key)
	if err != nil || lic.ID != "lic-1" {
		t.Fatalf("Validate: %+v err=%v", lic, err)
	}
	features, err := a.Features(context.Background(), key)
	if err != nil || len(features) != 1 {
		t.Fatalf("Features: %v err=%v", features, err)
	}
	expired, err := a.IsExpired(context.Background(), key)
	if err != nil || expired {
		t.Fatalf("IsExpired: %v err=%v", expired, err)
	}
	_, err = a.Refresh(context.Background(), key)
	if err == nil {
		t.Fatal("expected refresh unsupported")
	}
	_, err = a.Validate(context.Background(), "!!!")
	if !errors.Is(err, license.ErrInvalidLicense) {
		t.Fatalf("invalid key: %v", err)
	}
}
