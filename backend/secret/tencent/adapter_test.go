package tencent_test

import (
	"context"
	"os"
	"testing"

	"github.com/EthanShen10086/voxera-kit/secret"
	"github.com/EthanShen10086/voxera-kit/secret/contract"
	"github.com/EthanShen10086/voxera-kit/secret/tencent"
)

func TestSecretContract_Tencent(t *testing.T) {
	secretID := os.Getenv("TENCENT_SECRET_ID")
	secretKey := os.Getenv("TENCENT_SECRET_KEY")
	if secretID == "" || secretKey == "" {
		t.Skip("TENCENT_SECRET_ID and TENCENT_SECRET_KEY required for cloud integration")
	}

	contract.RunSecretContract(t, func(t *testing.T) (secret.Manager, func()) {
		mgr, err := tencent.NewManager(tencent.Config{
			SecretID:  secretID,
			SecretKey: secretKey,
			Region:    os.Getenv("TENCENT_REGION"),
		})
		if err != nil {
			t.Fatalf("NewManager: %v", err)
		}
		return mgr, func() {
			ctx := context.Background()
			for _, key := range []string{"api-key", "delete-me", "prefix/a", "prefix/b", "other/c"} {
				_ = mgr.Delete(ctx, key)
			}
		}
	})
}

func TestNewManager_MissingCredentials(t *testing.T) {
	_, err := tencent.NewManager(tencent.Config{})
	if err == nil {
		t.Fatal("expected error for missing credentials")
	}
}
