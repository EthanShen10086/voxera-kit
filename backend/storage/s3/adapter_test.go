package s3_test

import (
	"bytes"
	"context"
	"io"
	"testing"
	"time"

	"github.com/EthanShen10086/voxera-kit/storage"
	"github.com/EthanShen10086/voxera-kit/storage/contract"
	s3store "github.com/EthanShen10086/voxera-kit/storage/s3"
	"github.com/EthanShen10086/voxera-kit/storage/internal/testfixture"
)

func newS3Adapter(t *testing.T) *s3store.Adapter {
	t.Helper()
	srv := testfixture.StartS3(t, "voxera-s3")
	a, err := s3store.New(srv.StorageConfig())
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return a
}

func TestS3ObjectStoreContract(t *testing.T) {
	contract.RunObjectStoreContract(t, func(t *testing.T) storage.ObjectStore {
		return newS3Adapter(t)
	})
}

func TestS3MultipartContract(t *testing.T) {
	contract.RunMultipartContract(t, func(t *testing.T) (storage.MultipartUploader, storage.ObjectStore) {
		a := newS3Adapter(t)
		return a, a
	})
}

func TestS3VersioningAndAdmin(t *testing.T) {
	ctx := context.Background()
	a := newS3Adapter(t)
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

	key := "admin/versioned.txt"
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

	if err := a.Delete(ctx, key); err != nil {
		t.Fatalf("Delete: %v", err)
	}
}

func TestS3DownloadVersion(t *testing.T) {
	ctx := context.Background()
	a := newS3Adapter(t)
	defer func() { _ = a.Close() }()

	if err := a.EnableVersioning(ctx, true); err != nil {
		t.Fatalf("EnableVersioning: %v", err)
	}

	key := "restore/me.txt"
	if err := a.Upload(ctx, key, bytes.NewReader([]byte("v1")), nil); err != nil {
		t.Fatal(err)
	}
	if err := a.Upload(ctx, key, bytes.NewReader([]byte("v2")), nil); err != nil {
		t.Fatal(err)
	}

	versions, err := a.ListVersions(ctx, key)
	if err != nil || len(versions) < 2 {
		t.Fatalf("ListVersions: %v err=%v", versions, err)
	}

	oldVersion := versions[0].VersionID
	rc, err := a.DownloadVersion(ctx, key, oldVersion)
	if err != nil {
		t.Fatalf("DownloadVersion: %v", err)
	}
	body, _ := io.ReadAll(rc)
	_ = rc.Close()
	if string(body) != "v1" {
		t.Fatalf("version body = %q", body)
	}
}

func TestS3PutBucketNotificationUnsupported(t *testing.T) {
	ctx := context.Background()
	a := newS3Adapter(t)
	defer func() { _ = a.Close() }()

	err := a.PutBucketNotification(ctx, storage.NotificationDestination{Type: "webhook"})
	if err == nil {
		t.Fatal("expected error for unsupported notification type")
	}
}

func TestS3UploadLarge(t *testing.T) {
	ctx := context.Background()
	a := newS3Adapter(t)
	defer func() { _ = a.Close() }()

	data := bytes.Repeat([]byte("x"), 6*1024*1024)
	key := "large/object.bin"
	if err := a.UploadLarge(ctx, key, bytes.NewReader(data), int64(len(data)), nil); err != nil {
		t.Fatalf("UploadLarge: %v", err)
	}

	url, err := a.GetURL(ctx, key, time.Minute)
	if err != nil || url == "" {
		t.Fatalf("GetURL: url=%q err=%v", url, err)
	}
}
