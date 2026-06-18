package contract

import (
	"testing"

	"github.com/EthanShen10086/voxera-kit/storage"
	"github.com/EthanShen10086/voxera-kit/storage/memory"
)

func TestObjectStoreContract_MemoryAdapter(t *testing.T) {
	RunObjectStoreContract(t, func(t *testing.T) storage.ObjectStore {
		return memory.New(storage.Config{})
	})
}

func TestMultipartContract_MemoryAdapter(t *testing.T) {
	RunMultipartContract(t, func(t *testing.T) (storage.MultipartUploader, storage.ObjectStore) {
		a := memory.New(storage.Config{})
		return a, a
	})
}

func TestVersioningContract_MemoryAdapter(t *testing.T) {
	RunVersioningContract(t, func(t *testing.T) (storage.VersionedObjectStore, storage.StorageAdmin, storage.ObjectStore) {
		a := memory.New(storage.Config{})
		return a, a, a
	})
}
