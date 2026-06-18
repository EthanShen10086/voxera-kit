package memory

import (
	"bytes"
	"context"
	"testing"
	"time"

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

func TestMemoryUploadLargeAndAdmin(t *testing.T) {
	ctx := context.Background()
	cfg := storage.Config{MultipartThreshold: 64}
	a := New(cfg, Options{EventPublisher: func(_, _ string) {}})
	defer func() { _ = a.Close() }()

	data := bytes.Repeat([]byte("m"), 128)
	if err := a.UploadLarge(ctx, "big.bin", bytes.NewReader(data), int64(len(data)), nil); err != nil {
		t.Fatal(err)
	}

	if err := a.PutLifecycleRules(ctx, []storage.LifecycleRule{
		{ID: "r1", Prefix: "big", Status: "Enabled", ExpirationDays: 1},
	}); err != nil {
		t.Fatal(err)
	}
	if err := a.DeleteLifecycleRules(ctx); err != nil {
		t.Fatal(err)
	}

	if err := a.PutBucketNotification(ctx, storage.NotificationDestination{
		Type: "webhook", Target: "https://example.com/hook",
	}); err != nil {
		t.Fatal(err)
	}
	if err := a.DeleteBucketNotification(ctx); err != nil {
		t.Fatal(err)
	}

	url, err := a.GetURL(ctx, "big.bin", time.Minute)
	if err != nil || url == "" {
		t.Fatalf("GetURL: %q err=%v", url, err)
	}
}
