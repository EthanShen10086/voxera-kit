// Package storage defines the port interface for object storage operations.
// It abstracts file/object upload, download, and management across different
// storage backends (S3, MinIO, Alibaba Cloud OSS, Tencent COS).
package storage

import (
	"context"
	"io"
	"time"
)

// ObjectMeta contains metadata about a stored object.
type ObjectMeta struct {
	Key          string
	Size         int64
	ContentType  string
	ETag         string
	LastModified time.Time
}

// UploadOptions specifies optional parameters for object uploads.
type UploadOptions struct {
	ContentType string
	Metadata    map[string]string
	ACL         string
}

// CompletedPart identifies a finished multipart upload part.
type CompletedPart struct {
	PartNumber int
	ETag       string
}

// ObjectVersion describes a single object version or delete marker.
type ObjectVersion struct {
	VersionID      string
	Key            string
	Size           int64
	IsLatest       bool
	IsDeleteMarker bool
	LastModified   time.Time
}

// LifecycleRule describes bucket lifecycle configuration (minimal cross-cloud subset).
type LifecycleRule struct {
	ID                              string
	Prefix                          string
	Status                          string // Enabled / Disabled
	ExpirationDays                  int
	NoncurrentVersionExpirationDays int
	TransitionToIADays              int
}

// NotificationEvent identifies object storage events.
type NotificationEvent string

const (
	EventObjectCreated NotificationEvent = "ObjectCreated"
	EventObjectRemoved NotificationEvent = "ObjectRemoved"
)

// NotificationDestination routes bucket events to a downstream target.
type NotificationDestination struct {
	Type   string // mq, webhook, sqs, sns
	Target string
	Events []NotificationEvent
}

// ObjectStore is the interface for object storage operations.
// Implementations must be safe for concurrent use.
type ObjectStore interface {
	Upload(ctx context.Context, key string, reader io.Reader, opts *UploadOptions) error
	Download(ctx context.Context, key string) (io.ReadCloser, error)
	Delete(ctx context.Context, key string) error
	GetURL(ctx context.Context, key string, expiry time.Duration) (string, error)
	List(ctx context.Context, prefix string) ([]*ObjectMeta, error)
	Exists(ctx context.Context, key string) (bool, error)
	Close() error
}

// MultipartUploader provides explicit multipart upload APIs.
type MultipartUploader interface {
	InitiateMultipartUpload(ctx context.Context, key string, opts *UploadOptions) (uploadID string, err error)
	UploadPart(ctx context.Context, key, uploadID string, partNumber int, reader io.Reader, size int64) (etag string, err error)
	CompleteMultipartUpload(ctx context.Context, key, uploadID string, parts []CompletedPart) error
	AbortMultipartUpload(ctx context.Context, key, uploadID string) error
}

// LargeObjectStore combines ObjectStore with convenience large-file upload.
type LargeObjectStore interface {
	ObjectStore
	UploadLarge(ctx context.Context, key string, reader io.ReaderAt, size int64, opts *UploadOptions) error
}

// VersionedObjectStore supports object versioning operations.
type VersionedObjectStore interface {
	ListVersions(ctx context.Context, key string) ([]*ObjectVersion, error)
	DownloadVersion(ctx context.Context, key, versionID string) (io.ReadCloser, error)
	DeleteVersion(ctx context.Context, key, versionID string) error
	RestoreVersion(ctx context.Context, key, versionID string) error
}

// LifecycleManager manages bucket lifecycle rules.
type LifecycleManager interface {
	PutLifecycleRules(ctx context.Context, rules []LifecycleRule) error
	GetLifecycleRules(ctx context.Context) ([]LifecycleRule, error)
	DeleteLifecycleRules(ctx context.Context) error
}

// NotificationManager configures bucket event notifications.
type NotificationManager interface {
	PutBucketNotification(ctx context.Context, cfg NotificationDestination) error
	GetBucketNotification(ctx context.Context) (*NotificationDestination, error)
	DeleteBucketNotification(ctx context.Context) error
}

// StorageAdmin provides bucket-level administration.
type StorageAdmin interface {
	EnableVersioning(ctx context.Context, enabled bool) error
	GetVersioning(ctx context.Context) (bool, error)
	LifecycleManager
	NotificationManager
}

// Config holds the connection parameters for an object storage backend.
type Config struct {
	Endpoint             string
	AccessKey            string
	SecretKey            string
	Bucket               string
	Region               string
	UseSSL               bool
	PathStyle            bool
	SessionToken         string
	DisableSSLVerify     bool
	PartSize             int64
	MultipartThreshold   int64
}

// DefaultPartSize is the default multipart part size (8 MiB).
const DefaultPartSize = 8 << 20

// DefaultMultipartThreshold is the size above which UploadLarge uses multipart (100 MiB).
const DefaultMultipartThreshold = 100 << 20
