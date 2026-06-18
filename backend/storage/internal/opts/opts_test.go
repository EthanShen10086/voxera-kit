package opts_test

import (
	"testing"

	"github.com/EthanShen10086/voxera-kit/storage"
	"github.com/EthanShen10086/voxera-kit/storage/internal/opts"
)

func TestNormalizeKey(t *testing.T) {
	if got := opts.NormalizeKey("/path/key"); got != "path/key" {
		t.Fatalf("NormalizeKey = %q", got)
	}
}

func TestMergeUploadOptions(t *testing.T) {
	if got := opts.MergeUploadOptions(nil); got.ContentType != "" {
		t.Fatalf("MergeUploadOptions(nil) = %+v", got)
	}
	in := &storage.UploadOptions{ContentType: "text/plain"}
	if got := opts.MergeUploadOptions(in); got.ContentType != "text/plain" {
		t.Fatalf("MergeUploadOptions = %+v", got)
	}
}

func TestPartSizeAndThreshold(t *testing.T) {
	cfg := storage.Config{PartSize: 2048, MultipartThreshold: 4096}
	if opts.PartSize(cfg) != 2048 {
		t.Fatalf("PartSize = %d", opts.PartSize(cfg))
	}
	if opts.MultipartThreshold(cfg) != 4096 {
		t.Fatalf("MultipartThreshold = %d", opts.MultipartThreshold(cfg))
	}
	if opts.PartSize(storage.Config{}) != storage.DefaultPartSize {
		t.Fatal("expected default part size")
	}
	if opts.MultipartThreshold(storage.Config{}) != storage.DefaultMultipartThreshold {
		t.Fatal("expected default threshold")
	}
}
