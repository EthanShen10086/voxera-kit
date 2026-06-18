package oss_test

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/EthanShen10086/voxera-kit/storage"
	"github.com/EthanShen10086/voxera-kit/storage/contract"
	"github.com/EthanShen10086/voxera-kit/storage/internal/testfixture"
	ossstore "github.com/EthanShen10086/voxera-kit/storage/oss"
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

func TestOSSLifecycleAndUploadLarge(t *testing.T) {
	ctx := context.Background()
	a := newOSSAdapter(t)
	defer func() { _ = a.Close() }()

	if err := a.PutLifecycleRules(ctx, []storage.LifecycleRule{
		{ID: "r1", Prefix: "tmp/", Status: "Enabled", ExpirationDays: 3},
	}); err != nil {
		t.Fatalf("PutLifecycleRules: %v", err)
	}
	rules, err := a.GetLifecycleRules(ctx)
	if err != nil {
		t.Fatalf("GetLifecycleRules: %v", err)
	}
	if len(rules) == 0 {
		t.Fatal("expected lifecycle rules")
	}
	if err := a.DeleteLifecycleRules(ctx); err != nil {
		t.Fatalf("DeleteLifecycleRules: %v", err)
	}

	data := bytes.Repeat([]byte("o"), 2048)
	cfg := testfixture.StartOSSMock(t, "oss-large")
	cfg.MultipartThreshold = 1024
	a2, err := ossstore.New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = a2.Close() }()
	if err := a2.UploadLarge(ctx, "large.bin", bytes.NewReader(data), int64(len(data)), nil); err != nil {
		t.Fatalf("UploadLarge: %v", err)
	}
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

func TestOSSExistsAndList(t *testing.T) {
	ctx := context.Background()
	a := newOSSAdapter(t)
	defer func() { _ = a.Close() }()

	ok, err := a.Exists(ctx, "missing")
	if err != nil || ok {
		t.Fatalf("Exists() = %v, %v", ok, err)
	}
	if err := a.Upload(ctx, "prefix/obj.bin", bytes.NewReader([]byte("data")), nil); err != nil {
		t.Fatal(err)
	}
	items, err := a.List(ctx, "prefix/")
	if err != nil || len(items) == 0 {
		t.Fatalf("List() = %#v, %v", items, err)
	}

	url, err := a.GetURL(ctx, "prefix/obj.bin", time.Minute)
	if err != nil || url == "" {
		t.Fatalf("GetURL: %q err=%v", url, err)
	}

	rc, err := a.Download(ctx, "prefix/obj.bin")
	if err != nil {
		t.Fatal(err)
	}
	_ = rc.Close()
}
