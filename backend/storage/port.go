// Package storage defines the port interface for object storage operations.
// It abstracts file/object upload, download, and management across different
// storage backends (S3, MinIO, Alibaba Cloud OSS).
package storage

import (
	"context"
	"io"
	"time"
)

// ObjectMeta contains metadata about a stored object.
type ObjectMeta struct {
	// Key is the unique path/identifier of the object in the store.
	Key string
	// Size is the object size in bytes.
	Size int64
	// ContentType is the MIME type of the object.
	ContentType string
	// ETag is the entity tag for cache validation.
	ETag string
	// LastModified is the timestamp of the last modification.
	LastModified time.Time
}

// UploadOptions specifies optional parameters for object uploads.
type UploadOptions struct {
	// ContentType overrides the auto-detected MIME type.
	ContentType string
	// Metadata contains user-defined key-value pairs stored with the object.
	Metadata map[string]string
	// ACL sets the access control policy (e.g., "private", "public-read").
	ACL string
}

// ObjectStore is the interface for object storage operations.
// Implementations must be safe for concurrent use.
type ObjectStore interface {
	// Upload stores an object read from reader under the given key.
	Upload(ctx context.Context, key string, reader io.Reader, opts *UploadOptions) error
	// Download retrieves the object content as a readable stream.
	// The caller is responsible for closing the returned ReadCloser.
	Download(ctx context.Context, key string) (io.ReadCloser, error)
	// Delete removes the object identified by key.
	Delete(ctx context.Context, key string) error
	// GetURL generates a pre-signed URL for temporary access to the object.
	GetURL(ctx context.Context, key string, expiry time.Duration) (string, error)
	// List returns metadata for all objects matching the given prefix.
	List(ctx context.Context, prefix string) ([]*ObjectMeta, error)
	// Exists checks whether an object with the given key exists.
	Exists(ctx context.Context, key string) (bool, error)
	// Close releases all resources held by the storage client.
	Close() error
}

// Config holds the connection parameters for an object storage backend.
type Config struct {
	// Endpoint is the storage service endpoint URL.
	Endpoint string
	// AccessKey is the access key ID for authentication.
	AccessKey string
	// SecretKey is the secret access key for authentication.
	SecretKey string
	// Bucket is the target bucket/container name.
	Bucket string
	// Region is the storage service region.
	Region string
	// UseSSL enables HTTPS for the connection when true.
	UseSSL bool
}
