// Package oss provides an Alibaba Cloud OSS implementation of the storage.ObjectStore interface.
// It is intended to use github.com/aliyun/aliyun-oss-go-sdk/oss as the underlying SDK.
package oss

import (
	"context"
	"io"
	"time"

	"github.com/EthanShen10086/voxera-kit/storage"
)

// Adapter implements the storage.ObjectStore interface using Alibaba Cloud OSS.
//
// Intended dependency: github.com/aliyun/aliyun-oss-go-sdk/oss
type Adapter struct {
	// client *oss.Client // TODO: uncomment when aliyun-oss-go-sdk dependency is added
	// bucket *oss.Bucket
	cfg storage.Config
}

// New creates a new Alibaba Cloud OSS Adapter with the provided configuration.
func New(cfg storage.Config) *Adapter {
	return &Adapter{cfg: cfg}
}

// Upload stores an object in the OSS bucket.
func (a *Adapter) Upload(ctx context.Context, key string, reader io.Reader, opts *storage.UploadOptions) error {
	// TODO: implement using aliyun-oss-go-sdk
	return nil
}

// Download retrieves an object from the OSS bucket.
func (a *Adapter) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	// TODO: implement using aliyun-oss-go-sdk
	return nil, nil
}

// Delete removes an object from the OSS bucket.
func (a *Adapter) Delete(ctx context.Context, key string) error {
	// TODO: implement using aliyun-oss-go-sdk
	return nil
}

// GetURL generates a pre-signed URL for temporary access to an OSS object.
func (a *Adapter) GetURL(ctx context.Context, key string, expiry time.Duration) (string, error) {
	// TODO: implement using aliyun-oss-go-sdk
	return "", nil
}

// List returns metadata for all objects matching the given prefix in OSS.
func (a *Adapter) List(ctx context.Context, prefix string) ([]*storage.ObjectMeta, error) {
	// TODO: implement using aliyun-oss-go-sdk
	return nil, nil
}

// Exists checks whether an object exists in the OSS bucket.
func (a *Adapter) Exists(ctx context.Context, key string) (bool, error) {
	// TODO: implement using aliyun-oss-go-sdk
	return false, nil
}

// Close releases all resources held by the OSS client.
func (a *Adapter) Close() error {
	// TODO: implement using aliyun-oss-go-sdk
	return nil
}
