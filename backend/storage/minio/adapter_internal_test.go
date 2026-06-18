package minio

import (
	"errors"
	"testing"

	"github.com/EthanShen10086/voxera-kit/storage"
	"github.com/minio/minio-go/v7"
)

func TestMapError(t *testing.T) {
	if mapError(nil) != nil {
		t.Fatal("nil should stay nil")
	}
	resp := minio.ErrorResponse{Code: "NoSuchKey"}
	if !errors.Is(mapError(resp), storage.ErrNotFound) {
		t.Fatal("expected not found")
	}
	if mapError(errors.New("other")) == nil {
		t.Fatal("expected original error")
	}
}

func TestPutOptions(t *testing.T) {
	opts := putOptions(&storage.UploadOptions{ContentType: "text/plain", Metadata: map[string]string{"a": "b"}})
	if opts.ContentType != "text/plain" {
		t.Fatalf("content type = %q", opts.ContentType)
	}
}

func TestNewClient(t *testing.T) {
	a, err := New(storage.Config{
		Endpoint: "localhost:9000", Bucket: "b", AccessKey: "ak", SecretKey: "sk",
	})
	if err != nil {
		t.Fatal(err)
	}
	if a.cfg.Bucket != "b" {
		t.Fatalf("bucket = %q", a.cfg.Bucket)
	}
}
