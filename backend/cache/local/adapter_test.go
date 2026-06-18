package local_test

import (
	"testing"

	"github.com/EthanShen10086/voxera-kit/cache"
	"github.com/EthanShen10086/voxera-kit/cache/contract"
	"github.com/EthanShen10086/voxera-kit/cache/local"
)

func TestLocalCacheContract(t *testing.T) {
	contract.RunCacheContract(t, func(t *testing.T) (cache.Cache, func()) {
		c, err := local.New(cache.Config{})
		if err != nil {
			t.Fatalf("New: %v", err)
		}
		return c, nil
	})
}
