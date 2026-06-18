package contract

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/EthanShen10086/voxera-kit/secret"
	"github.com/EthanShen10086/voxera-kit/secret/env"
)

func TestSecretContract_Env(t *testing.T) {
	RunSecretContract(t, func(t *testing.T) (secret.Manager, func()) {
		prefix := "VOXERA_CONTRACT_" + strings.ReplaceAll(t.Name(), "/", "_")
		mgr := env.NewManager(prefix)
		return mgr, func() {
			ctx := context.Background()
			for _, key := range []string{"api-key", "delete-me", "prefix/a", "prefix/b", "other/c"} {
				_ = mgr.Delete(ctx, key)
			}
			_ = os.Unsetenv(prefix + "_LIST_SMOKE")
		}
	})
}
