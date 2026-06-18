package minio_test

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/EthanShen10086/voxera-kit/storage"
	"github.com/EthanShen10086/voxera-kit/storage/internal/testfixture"
	miniostore "github.com/EthanShen10086/voxera-kit/storage/minio"
)

func TestMinIOWithGofakes3(t *testing.T) {
	srv := testfixture.StartS3(t, "minio-fake")
	cfg := srv.MinIOEndpoint()
	a, err := miniostore.New(cfg)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer func() { _ = a.Close() }()

	ctx := context.Background()
	if err := a.Upload(ctx, "probe.txt", bytes.NewReader([]byte("ok")), nil); err != nil {
		t.Skipf("minio-go incompatible with gofakes3: %v", err)
	}

	ok, err := a.Exists(ctx, "probe.txt")
	if err != nil || !ok {
		t.Fatalf("Exists() = %v, %v", ok, err)
	}

	items, err := a.List(ctx, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(items) == 0 {
		t.Fatal("expected list results")
	}

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

	if err := a.EnableVersioning(ctx, false); err != nil {
		t.Fatalf("disable versioning: %v", err)
	}
	enabled, err = a.GetVersioning(ctx)
	if err != nil || enabled {
		t.Fatalf("GetVersioning after disable = %v, %v", enabled, err)
	}
	if err := a.EnableVersioning(ctx, true); err != nil {
		t.Fatalf("re-enable versioning: %v", err)
	}

	if err := a.Upload(ctx, "probe.txt", bytes.NewReader([]byte("v2")), nil); err != nil {
		t.Fatal(err)
	}
	versions, err := a.ListVersions(ctx, "probe.txt")
	if err != nil {
		t.Fatalf("ListVersions: %v", err)
	}
	if len(versions) == 0 {
		t.Fatal("expected versions")
	}

	uploadID, err := a.InitiateMultipartUpload(ctx, "big.bin", nil)
	if err != nil {
		t.Fatalf("InitiateMultipartUpload: %v", err)
	}
	etag, err := a.UploadPart(ctx, "big.bin", uploadID, 1, bytes.NewReader([]byte("part")), 4)
	if err != nil || etag == "" {
		t.Fatalf("UploadPart: etag=%q err=%v", etag, err)
	}
	if err := a.CompleteMultipartUpload(ctx, "big.bin", uploadID, []storage.CompletedPart{{PartNumber: 1, ETag: etag}}); err != nil {
		t.Fatalf("CompleteMultipartUpload: %v", err)
	}

	abortID, err := a.InitiateMultipartUpload(ctx, "abort.bin", nil)
	if err != nil {
		t.Fatalf("InitiateMultipartUpload abort: %v", err)
	}
	if err := a.AbortMultipartUpload(ctx, "abort.bin", abortID); err != nil {
		t.Fatalf("AbortMultipartUpload: %v", err)
	}

	if len(versions) > 0 {
		if err := a.RestoreVersion(ctx, "probe.txt", versions[0].VersionID); err != nil {
			t.Fatalf("RestoreVersion: %v", err)
		}
		if err := a.DeleteVersion(ctx, "probe.txt", versions[0].VersionID); err != nil {
			t.Fatalf("DeleteVersion: %v", err)
		}
	}

	if err := a.Delete(ctx, "probe.txt"); err != nil {
		t.Fatal(err)
	}
	ok, err = a.Exists(ctx, "probe.txt")
	if err != nil || ok {
		t.Fatalf("after delete Exists() = %v, %v", ok, err)
	}

	data := bytes.Repeat([]byte("z"), 2048)
	cfg.MultipartThreshold = 1024
	a2, err := miniostore.New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = a2.Close() }()
	if err := a2.UploadLarge(ctx, "large.bin", bytes.NewReader(data), int64(len(data)), nil); err != nil {
		t.Fatalf("UploadLarge: %v", err)
	}

	url, err := a2.GetURL(ctx, "large.bin", time.Minute)
	if err != nil || url == "" {
		t.Fatalf("GetURL: %q err=%v", url, err)
	}

	rc, err := a2.Download(ctx, "large.bin")
	if err != nil {
		t.Fatalf("Download: %v", err)
	}
	_ = rc.Close()
}
