//go:build integration

package minio_test

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/EthanShen10086/voxera-kit/storage"
	"github.com/EthanShen10086/voxera-kit/storage/contract"
	miniostore "github.com/EthanShen10086/voxera-kit/storage/minio"
	"github.com/EthanShen10086/voxera-kit/testkit/containers"
)

func startMinIOAdapter(t *testing.T) *miniostore.Adapter {
	t.Helper()
	ctx := context.Background()
	c, err := containers.StartMinIO(ctx, "vm-"+strings.ReplaceAll(strings.ReplaceAll(t.Name(), "/", "-"), " ", ""))
	if err != nil {
		t.Fatalf("StartMinIO: %v", err)
	}
	t.Cleanup(func() { _ = c.Terminate(context.Background()) })

	a, err := miniostore.New(c.Config)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return a
}

func TestMinIOObjectStoreContract(t *testing.T) {
	contract.RunObjectStoreContract(t, func(t *testing.T) storage.ObjectStore {
		return startMinIOAdapter(t)
	})
}

func TestMinIOMultipartContract(t *testing.T) {
	contract.RunMultipartContract(t, func(t *testing.T) (storage.MultipartUploader, storage.ObjectStore) {
		a := startMinIOAdapter(t)
		return a, a
	})
}

func TestMinIOVersioningAndLifecycle(t *testing.T) {
	ctx := context.Background()
	a := startMinIOAdapter(t)
	defer func() { _ = a.Close() }()

	if err := a.EnableVersioning(ctx, true); err != nil {
		t.Fatalf("EnableVersioning: %v", err)
	}
	enabled, err := a.GetVersioning(ctx)
	if err != nil {
		t.Fatalf("GetVersioning: %v", err)
	}
	if !enabled {
		t.Fatal("expected versioning enabled")
	}

	key := "contract/versioned.txt"
	if err := a.Upload(ctx, key, bytes.NewReader([]byte("v1")), nil); err != nil {
		t.Fatalf("upload v1: %v", err)
	}
	if err := a.Upload(ctx, key, bytes.NewReader([]byte("v2")), nil); err != nil {
		t.Fatalf("upload v2: %v", err)
	}

	versions, err := a.ListVersions(ctx, key)
	if err != nil {
		t.Fatalf("ListVersions: %v", err)
	}
	if len(versions) < 2 {
		t.Fatalf("expected >=2 versions, got %d", len(versions))
	}

	if err := a.PutLifecycleRules(ctx, []storage.LifecycleRule{
		{ID: "expire", Prefix: "contract/", Status: "Enabled", ExpirationDays: 30},
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

	if err := a.PutBucketNotification(ctx, storage.NotificationDestination{Type: "webhook"}); err == nil {
		t.Fatal("expected PutBucketNotification error")
	}
}
