package contract

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/EthanShen10086/voxera-kit/storage"
)

// RunObjectStoreContract verifies basic object store operations.
func RunObjectStoreContract(t *testing.T, factory func(t *testing.T) storage.ObjectStore) {
	t.Helper()
	ctx := context.Background()
	store := factory(t)
	t.Cleanup(func() { _ = store.Close() })

	key := "contract/basic.txt"
	content := []byte("hello contract")
	if err := store.Upload(ctx, key, bytes.NewReader(content), &storage.UploadOptions{ContentType: "text/plain"}); err != nil {
		t.Fatalf("upload: %v", err)
	}

	exists, err := store.Exists(ctx, key)
	if err != nil {
		t.Fatalf("exists: %v", err)
	}
	if !exists {
		t.Fatal("expected object to exist")
	}

	rc, err := store.Download(ctx, key)
	if err != nil {
		t.Fatalf("download: %v", err)
	}
	got, err := io.ReadAll(rc)
	_ = rc.Close()
	if err != nil {
		t.Fatalf("read download: %v", err)
	}
	if !bytes.Equal(got, content) {
		t.Fatalf("content mismatch: got %q want %q", got, content)
	}

	items, err := store.List(ctx, "contract/")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	found := false
	for _, item := range items {
		if item.Key == key {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("list missing key %q: %#v", key, items)
	}

	url, err := store.GetURL(ctx, key, time.Minute)
	if err != nil {
		t.Fatalf("get url: %v", err)
	}
	if url == "" {
		t.Fatal("expected non-empty url")
	}

	if err := store.Delete(ctx, key); err != nil {
		t.Fatalf("delete: %v", err)
	}
	exists, err = store.Exists(ctx, key)
	if err != nil {
		t.Fatalf("exists after delete: %v", err)
	}
	if exists {
		t.Fatal("expected object to be deleted")
	}
}

// RunMultipartContract verifies multipart upload operations.
func RunMultipartContract(t *testing.T, factory func(t *testing.T) (storage.MultipartUploader, storage.ObjectStore)) {
	t.Helper()
	ctx := context.Background()
	multipart, store := factory(t)
	t.Cleanup(func() { _ = store.Close() })

	key := "contract/multipart.bin"
	part1 := []byte("part-one-")
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
	if etag1 == "" {
		t.Fatal("expected etag for part 1")
	}

	etag2, err := multipart.UploadPart(ctx, key, uploadID, 2, bytes.NewReader(part2), int64(len(part2)))
	if err != nil {
		t.Fatalf("upload part 2: %v", err)
	}
	if etag2 == "" {
		t.Fatal("expected etag for part 2")
	}

	if err := multipart.CompleteMultipartUpload(ctx, key, uploadID, []storage.CompletedPart{
		{PartNumber: 1, ETag: etag1},
		{PartNumber: 2, ETag: etag2},
	}); err != nil {
		t.Fatalf("complete multipart: %v", err)
	}

	rc, err := store.Download(ctx, key)
	if err != nil {
		t.Fatalf("download multipart object: %v", err)
	}
	got, err := io.ReadAll(rc)
	_ = rc.Close()
	if err != nil {
		t.Fatalf("read multipart object: %v", err)
	}
	if !bytes.Equal(got, expected) {
		t.Fatalf("multipart content mismatch: got %q want %q", got, expected)
	}

	abortID, err := multipart.InitiateMultipartUpload(ctx, "contract/abort.bin", nil)
	if err != nil {
		t.Fatalf("initiate abort upload: %v", err)
	}
	if err := multipart.AbortMultipartUpload(ctx, "contract/abort.bin", abortID); err != nil {
		t.Fatalf("abort multipart: %v", err)
	}
}

// RunVersioningContract verifies versioned object operations.
func RunVersioningContract(t *testing.T, factory func(t *testing.T) (storage.VersionedObjectStore, storage.StorageAdmin, storage.ObjectStore)) {
	t.Helper()
	ctx := context.Background()
	versioned, admin, store := factory(t)
	t.Cleanup(func() { _ = store.Close() })

	if err := admin.EnableVersioning(ctx, true); err != nil {
		t.Fatalf("enable versioning: %v", err)
	}
	enabled, err := admin.GetVersioning(ctx)
	if err != nil {
		t.Fatalf("get versioning: %v", err)
	}
	if !enabled {
		t.Fatal("expected versioning enabled")
	}

	key := "contract/versioned.txt"
	v1 := []byte("version-one")
	v2 := []byte("version-two")

	if err := store.Upload(ctx, key, bytes.NewReader(v1), nil); err != nil {
		t.Fatalf("upload v1: %v", err)
	}
	if err := store.Upload(ctx, key, bytes.NewReader(v2), nil); err != nil {
		t.Fatalf("upload v2: %v", err)
	}

	versions, err := versioned.ListVersions(ctx, key)
	if err != nil {
		t.Fatalf("list versions: %v", err)
	}
	if len(versions) < 2 {
		t.Fatalf("expected at least 2 versions, got %d", len(versions))
	}

	var firstVersionID string
	for _, v := range versions {
		if !v.IsLatest && !v.IsDeleteMarker {
			firstVersionID = v.VersionID
			break
		}
	}
	if firstVersionID == "" {
		t.Fatalf("could not find non-latest version: %#v", versions)
	}

	rc, err := versioned.DownloadVersion(ctx, key, firstVersionID)
	if err != nil {
		t.Fatalf("download version: %v", err)
	}
	got, err := io.ReadAll(rc)
	_ = rc.Close()
	if err != nil {
		t.Fatalf("read version: %v", err)
	}
	if !bytes.Equal(got, v1) {
		t.Fatalf("version content mismatch: got %q want %q", got, v1)
	}

	if err := versioned.RestoreVersion(ctx, key, firstVersionID); err != nil {
		t.Fatalf("restore version: %v", err)
	}
	rc, err = store.Download(ctx, key)
	if err != nil {
		t.Fatalf("download restored: %v", err)
	}
	got, err = io.ReadAll(rc)
	_ = rc.Close()
	if err != nil {
		t.Fatalf("read restored: %v", err)
	}
	if !bytes.Equal(got, v1) {
		t.Fatalf("restored content mismatch: got %q want %q", got, v1)
	}

	if err := store.Delete(ctx, key); err != nil {
		t.Fatalf("delete versioned object: %v", err)
	}
	exists, err := store.Exists(ctx, key)
	if err != nil {
		t.Fatalf("exists after delete marker: %v", err)
	}
	if exists {
		t.Fatal("expected object hidden after delete marker")
	}

	versions, err = versioned.ListVersions(ctx, key)
	if err != nil {
		t.Fatalf("list versions after delete: %v", err)
	}
	hasDeleteMarker := false
	for _, v := range versions {
		if v.IsDeleteMarker && v.IsLatest {
			hasDeleteMarker = true
		}
	}
	if !hasDeleteMarker {
		t.Fatalf("expected latest delete marker, got %#v", versions)
	}

	if err := versioned.DeleteVersion(ctx, key, firstVersionID); err != nil {
		t.Fatalf("delete version: %v", err)
	}
	versions, err = versioned.ListVersions(ctx, key)
	if err != nil {
		t.Fatalf("list versions after delete version: %v", err)
	}
	for _, v := range versions {
		if v.VersionID == firstVersionID {
			t.Fatalf("deleted version still listed: %#v", versions)
		}
	}

	if err := admin.PutLifecycleRules(ctx, []storage.LifecycleRule{
		{ID: "expire", Prefix: "contract/", Status: "Enabled", ExpirationDays: 30},
	}); err != nil {
		t.Fatalf("put lifecycle: %v", err)
	}
	rules, err := admin.GetLifecycleRules(ctx)
	if err != nil {
		t.Fatalf("get lifecycle: %v", err)
	}
	if len(rules) == 0 {
		t.Fatal("expected lifecycle rules")
	}

	if err := admin.PutBucketNotification(ctx, storage.NotificationDestination{
		Type:   "webhook",
		Target: "https://example.com/hook",
		Events: []storage.NotificationEvent{storage.EventObjectCreated},
	}); err != nil {
		t.Fatalf("put notification: %v", err)
	}
	cfg, err := admin.GetBucketNotification(ctx)
	if err != nil {
		t.Fatalf("get notification: %v", err)
	}
	if cfg == nil || cfg.Target == "" || !strings.Contains(cfg.Target, "example.com") {
		t.Fatalf("unexpected notification config: %#v", cfg)
	}
}
