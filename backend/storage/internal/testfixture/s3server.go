// Package testfixture provides in-process S3-compatible servers for adapter tests.
package testfixture

import (
	"net/http/httptest"
	"testing"

	"github.com/EthanShen10086/voxera-kit/storage"
	"github.com/johannesboyne/gofakes3"
	"github.com/johannesboyne/gofakes3/backend/s3mem"
)

// S3Server runs an httptest-backed gofakes3 instance.
type S3Server struct {
	Server *httptest.Server
	Bucket string
}

// StartS3 launches gofakes3 with an in-memory backend and creates bucket.
func StartS3(t *testing.T, bucket string) *S3Server {
	t.Helper()
	if bucket == "" {
		bucket = "voxera-test"
	}

	backend := s3mem.New()
	faker := gofakes3.New(backend, gofakes3.WithAutoBucket(true))
	ts := httptest.NewServer(faker.Server())
	t.Cleanup(ts.Close)

	if err := backend.CreateBucket(bucket); err != nil {
		t.Fatalf("create bucket: %v", err)
	}

	return &S3Server{Server: ts, Bucket: bucket}
}

// StorageConfig returns config for the voxera-kit storage adapters.
func (s *S3Server) StorageConfig() storage.Config {
	if s == nil || s.Server == nil {
		return storage.Config{}
	}
	return storage.Config{
		Endpoint:         s.Server.URL,
		AccessKey:        "test-access",
		SecretKey:        "test-secret",
		Bucket:           s.Bucket,
		Region:           "us-east-1",
		PathStyle:        true,
		DisableSSLVerify: true,
	}
}

// MinIOEndpoint returns host:port endpoint for minio-go (no scheme).
func (s *S3Server) MinIOEndpoint() storage.Config {
	cfg := s.StorageConfig()
	cfg.Endpoint = s.Server.Listener.Addr().String()
	cfg.UseSSL = false
	return cfg
}
