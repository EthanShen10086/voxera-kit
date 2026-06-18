package vault_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/EthanShen10086/voxera-kit/secret"
	"github.com/EthanShen10086/voxera-kit/secret/contract"
	vaultadapter "github.com/EthanShen10086/voxera-kit/secret/vault"
)

type vaultStore struct {
	mu   sync.Mutex
	data map[string]string
}

func startVaultMock(t *testing.T) string {
	t.Helper()
	store := &vaultStore{data: make(map[string]string)}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/v1/secret/")
		switch {
		case r.Method == http.MethodGet && strings.HasPrefix(path, "data/"):
			key := strings.TrimPrefix(path, "data/")
			store.mu.Lock()
			val, ok := store.data[key]
			store.mu.Unlock()
			if !ok {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": map[string]any{"data": map[string]string{"value": val}},
			})
		case r.Method == http.MethodPost && strings.HasPrefix(path, "data/"):
			fallthrough
		case r.Method == http.MethodPut && strings.HasPrefix(path, "data/"):
			key := strings.TrimPrefix(path, "data/")
			var body struct {
				Data map[string]string `json:"data"`
			}
			_ = json.NewDecoder(r.Body).Decode(&body)
			store.mu.Lock()
			store.data[key] = body.Data["value"]
			store.mu.Unlock()
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": map[string]any{
					"created_time": "2024-01-01T00:00:00Z",
					"version":      1,
				},
			})
		case r.Method == http.MethodDelete && strings.HasPrefix(path, "metadata/"):
			key := strings.TrimPrefix(path, "metadata/")
			store.mu.Lock()
			delete(store.data, key)
			store.mu.Unlock()
			w.WriteHeader(http.StatusNoContent)
		case (r.Method == "LIST" || r.Method == http.MethodGet) && strings.HasPrefix(path, "metadata"):
			prefix := strings.TrimPrefix(path, "metadata/")
			prefix = strings.Trim(prefix, "/")
			store.mu.Lock()
			var keys []string
			for k := range store.data {
				if prefix == "" {
					keys = append(keys, k)
					continue
				}
				if strings.HasPrefix(k, prefix+"/") {
					keys = append(keys, strings.TrimPrefix(k, prefix+"/"))
				} else if k == prefix {
					keys = append(keys, k)
				}
			}
			store.mu.Unlock()
			list := make([]any, len(keys))
			for i, k := range keys {
				list[i] = k
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"keys": list}})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(srv.Close)
	return srv.URL
}

func TestVaultSecretContract(t *testing.T) {
	contract.RunSecretContract(t, func(t *testing.T) (secret.Manager, func()) {
		mgr, err := vaultadapter.NewManager(vaultadapter.Config{
			Address: startVaultMock(t),
			Token:   "test-token",
			Mount:   "secret",
		})
		if err != nil {
			t.Fatalf("NewManager: %v", err)
		}
		return mgr, nil
	})
}

func TestVaultGetNotFound(t *testing.T) {
	mgr, err := vaultadapter.NewManager(vaultadapter.Config{
		Address: startVaultMock(t),
		Token:   "t",
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = mgr.Get(context.Background(), "missing")
	if err != secret.ErrNotFound {
		t.Fatalf("Get() = %v", err)
	}
}
