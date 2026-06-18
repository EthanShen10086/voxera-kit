package memory

import (
	"testing"

	"github.com/EthanShen10086/voxera-kit/storage"
	"github.com/EthanShen10086/voxera-kit/storage/contract"
)

func TestMemoryObjectStoreContract(t *testing.T) {
	contract.RunObjectStoreContract(t, func(t *testing.T) storage.ObjectStore {
		return New(storage.Config{})
	})
}

func TestMemoryMultipartContract(t *testing.T) {
	contract.RunMultipartContract(t, func(t *testing.T) (storage.MultipartUploader, storage.ObjectStore) {
		a := New(storage.Config{})
		return a, a
	})
}

func TestMemoryVersioningContract(t *testing.T) {
	contract.RunVersioningContract(t, func(t *testing.T) (storage.VersionedObjectStore, storage.StorageAdmin, storage.ObjectStore) {
		a := New(storage.Config{})
		return a, a, a
	})
}
