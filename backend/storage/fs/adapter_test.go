package fs_test

import (
	"bytes"
	"context"
	"io"
	"testing"
	"time"

	"github.com/EthanShen10086/voxera-kit/storage"
	"github.com/EthanShen10086/voxera-kit/storage/contract"
	"github.com/EthanShen10086/voxera-kit/storage/fs"
)

func TestFSObjectStoreContract(t *testing.T) {
	contract.RunObjectStoreContract(t, func(t *testing.T) storage.ObjectStore {
		a, err := fs.New(t.TempDir(), storage.Config{})
		if err != nil {
			t.Fatalf("New: %v", err)
		}
		return a
	})
}

func TestFSMultipartContract(t *testing.T) {
	contract.RunMultipartContract(t, func(t *testing.T) (storage.MultipartUploader, storage.ObjectStore) {
		a, err := fs.New(t.TempDir(), storage.Config{})
		if err != nil {
			t.Fatalf("New: %v", err)
		}
		return a, a
	})
}

func TestFSUploadLarge(t *testing.T) {
	a, err := fs.New(t.TempDir(), storage.Config{MultipartThreshold: 1024})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	ctx := context.Background()
	data := bytes.Repeat([]byte("z"), 2048)
	if err := a.UploadLarge(ctx, "large.bin", bytes.NewReader(data), int64(len(data)), nil); err != nil {
		t.Fatalf("UploadLarge: %v", err)
	}
	rc, err := a.Download(ctx, "large.bin")
	if err != nil {
		t.Fatalf("Download: %v", err)
	}
	got, _ := io.ReadAll(rc)
	_ = rc.Close()
	if len(got) != len(data) {
		t.Fatalf("size = %d want %d", len(got), len(data))
	}
}

func TestFSInvalidKey(t *testing.T) {
	a, err := fs.New(t.TempDir(), storage.Config{})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	ctx := context.Background()
	if err := a.Upload(ctx, "../escape", bytes.NewReader([]byte("x")), nil); err == nil {
		t.Fatal("expected invalid key error")
	}
	_, err = a.Exists(ctx, "../escape")
	if err == nil {
		t.Fatal("expected invalid key error on Exists")
	}
}

func TestFSGetURL(t *testing.T) {
	a, err := fs.New(t.TempDir(), storage.Config{})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	ctx := context.Background()
	_ = a.Upload(ctx, "url.txt", bytes.NewReader([]byte("u")), nil)
	url, err := a.GetURL(ctx, "url.txt", time.Minute)
	if err != nil || url == "" {
		t.Fatalf("GetURL: %q err=%v", url, err)
	}
}
