package uploadlarge_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"

	"github.com/EthanShen10086/voxera-kit/storage"
	"github.com/EthanShen10086/voxera-kit/storage/internal/uploadlarge"
	"github.com/EthanShen10086/voxera-kit/storage/memory"
)

func TestUploadSmallUsesSinglePut(t *testing.T) {
	ctx := context.Background()
	mem := memory.New(storage.Config{})
	data := []byte("small")
	cfg := storage.Config{MultipartThreshold: 1024}
	err := uploadlarge.Upload(ctx, mem, mem, cfg, "k", bytes.NewReader(data), int64(len(data)), nil)
	if err != nil {
		t.Fatal(err)
	}
	rc, err := mem.Download(ctx, "k")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rc.Close() }()
	got, _ := io.ReadAll(rc)
	if !bytes.Equal(got, data) {
		t.Fatalf("data = %q", got)
	}
}

func TestUploadLargeMultipart(t *testing.T) {
	ctx := context.Background()
	mem := memory.New(storage.Config{})
	data := bytes.Repeat([]byte("x"), 2048)
	cfg := storage.Config{MultipartThreshold: 512, PartSize: 512}
	err := uploadlarge.Upload(ctx, mem, mem, cfg, "large", bytes.NewReader(data), int64(len(data)), nil)
	if err != nil {
		t.Fatal(err)
	}
	rc, err := mem.Download(ctx, "large")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rc.Close() }()
	got, _ := io.ReadAll(rc)
	if !bytes.Equal(got, data) {
		t.Fatalf("len = %d", len(got))
	}
}

type failUploader struct {
	*memory.Adapter
}

func (f *failUploader) UploadPart(_ context.Context, _, _ string, _ int, _ io.Reader, _ int64) (string, error) {
	return "", errors.New("part failed")
}

func TestUploadLargeAbortsOnPartFailure(t *testing.T) {
	ctx := context.Background()
	mem := memory.New(storage.Config{})
	uploader := &failUploader{Adapter: mem}
	data := bytes.Repeat([]byte("y"), 1024)
	cfg := storage.Config{MultipartThreshold: 256, PartSize: 256}
	err := uploadlarge.Upload(ctx, mem, uploader, cfg, "bad", bytes.NewReader(data), int64(len(data)), nil)
	if err == nil {
		t.Fatal("expected upload error")
	}
}
