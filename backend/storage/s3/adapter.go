// Package s3 provides an Amazon S3 implementation of the storage.ObjectStore interface.
// It is intended to use github.com/aws/aws-sdk-go-v2 as the underlying SDK.
package s3

import (
	"context"
	"io"
	"time"

	"github.com/EthanShen10086/voxera-kit/storage"
)

// Adapter implements the storage.ObjectStore interface using Amazon S3.
//
// Intended dependency: github.com/aws/aws-sdk-go-v2/service/s3
type Adapter struct {
	// client *s3.Client // TODO: uncomment when aws-sdk-go-v2 dependency is added
	cfg storage.Config
}

// New creates a new S3 Adapter with the provided configuration.
func New(cfg storage.Config) *Adapter {
	return &Adapter{cfg: cfg}
}

// Upload stores an object in the S3 bucket.
func (a *Adapter) Upload(ctx context.Context, key string, reader io.Reader, opts *storage.UploadOptions) error {
	// TODO: implement using aws-sdk-go-v2
	return nil
}

// Download retrieves an object from the S3 bucket.
func (a *Adapter) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	// TODO: implement using aws-sdk-go-v2
	return nil, nil
}

// Delete removes an object from the S3 bucket.
func (a *Adapter) Delete(ctx context.Context, key string) error {
	// TODO: implement using aws-sdk-go-v2
	return nil
}

// GetURL generates a pre-signed URL for temporary access to an S3 object.
func (a *Adapter) GetURL(ctx context.Context, key string, expiry time.Duration) (string, error) {
	// TODO: implement using aws-sdk-go-v2 presign
	return "", nil
}

// List returns metadata for all objects matching the given prefix in S3.
func (a *Adapter) List(ctx context.Context, prefix string) ([]*storage.ObjectMeta, error) {
	// TODO: implement using aws-sdk-go-v2
	return nil, nil
}

// Exists checks whether an object exists in the S3 bucket.
func (a *Adapter) Exists(ctx context.Context, key string) (bool, error) {
	// TODO: implement using aws-sdk-go-v2
	return false, nil
}

// Close releases all resources held by the S3 client.
func (a *Adapter) Close() error {
	// TODO: implement using aws-sdk-go-v2
	return nil
}
