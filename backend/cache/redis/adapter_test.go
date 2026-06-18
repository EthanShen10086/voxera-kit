package redis_test

import (
	"testing"

	"github.com/EthanShen10086/voxera-kit/cache"
	"github.com/EthanShen10086/voxera-kit/cache/contract"
	"github.com/EthanShen10086/voxera-kit/cache/local"
)

func TestAdapterContract(t *testing.T) {
	contract.RunCacheContract(t, func(t *testing.T) (cache.Cache, func()) {
		t.Helper()
		adapter, err := local.New(cache.Config{})
		if err != nil {
			t.Fatalf("local.New: %v", err)
		}
		return adapter, nil
	})
}

func TestAdapterContract_TestDouble(t *testing.T) {
	contract.RunCacheContract(t, func(t *testing.T) (cache.Cache, func()) {
		return contract.NewTestDouble(), nil
	})
}
