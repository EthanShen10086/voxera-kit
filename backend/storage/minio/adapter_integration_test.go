//go:build integration

package minio_test

import (
	"bytes"
	"context"
	"io"
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
	ctx := context.Background()
	a := startMinIOAdapter(t)
	multipart, store := a, a
	t.Cleanup(func() { _ = store.Close() })

	// MinIO/S3 require non-terminal parts to be at least 5 MiB.
	key := "contract/multipart.bin"
	part1 := bytes.Repeat([]byte("a"), 5<<20)
	part2 := []byte("part-two")
	expected := append(append([]byte(nil), part1...), part2...)

	uploadID, err := multipart.InitiateMultipartUpload(ctx, key, &storage.UploadOptions{ContentType: "application/octet-stream"})
	if err != nil {
		t.Fatalf("initiate multipart: %v", err)
	}
	etag1, err := multipart.UploadPart(ctx, key, uploadID, 1, bytes.NewReader(part1), int64(len(part1)))
	if err != nil {
		t.Fatalf("upload part 1: %v", err)
	}
	etag2, err := multipart.UploadPart(ctx, key, uploadID, 2, bytes.NewReader(part2), int64(len(part2)))
	if err != nil {
		t.Fatalf("upload part 2: %v", err)
	}
	if err := multipart.CompleteMultipartUpload(ctx, key, uploadID, []storage.CompletedPart{
		{PartNumber: 1, ETag: etag1},
		{PartNumber: 2, ETag: etag2},
	}); err != nil {
		t.Fatalf("complete multipart: %v", err)
	}
	rc, err := store.Download(ctx, key)
	if err != nil {
		t.Fatalf("download: %v", err)
	}
	got, err := io.ReadAll(rc)
	_ = rc.Close()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, expected) {
		t.Fatalf("multipart content mismatch: len(got)=%d len(want)=%d", len(got), len(expected))
	}
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
