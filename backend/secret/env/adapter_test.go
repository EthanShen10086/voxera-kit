package env_test

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/EthanShen10086/voxera-kit/secret"
	"github.com/EthanShen10086/voxera-kit/secret/contract"
	"github.com/EthanShen10086/voxera-kit/secret/env"
)

func TestEnvSecretContract(t *testing.T) {
	contract.RunSecretContract(t, func(t *testing.T) (secret.Manager, func()) {
		prefix := "VOXERA_ENV_" + strings.ReplaceAll(t.Name(), "/", "_")
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
