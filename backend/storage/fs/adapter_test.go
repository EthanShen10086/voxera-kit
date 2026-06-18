package fs_test

import (
	"testing"

	"github.com/EthanShen10086/voxera-kit/storage"
	"github.com/EthanShen10086/voxera-kit/storage/contract"
	"github.com/EthanShen10086/voxera-kit/storage/fs"
)

func TestFSObjectStoreContract(t *testing.T) {
	contract.RunObjectStoreContract(t, func(t *testing.T) storage.ObjectStore {
		a, err := fs.New(t.TempDir(), storage.Config{})
		if err != nil {
			t.Fatalf("New: %v", err)
		}
		return a
	})
}

func TestFSMultipartContract(t *testing.T) {
	contract.RunMultipartContract(t, func(t *testing.T) (storage.MultipartUploader, storage.ObjectStore) {
		a, err := fs.New(t.TempDir(), storage.Config{})
		if err != nil {
			t.Fatalf("New: %v", err)
		}
		return a, a
	})
}
