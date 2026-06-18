package memory_test

import (
	"testing"

	"github.com/EthanShen10086/voxera-kit/cache"
	"github.com/EthanShen10086/voxera-kit/cache/contract"
	"github.com/EthanShen10086/voxera-kit/cache/memory"
)

func TestMemoryCacheContract(t *testing.T) {
	contract.RunCacheContract(t, func(t *testing.T) (cache.Cache, func()) {
		return memory.New(), nil
	})
}
