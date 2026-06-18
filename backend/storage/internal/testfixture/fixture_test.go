package testfixture_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/EthanShen10086/voxera-kit/storage"
	cosstore "github.com/EthanShen10086/voxera-kit/storage/cos"
	"github.com/EthanShen10086/voxera-kit/storage/internal/testfixture"
	ossstore "github.com/EthanShen10086/voxera-kit/storage/oss"
	s3store "github.com/EthanShen10086/voxera-kit/storage/s3"
)

func TestS3MockSmoke(t *testing.T) {
	srv := testfixture.StartS3(t, "smoke")
	a, err := s3store.New(srv.StorageConfig())
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err := a.Upload(ctx, "k", bytes.NewReader([]byte("v")), nil); err != nil {
		t.Fatal(err)
	}
}

func TestCOSMockSmoke(t *testing.T) {
	cfg := testfixture.StartCOSMock(t, "smoke")
	a, err := cosstore.New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if err := a.Upload(context.Background(), "k", bytes.NewReader([]byte("v")), nil); err != nil {
		t.Fatal(err)
	}
}

func TestOSSMockSmoke(t *testing.T) {
	cfg := testfixture.StartOSSMock(t, "smoke")
	a, err := ossstore.New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if err := a.Upload(context.Background(), "k", bytes.NewReader([]byte("v")), nil); err != nil {
		t.Fatal(err)
	}
}

func TestS3StorageConfig(t *testing.T) {
	var cfg storage.Config
	if cfg = testfixture.StartS3(t, "b").StorageConfig(); cfg.Bucket != "b" {
		t.Fatalf("bucket = %q", cfg.Bucket)
	}
}
