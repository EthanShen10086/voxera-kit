// Package minio provides a MinIO implementation of the storage.ObjectStore interface.
// It is intended to use github.com/minio/minio-go/v7 as the underlying SDK.
package minio

import (
	"context"
	"io"
	"time"

	"github.com/EthanShen10086/voxera-kit/storage"
)

// Adapter implements the storage.ObjectStore interface using MinIO.
//
// Intended dependency: github.com/minio/minio-go/v7
type Adapter struct {
	// client *minio.Client // TODO: uncomment when minio-go dependency is added
	cfg storage.StorageConfig
}

// New creates a new MinIO Adapter with the provided configuration.
func New(cfg storage.StorageConfig) *Adapter {
	return &Adapter{cfg: cfg}
}

// Upload stores an object in the MinIO bucket.
func (a *Adapter) Upload(ctx context.Context, key string, reader io.Reader, opts *storage.UploadOptions) error {
	// TODO: implement using minio-go
	return nil
}

// Download retrieves an object from the MinIO bucket.
func (a *Adapter) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	// TODO: implement using minio-go
	return nil, nil
}

// Delete removes an object from the MinIO bucket.
func (a *Adapter) Delete(ctx context.Context, key string) error {
	// TODO: implement using minio-go
	return nil
}

// GetURL generates a pre-signed URL for temporary access to a MinIO object.
func (a *Adapter) GetURL(ctx context.Context, key string, expiry time.Duration) (string, error) {
	// TODO: implement using minio-go
	return "", nil
}

// List returns metadata for all objects matching the given prefix in MinIO.
func (a *Adapter) List(ctx context.Context, prefix string) ([]*storage.ObjectMeta, error) {
	// TODO: implement using minio-go
	return nil, nil
}

// Exists checks whether an object exists in the MinIO bucket.
func (a *Adapter) Exists(ctx context.Context, key string) (bool, error) {
	// TODO: implement using minio-go
	return false, nil
}

// Close releases all resources held by the MinIO client.
func (a *Adapter) Close() error {
	// TODO: implement using minio-go
	return nil
}
