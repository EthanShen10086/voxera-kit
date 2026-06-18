package containers

import (
	"context"
	"fmt"
	"strings"

	"github.com/EthanShen10086/voxera-kit/storage"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	tcminio "github.com/testcontainers/testcontainers-go/modules/minio"
)

// MinIO holds a running MinIO testcontainer.
type MinIO struct {
	Config    storage.Config
	terminate func(context.Context) error
}

// StartMinIO launches minio/minio and creates the target bucket.
func StartMinIO(ctx context.Context, bucket string) (*MinIO, error) {
	if bucket == "" {
		bucket = "voxera-test"
	}
	c, err := tcminio.Run(ctx, "minio/minio:RELEASE.2024-01-16T16-07-38Z",
		tcminio.WithUsername("minioadmin"),
		tcminio.WithPassword("minioadmin"),
	)
	if err != nil {
		return nil, fmt.Errorf("containers: start minio: %w", err)
	}
	terminate := func(ctx context.Context) error { return c.Terminate(ctx) }

	endpoint, err := c.ConnectionString(ctx)
	if err != nil {
		_ = terminate(ctx)
		return nil, fmt.Errorf("containers: minio endpoint: %w", err)
	}
	endpoint = strings.TrimPrefix(endpoint, "http://")
	endpoint = strings.TrimPrefix(endpoint, "https://")

	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(c.Username, c.Password, ""),
		Secure: false,
	})
	if err != nil {
		_ = terminate(ctx)
		return nil, fmt.Errorf("containers: minio client: %w", err)
	}
	if err := client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{Region: "us-east-1"}); err != nil {
		exists, errBucketExists := client.BucketExists(ctx, bucket)
		if errBucketExists != nil || !exists {
			_ = terminate(ctx)
			return nil, fmt.Errorf("containers: minio make bucket: %w", err)
		}
	}

	return &MinIO{
		Config: storage.Config{
			Endpoint:  endpoint,
			AccessKey: c.Username,
			SecretKey: c.Password,
			Bucket:    bucket,
			Region:    "us-east-1",
			UseSSL:    false,
			PathStyle: true,
		},
		terminate: terminate,
	}, nil
}

// Terminate stops the container.
func (m *MinIO) Terminate(ctx context.Context) error {
	if m == nil || m.terminate == nil {
		return nil
	}
	return m.terminate(ctx)
}

// S3CompatConfig returns storage.Config suitable for the S3 adapter against this MinIO instance.
func (m *MinIO) S3CompatConfig() storage.Config {
	if m == nil {
		return storage.Config{}
	}
	cfg := m.Config
	if cfg.Endpoint != "" && !strings.HasPrefix(cfg.Endpoint, "http") {
		scheme := "http"
		if cfg.UseSSL {
			scheme = "https"
		}
		cfg.Endpoint = scheme + "://" + cfg.Endpoint
	}
	cfg.DisableSSLVerify = true
	return cfg
}
