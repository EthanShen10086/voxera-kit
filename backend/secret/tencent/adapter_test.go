package tencent_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/EthanShen10086/voxera-kit/secret"
	"github.com/EthanShen10086/voxera-kit/secret/contract"
	"github.com/EthanShen10086/voxera-kit/secret/tencent"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	ssm "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/ssm/v20190923"
)

type ssmStore struct {
	mu      sync.Mutex
	secrets map[string]string
}

func startSSMMock(t *testing.T) *ssm.Client {
	t.Helper()
	store := &ssmStore{secrets: make(map[string]string)}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		action := r.Header.Get("X-TC-Action")
		if action == "" {
			action = r.URL.Query().Get("Action")
		}
		body, _ := io.ReadAll(r.Body)
		var req map[string]any
		_ = json.Unmarshal(body, &req)

		w.Header().Set("Content-Type", "application/json")
		writeErr := func(code string) {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"Response": map[string]any{
					"Error":     map[string]string{"Code": code, "Message": code},
					"RequestId": "req",
				},
			})
		}

		switch action {
		case "GetSecretValue":
			name, _ := req["SecretName"].(string)
			store.mu.Lock()
			val, ok := store.secrets[name]
			store.mu.Unlock()
			if !ok {
				writeErr("ResourceNotFound.SecretNotFound")
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"Response": map[string]any{
					"SecretName":   name,
					"SecretString": val,
					"RequestId":    "req",
				},
			})
		case "CreateSecret":
			name, _ := req["SecretName"].(string)
			val, _ := req["SecretString"].(string)
			store.mu.Lock()
			store.secrets[name] = val
			store.mu.Unlock()
			_ = json.NewEncoder(w).Encode(map[string]any{"Response": map[string]any{"RequestId": "req"}})
		case "UpdateSecret":
			name, _ := req["SecretName"].(string)
			val, _ := req["SecretString"].(string)
			store.mu.Lock()
			store.secrets[name] = val
			store.mu.Unlock()
			_ = json.NewEncoder(w).Encode(map[string]any{"Response": map[string]any{"RequestId": "req"}})
		case "DeleteSecret":
			name, _ := req["SecretName"].(string)
			store.mu.Lock()
			_, ok := store.secrets[name]
			if ok {
				delete(store.secrets, name)
			}
			store.mu.Unlock()
			if !ok {
				writeErr("ResourceNotFound.SecretNotFound")
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"Response": map[string]any{"RequestId": "req"}})
		case "ListSecrets":
			store.mu.Lock()
			names := make([]string, 0, len(store.secrets))
			for n := range store.secrets {
				names = append(names, n)
			}
			store.mu.Unlock()
			metas := make([]map[string]any, 0, len(names))
			for _, n := range names {
				metas = append(metas, map[string]any{"SecretName": n})
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"Response": map[string]any{
					"SecretMetadatas": metas,
					"RequestId":       "req",
				},
			})
		default:
			w.WriteHeader(http.StatusBadRequest)
		}
	}))
	t.Cleanup(srv.Close)

	credential := common.NewCredential("test-id", "test-key")
	cpf := profile.NewClientProfile()
	host := strings.TrimPrefix(srv.URL, "http://")
	cpf.HttpProfile.Endpoint = host
	cpf.HttpProfile.Scheme = "http"
	client, err := ssm.NewClient(credential, "ap-guangzhou", cpf)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	return client
}

func TestNewManagerValidation(t *testing.T) {
	_, err := tencent.NewManager(tencent.Config{})
	if err == nil {
		t.Fatal("expected credentials error")
	}
}

func TestTencentSecretContract(t *testing.T) {
	contract.RunSecretContract(t, func(t *testing.T) (secret.Manager, func()) {
		client := startSSMMock(t)
		mgr, err := tencent.NewManager(tencent.Config{Client: client})
		if err != nil {
			t.Fatalf("NewManager: %v", err)
		}
		return mgr, nil
	})
}

func TestTencentGetNotFound(t *testing.T) {
	client := startSSMMock(t)
	mgr, err := tencent.NewManager(tencent.Config{Client: client})
	if err != nil {
		t.Fatal(err)
	}
	_, err = mgr.Get(context.Background(), "missing")
	if !errors.Is(err, secret.ErrNotFound) {
		t.Fatalf("Get() = %v", err)
	}
}

func TestTencentListWithPrefix(t *testing.T) {
	client := startSSMMock(t)
	mgr, err := tencent.NewManager(tencent.Config{Client: client})
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err := mgr.Set(ctx, "app/db", "1"); err != nil {
		t.Fatal(err)
	}
	if err := mgr.Set(ctx, "app/cache", "2"); err != nil {
		t.Fatal(err)
	}
	if err := mgr.Set(ctx, "other", "3"); err != nil {
		t.Fatal(err)
	}
	keys, err := mgr.List(ctx, "app/")
	if err != nil {
		t.Fatal(err)
	}
	if len(keys) != 2 {
		t.Fatalf("List() = %v", keys)
	}
}

func TestContextCancel(t *testing.T) {
	client := startSSMMock(t)
	mgr, err := tencent.NewManager(tencent.Config{Client: client})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if _, err := mgr.Get(ctx, "k"); !errors.Is(err, context.Canceled) {
		t.Fatalf("Get() = %v", err)
	}
	if err := mgr.Set(ctx, "k", "v"); !errors.Is(err, context.Canceled) {
		t.Fatalf("Set() = %v", err)
	}
	if err := mgr.Delete(ctx, "k"); !errors.Is(err, context.Canceled) {
		t.Fatalf("Delete() = %v", err)
	}
	if _, err := mgr.List(ctx, ""); !errors.Is(err, context.Canceled) {
		t.Fatalf("List() = %v", err)
	}
}
