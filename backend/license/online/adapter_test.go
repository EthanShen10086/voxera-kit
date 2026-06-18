package online_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/EthanShen10086/voxera-kit/license"
	"github.com/EthanShen10086/voxera-kit/license/online"
)

func TestValidateAndFeatures(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id": "lic-1", "tenant_id": "t1", "type": "pro",
			"features": []string{"analytics"}, "max_users": 10,
			"issued_at": time.Now().Unix(), "expires_at": time.Now().Add(24 * time.Hour).Unix(),
			"signature": "sig", "valid": true,
		})
	}))
	defer srv.Close()

	a := online.NewAdapter(srv.URL)
	lic, err := a.Validate(context.Background(), "key")
	if err != nil || lic.ID != "lic-1" {
		t.Fatalf("Validate: %+v err=%v", lic, err)
	}
	features, err := a.Features(context.Background(), "key")
	if err != nil || len(features) != 1 {
		t.Fatalf("Features: %v err=%v", features, err)
	}
	expired, err := a.IsExpired(context.Background(), "key")
	if err != nil || expired {
		t.Fatalf("IsExpired: %v err=%v", expired, err)
	}
}

func TestValidateInvalid(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"valid": false, "error": "bad key"})
	}))
	defer srv.Close()

	a := online.NewAdapter(srv.URL)
	_, err := a.Validate(context.Background(), "bad")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestRefresh(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/refresh" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id": "lic-2", "valid": true, "expires_at": time.Now().Add(time.Hour).Unix(),
		})
	}))
	defer srv.Close()

	a := online.NewAdapter(srv.URL)
	lic, err := a.Refresh(context.Background(), "key")
	if err != nil || lic.ID != "lic-2" {
		t.Fatalf("Refresh: %+v err=%v", lic, err)
	}
}

func TestValidateServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	a := online.NewAdapter(srv.URL)
	_, err := a.Validate(context.Background(), "key")
	if err == nil {
		t.Fatal("expected error")
	}
	_ = license.ErrInvalidLicense
}
