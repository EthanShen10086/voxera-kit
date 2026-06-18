// Package minio provides a MinIO implementation of the storage object store interfaces.
package minio

import (
	"context"
	"errors"
	"io"
	"net/url"
	"strings"
	"time"

	"github.com/EthanShen10086/voxera-kit/storage"
	"github.com/EthanShen10086/voxera-kit/storage/internal/opts"
	"github.com/EthanShen10086/voxera-kit/storage/internal/uploadlarge"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/minio/minio-go/v7/pkg/lifecycle"
)

// Adapter implements storage interfaces using MinIO.
type Adapter struct {
	client *minio.Client
	core   *minio.Core
	cfg    storage.Config
}

// New creates a MinIO adapter connected to the configured endpoint.
func New(cfg storage.Config) (*Adapter, error) {
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, cfg.SessionToken),
		Secure: cfg.UseSSL,
		Region: cfg.Region,
	})
	if err != nil {
		return nil, err
	}
	core, err := minio.NewCore(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, cfg.SessionToken),
		Secure: cfg.UseSSL,
		Region: cfg.Region,
	})
	if err != nil {
		return nil, err
	}
	return &Adapter{client: client, core: core, cfg: cfg}, nil
}

func putOptions(uploadOpts *storage.UploadOptions) minio.PutObjectOptions {
	merged := opts.MergeUploadOptions(uploadOpts)
	options := minio.PutObjectOptions{
		ContentType:  merged.ContentType,
		UserMetadata: merged.Metadata,
	}
	return options
}

func mapError(err error) error {
	if err == nil {
		return nil
	}
	resp := minio.ToErrorResponse(err)
	if resp.Code == "NoSuchKey" || resp.Code == "NotFound" {
		return storage.ErrNotFound
	}
	return err
}

// Upload stores an object in the MinIO bucket.
func (a *Adapter) Upload(ctx context.Context, key string, reader io.Reader, uploadOpts *storage.UploadOptions) error {
	key = opts.NormalizeKey(key)
	_, err := a.client.PutObject(ctx, a.cfg.Bucket, key, reader, -1, putOptions(uploadOpts))
	return mapError(err)
}

// Download retrieves an object from the MinIO bucket.
func (a *Adapter) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	key = opts.NormalizeKey(key)
	obj, err := a.client.GetObject(ctx, a.cfg.Bucket, key, minio.GetObjectOptions{})
	if err != nil {
		return nil, mapError(err)
	}
	if _, err := obj.Stat(); err != nil {
		_ = obj.Close()
		return nil, mapError(err)
	}
	return obj, nil
}

// Delete removes an object from the MinIO bucket.
func (a *Adapter) Delete(ctx context.Context, key string) error {
	key = opts.NormalizeKey(key)
	err := a.client.RemoveObject(ctx, a.cfg.Bucket, key, minio.RemoveObjectOptions{})
	return mapError(err)
}

// GetURL generates a pre-signed URL for temporary access to a MinIO object.
func (a *Adapter) GetURL(ctx context.Context, key string, expiry time.Duration) (string, error) {
	key = opts.NormalizeKey(key)
	u, err := a.client.PresignedGetObject(ctx, a.cfg.Bucket, key, expiry, url.Values{})
	if err != nil {
		return "", mapError(err)
	}
	return u.String(), nil
}

// List returns metadata for all objects matching the given prefix in MinIO.
func (a *Adapter) List(ctx context.Context, prefix string) ([]*storage.ObjectMeta, error) {
	prefix = opts.NormalizeKey(prefix)
	ch := a.client.ListObjects(ctx, a.cfg.Bucket, minio.ListObjectsOptions{Prefix: prefix, Recursive: true})
	var out []*storage.ObjectMeta
	for obj := range ch {
		if obj.Err != nil {
			return nil, mapError(obj.Err)
		}
		out = append(out, &storage.ObjectMeta{
			Key:          obj.Key,
			Size:         obj.Size,
			ContentType:  obj.ContentType,
			ETag:         strings.Trim(obj.ETag, "\""),
			LastModified: obj.LastModified,
		})
	}
	return out, nil
}

// Exists checks whether an object exists in the MinIO bucket.
func (a *Adapter) Exists(ctx context.Context, key string) (bool, error) {
	key = opts.NormalizeKey(key)
	_, err := a.client.StatObject(ctx, a.cfg.Bucket, key, minio.StatObjectOptions{})
	if err == nil {
		return true, nil
	}
	if mapError(err) == storage.ErrNotFound {
		return false, nil
	}
	return false, err
}

// Close releases all resources held by the MinIO client.
func (a *Adapter) Close() error {
	return nil
}

// UploadLarge uploads a large object using multipart when above threshold.
func (a *Adapter) UploadLarge(ctx context.Context, key string, reader io.ReaderAt, size int64, uploadOpts *storage.UploadOptions) error {
	return uploadlarge.Upload(ctx, a, a, a.cfg, key, reader, size, uploadOpts)
}

// InitiateMultipartUpload starts a multipart upload.
func (a *Adapter) InitiateMultipartUpload(ctx context.Context, key string, uploadOpts *storage.UploadOptions) (string, error) {
	key = opts.NormalizeKey(key)
	id, err := a.core.NewMultipartUpload(ctx, a.cfg.Bucket, key, putOptions(uploadOpts))
	return id, mapError(err)
}

// UploadPart uploads one multipart part.
func (a *Adapter) UploadPart(ctx context.Context, key, uploadID string, partNumber int, reader io.Reader, size int64) (string, error) {
	key = opts.NormalizeKey(key)
	part, err := a.core.PutObjectPart(ctx, a.cfg.Bucket, key, uploadID, partNumber, reader, size, minio.PutObjectPartOptions{})
	if err != nil {
		return "", mapError(err)
	}
	return part.ETag, nil
}

// CompleteMultipartUpload completes a multipart upload.
func (a *Adapter) CompleteMultipartUpload(ctx context.Context, key, uploadID string, parts []storage.CompletedPart) error {
	key = opts.NormalizeKey(key)
	completeParts := make([]minio.CompletePart, len(parts))
	for i, p := range parts {
		completeParts[i] = minio.CompletePart{PartNumber: p.PartNumber, ETag: p.ETag}
	}
	_, err := a.core.CompleteMultipartUpload(ctx, a.cfg.Bucket, key, uploadID, completeParts, putOptions(nil))
	return mapError(err)
}

// AbortMultipartUpload aborts a multipart upload.
func (a *Adapter) AbortMultipartUpload(ctx context.Context, key, uploadID string) error {
	key = opts.NormalizeKey(key)
	err := a.core.AbortMultipartUpload(ctx, a.cfg.Bucket, key, uploadID)
	return mapError(err)
}

// ListVersions returns object versions for a key.
func (a *Adapter) ListVersions(ctx context.Context, key string) ([]*storage.ObjectVersion, error) {
	key = opts.NormalizeKey(key)
	enabled, err := a.GetVersioning(ctx)
	if err != nil {
		return nil, err
	}
	if !enabled {
		return nil, storage.ErrVersioningDisabled
	}

	ch := a.client.ListObjects(ctx, a.cfg.Bucket, minio.ListObjectsOptions{
		Prefix:       key,
		Recursive:    true,
		WithVersions: true,
	})
	var out []*storage.ObjectVersion
	for obj := range ch {
		if obj.Err != nil {
			return nil, mapError(obj.Err)
		}
		if obj.Key != key {
			continue
		}
		out = append(out, &storage.ObjectVersion{
			VersionID:      obj.VersionID,
			Key:            obj.Key,
			Size:           obj.Size,
			IsLatest:       obj.IsLatest,
			IsDeleteMarker: obj.IsDeleteMarker,
			LastModified:   obj.LastModified,
		})
	}
	if len(out) == 0 {
		return nil, storage.ErrNotFound
	}
	return out, nil
}

// DownloadVersion retrieves a specific object version.
func (a *Adapter) DownloadVersion(ctx context.Context, key, versionID string) (io.ReadCloser, error) {
	key = opts.NormalizeKey(key)
	options := minio.GetObjectOptions{VersionID: versionID}
	obj, err := a.client.GetObject(ctx, a.cfg.Bucket, key, options)
	if err != nil {
		return nil, mapError(err)
	}
	if _, err := obj.Stat(); err != nil {
		_ = obj.Close()
		return nil, mapError(err)
	}
	return obj, nil
}

// DeleteVersion deletes a specific object version.
func (a *Adapter) DeleteVersion(ctx context.Context, key, versionID string) error {
	key = opts.NormalizeKey(key)
	err := a.client.RemoveObject(ctx, a.cfg.Bucket, key, minio.RemoveObjectOptions{VersionID: versionID})
	return mapError(err)
}

// RestoreVersion makes a historical version current by copying it.
func (a *Adapter) RestoreVersion(ctx context.Context, key, versionID string) error {
	rc, err := a.DownloadVersion(ctx, key, versionID)
	if err != nil {
		return err
	}
	defer rc.Close()
	return a.Upload(ctx, key, rc, nil)
}

// EnableVersioning toggles bucket versioning.
func (a *Adapter) EnableVersioning(ctx context.Context, enabled bool) error {
	status := "Suspended"
	if enabled {
		status = "Enabled"
	}
	return a.client.SetBucketVersioning(ctx, a.cfg.Bucket, minio.BucketVersioningConfiguration{Status: status})
}

// GetVersioning returns whether bucket versioning is enabled.
func (a *Adapter) GetVersioning(ctx context.Context) (bool, error) {
	cfg, err := a.client.GetBucketVersioning(ctx, a.cfg.Bucket)
	if err != nil {
		return false, mapError(err)
	}
	return cfg.Status == "Enabled", nil
}

// PutLifecycleRules configures bucket lifecycle rules.
func (a *Adapter) PutLifecycleRules(ctx context.Context, rules []storage.LifecycleRule) error {
	lcfg := lifecycle.NewConfiguration()
	for _, r := range rules {
		rule := lifecycle.Rule{
			ID:     r.ID,
			Status: r.Status,
			Prefix: r.Prefix,
		}
		if r.ExpirationDays > 0 {
			rule.Expiration = lifecycle.Expiration{Days: lifecycle.ExpirationDays(r.ExpirationDays)}
		}
		if r.NoncurrentVersionExpirationDays > 0 {
			rule.NoncurrentVersionExpiration = lifecycle.NoncurrentVersionExpiration{
				NoncurrentDays: lifecycle.ExpirationDays(r.NoncurrentVersionExpirationDays),
			}
		}
		if r.TransitionToIADays > 0 {
			rule.Transition = lifecycle.Transition{
				Days:         lifecycle.ExpirationDays(r.TransitionToIADays),
				StorageClass: "STANDARD_IA",
			}
		}
		lcfg.Rules = append(lcfg.Rules, rule)
	}
	return a.client.SetBucketLifecycle(ctx, a.cfg.Bucket, lcfg)
}

// GetLifecycleRules returns bucket lifecycle rules.
func (a *Adapter) GetLifecycleRules(ctx context.Context) ([]storage.LifecycleRule, error) {
	lcfg, err := a.client.GetBucketLifecycle(ctx, a.cfg.Bucket)
	if err != nil {
		if resp := minio.ToErrorResponse(err); resp.Code == "NoSuchLifecycleConfiguration" {
			return nil, nil
		}
		return nil, mapError(err)
	}
	out := make([]storage.LifecycleRule, 0, len(lcfg.Rules))
	for _, r := range lcfg.Rules {
		rule := storage.LifecycleRule{
			ID:     r.ID,
			Status: r.Status,
			Prefix: r.Prefix,
		}
		if r.Expiration.Days > 0 {
			rule.ExpirationDays = int(r.Expiration.Days)
		}
		if r.NoncurrentVersionExpiration.NoncurrentDays > 0 {
			rule.NoncurrentVersionExpirationDays = int(r.NoncurrentVersionExpiration.NoncurrentDays)
		}
		if r.Transition.Days > 0 {
			rule.TransitionToIADays = int(r.Transition.Days)
		}
		out = append(out, rule)
	}
	return out, nil
}

// DeleteLifecycleRules removes bucket lifecycle configuration.
func (a *Adapter) DeleteLifecycleRules(ctx context.Context) error {
	return a.client.SetBucketLifecycle(ctx, a.cfg.Bucket, lifecycle.NewConfiguration())
}

// PutBucketNotification configures bucket notifications.
func (a *Adapter) PutBucketNotification(_ context.Context, _ storage.NotificationDestination) error {
	return errors.New("minio: bucket notification configuration not supported in this adapter")
}

// GetBucketNotification returns bucket notification configuration.
func (a *Adapter) GetBucketNotification(_ context.Context) (*storage.NotificationDestination, error) {
	return nil, nil
}

// DeleteBucketNotification removes bucket notification configuration.
func (a *Adapter) DeleteBucketNotification(_ context.Context) error {
	return nil
}
