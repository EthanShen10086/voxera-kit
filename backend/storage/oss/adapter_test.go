package oss_test

import (
	"context"
	"testing"

	"github.com/EthanShen10086/voxera-kit/storage"
	"github.com/EthanShen10086/voxera-kit/storage/contract"
	ossstore "github.com/EthanShen10086/voxera-kit/storage/oss"
	"github.com/EthanShen10086/voxera-kit/storage/internal/testfixture"
)

func newOSSAdapter(t *testing.T) *ossstore.Adapter {
	t.Helper()
	cfg := testfixture.StartOSSMock(t, "voxera-oss")
	a, err := ossstore.New(cfg)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return a
}

func TestOSSObjectStoreContract(t *testing.T) {
	contract.RunObjectStoreContract(t, func(t *testing.T) storage.ObjectStore {
		return newOSSAdapter(t)
	})
}

func TestOSSMultipartContract(t *testing.T) {
	contract.RunMultipartContract(t, func(t *testing.T) (storage.MultipartUploader, storage.ObjectStore) {
		a := newOSSAdapter(t)
		return a, a
	})
}

func TestOSSAdminStubs(t *testing.T) {
	ctx := context.Background()
	a := newOSSAdapter(t)
	defer func() { _ = a.Close() }()

	if err := a.PutBucketNotification(ctx, storage.NotificationDestination{Type: "webhook"}); err == nil {
		t.Fatal("expected PutBucketNotification error")
	}
	cfg, err := a.GetBucketNotification(ctx)
	if err != nil || cfg != nil {
		t.Fatalf("GetBucketNotification: cfg=%#v err=%v", cfg, err)
	}
	if err := a.DeleteBucketNotification(ctx); err != nil {
		t.Fatalf("DeleteBucketNotification: %v", err)
	}

	enabled, err := a.GetVersioning(ctx)
	if err != nil {
		t.Fatalf("GetVersioning: %v", err)
	}
	if enabled {
		t.Fatal("expected versioning disabled on mock")
	}
	_, err = a.ListVersions(ctx, "any")
	if err != storage.ErrVersioningDisabled {
		t.Fatalf("ListVersions: %v", err)
	}
}
