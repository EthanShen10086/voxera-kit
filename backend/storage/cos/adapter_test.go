package cos_test

import (
	"context"
	"testing"

	"github.com/EthanShen10086/voxera-kit/storage"
	"github.com/EthanShen10086/voxera-kit/storage/contract"
	cosstore "github.com/EthanShen10086/voxera-kit/storage/cos"
	"github.com/EthanShen10086/voxera-kit/storage/internal/testfixture"
)

func newCOSAdapter(t *testing.T) *cosstore.Adapter {
	t.Helper()
	cfg := testfixture.StartCOSMock(t, "voxera-cos")
	a, err := cosstore.New(cfg)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return a
}

func TestCOSObjectStoreContract(t *testing.T) {
	contract.RunObjectStoreContract(t, func(t *testing.T) storage.ObjectStore {
		return newCOSAdapter(t)
	})
}

func TestCOSMultipartContract(t *testing.T) {
	contract.RunMultipartContract(t, func(t *testing.T) (storage.MultipartUploader, storage.ObjectStore) {
		a := newCOSAdapter(t)
		return a, a
	})
}

func TestCOSAdminStubs(t *testing.T) {
	ctx := context.Background()
	a := newCOSAdapter(t)
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
