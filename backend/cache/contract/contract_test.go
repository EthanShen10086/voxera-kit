package contract

import (
	"testing"

	"github.com/EthanShen10086/voxera-kit/cache"
	"github.com/EthanShen10086/voxera-kit/cache/memory"
)

func TestCacheContract_Memory(t *testing.T) {
	RunCacheContract(t, func(t *testing.T) (cache.Cache, func()) {
		return memory.New(), nil
	})
}

func TestCacheContract_TestDouble(t *testing.T) {
	RunCacheContract(t, func(t *testing.T) (cache.Cache, func()) {
		return NewTestDouble(), nil
	})
}
