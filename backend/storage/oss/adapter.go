// Package oss provides an Alibaba Cloud OSS implementation of the storage object store interfaces.
package oss

import (
	"context"
	"errors"
	"io"
	"strings"
	"time"

	"github.com/EthanShen10086/voxera-kit/storage"
	"github.com/EthanShen10086/voxera-kit/storage/internal/opts"
	"github.com/EthanShen10086/voxera-kit/storage/internal/uploadlarge"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

// Adapter implements storage interfaces using Alibaba Cloud OSS.
type Adapter struct {
	client *oss.Client
	bucket *oss.Bucket
	cfg    storage.Config
}

// New creates an OSS adapter connected to the configured endpoint.
func New(cfg storage.Config) (*Adapter, error) {
	client, err := oss.New(cfg.Endpoint, cfg.AccessKey, cfg.SecretKey)
	if err != nil {
		return nil, err
	}
	bucket, err := client.Bucket(cfg.Bucket)
	if err != nil {
		return nil, err
	}
	return &Adapter{client: client, bucket: bucket, cfg: cfg}, nil
}

func putOptions(uploadOpts *storage.UploadOptions) []oss.Option {
	merged := opts.MergeUploadOptions(uploadOpts)
	var options []oss.Option
	if merged.ContentType != "" {
		options = append(options, oss.ContentType(merged.ContentType))
	}
	for k, v := range merged.Metadata {
		options = append(options, oss.Meta(k, v))
	}
	return options
}

func mapError(err error) error {
	if err == nil {
		return nil
	}
	if strings.Contains(strings.ToLower(err.Error()), "nosuchkey") ||
		strings.Contains(strings.ToLower(err.Error()), "not found") ||
		strings.Contains(strings.ToLower(err.Error()), "status code: 404") {
		return storage.ErrNotFound
	}
	return err
}

// Upload stores an object in the OSS bucket.
func (a *Adapter) Upload(ctx context.Context, key string, reader io.Reader, uploadOpts *storage.UploadOptions) error {
	key = opts.NormalizeKey(key)
	options := putOptions(uploadOpts)
	options = append(options, oss.WithContext(ctx))
	return mapError(a.bucket.PutObject(key, reader, options...))
}

// Download retrieves an object from the OSS bucket.
func (a *Adapter) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	key = opts.NormalizeKey(key)
	body, err := a.bucket.GetObject(key, oss.WithContext(ctx))
	if err != nil {
		return nil, mapError(err)
	}
	return body, nil
}

// Delete removes an object from the OSS bucket.
func (a *Adapter) Delete(ctx context.Context, key string) error {
	key = opts.NormalizeKey(key)
	return mapError(a.bucket.DeleteObject(key, oss.WithContext(ctx)))
}

// GetURL generates a pre-signed URL for temporary access to an OSS object.
func (a *Adapter) GetURL(ctx context.Context, key string, expiry time.Duration) (string, error) {
	key = opts.NormalizeKey(key)
	url, err := a.bucket.SignURL(key, oss.HTTPGet, int64(expiry.Seconds()), oss.WithContext(ctx))
	if err != nil {
		return "", mapError(err)
	}
	return url, nil
}

// List returns metadata for all objects matching the given prefix in OSS.
func (a *Adapter) List(ctx context.Context, prefix string) ([]*storage.ObjectMeta, error) {
	prefix = opts.NormalizeKey(prefix)
	marker := ""
	var out []*storage.ObjectMeta
	for {
		result, err := a.bucket.ListObjects(oss.Prefix(prefix), oss.Marker(marker), oss.WithContext(ctx))
		if err != nil {
			return nil, mapError(err)
		}
		for _, obj := range result.Objects {
			out = append(out, &storage.ObjectMeta{
				Key:          obj.Key,
				Size:         obj.Size,
				ETag:         strings.Trim(obj.ETag, "\""),
				LastModified: obj.LastModified,
			})
		}
		if !result.IsTruncated {
			break
		}
		marker = result.NextMarker
	}
	return out, nil
}

// Exists checks whether an object exists in the OSS bucket.
func (a *Adapter) Exists(ctx context.Context, key string) (bool, error) {
	key = opts.NormalizeKey(key)
	ok, err := a.bucket.IsObjectExist(key, oss.WithContext(ctx))
	if err != nil {
		return false, mapError(err)
	}
	return ok, nil
}

// Close releases all resources held by the OSS client.
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
	imur, err := a.bucket.InitiateMultipartUpload(key, append(putOptions(uploadOpts), oss.WithContext(ctx))...)
	if err != nil {
		return "", mapError(err)
	}
	return imur.UploadID, nil
}

// UploadPart uploads one multipart part.
func (a *Adapter) UploadPart(ctx context.Context, key, uploadID string, partNumber int, reader io.Reader, size int64) (string, error) {
	key = opts.NormalizeKey(key)
	imur := oss.InitiateMultipartUploadResult{Bucket: a.cfg.Bucket, Key: key, UploadID: uploadID}
	part, err := a.bucket.UploadPart(imur, reader, size, partNumber, oss.WithContext(ctx))
	if err != nil {
		return "", mapError(err)
	}
	return part.ETag, nil
}

// CompleteMultipartUpload completes a multipart upload.
func (a *Adapter) CompleteMultipartUpload(ctx context.Context, key, uploadID string, parts []storage.CompletedPart) error {
	key = opts.NormalizeKey(key)
	imur := oss.InitiateMultipartUploadResult{Bucket: a.cfg.Bucket, Key: key, UploadID: uploadID}
	ossParts := make([]oss.UploadPart, len(parts))
	for i, p := range parts {
		ossParts[i] = oss.UploadPart{PartNumber: p.PartNumber, ETag: p.ETag}
	}
	_, err := a.bucket.CompleteMultipartUpload(imur, ossParts, oss.WithContext(ctx))
	return mapError(err)
}

// AbortMultipartUpload aborts a multipart upload.
func (a *Adapter) AbortMultipartUpload(ctx context.Context, key, uploadID string) error {
	key = opts.NormalizeKey(key)
	imur := oss.InitiateMultipartUploadResult{Bucket: a.cfg.Bucket, Key: key, UploadID: uploadID}
	return mapError(a.bucket.AbortMultipartUpload(imur, oss.WithContext(ctx)))
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

	result, err := a.bucket.ListObjectVersions(oss.Prefix(key), oss.WithContext(ctx))
	if err != nil {
		return nil, mapError(err)
	}
	var out []*storage.ObjectVersion
	for _, v := range result.ObjectVersions {
		if v.Key != key {
			continue
		}
		out = append(out, &storage.ObjectVersion{
			VersionID:    v.VersionId,
			Key:          v.Key,
			Size:         v.Size,
			IsLatest:     v.IsLatest,
			LastModified: v.LastModified,
		})
	}
	for _, m := range result.ObjectDeleteMarkers {
		if m.Key != key {
			continue
		}
		out = append(out, &storage.ObjectVersion{
			VersionID:      m.VersionId,
			Key:            m.Key,
			IsLatest:       m.IsLatest,
			IsDeleteMarker: true,
			LastModified:   m.LastModified,
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
	body, err := a.bucket.GetObject(key, oss.VersionId(versionID), oss.WithContext(ctx))
	if err != nil {
		return nil, mapError(err)
	}
	return body, nil
}

// DeleteVersion deletes a specific object version.
func (a *Adapter) DeleteVersion(ctx context.Context, key, versionID string) error {
	key = opts.NormalizeKey(key)
	return mapError(a.bucket.DeleteObject(key, oss.VersionId(versionID), oss.WithContext(ctx)))
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
	status := oss.VersionSuspended
	if enabled {
		status = oss.VersionEnabled
	}
	return mapError(a.client.SetBucketVersioning(a.cfg.Bucket, oss.VersioningConfig{Status: string(status)}, oss.WithContext(ctx)))
}

// GetVersioning returns whether bucket versioning is enabled.
func (a *Adapter) GetVersioning(ctx context.Context) (bool, error) {
	cfg, err := a.client.GetBucketVersioning(a.cfg.Bucket, oss.WithContext(ctx))
	if err != nil {
		return false, mapError(err)
	}
	return cfg.Status == string(oss.VersionEnabled), nil
}

// PutLifecycleRules configures bucket lifecycle rules.
func (a *Adapter) PutLifecycleRules(ctx context.Context, rules []storage.LifecycleRule) error {
	lifecycleRules := make([]oss.LifecycleRule, 0, len(rules))
	for _, r := range rules {
		rule := oss.LifecycleRule{
			ID:     r.ID,
			Prefix: r.Prefix,
			Status: r.Status,
		}
		if r.ExpirationDays > 0 {
			rule.Expiration = &oss.LifecycleExpiration{Days: r.ExpirationDays}
		}
		if r.NoncurrentVersionExpirationDays > 0 {
			rule.NonVersionExpiration = &oss.LifecycleVersionExpiration{NoncurrentDays: r.NoncurrentVersionExpirationDays}
		}
		if r.TransitionToIADays > 0 {
			rule.Transitions = []oss.LifecycleTransition{{Days: r.TransitionToIADays, StorageClass: oss.StorageIA}}
		}
		lifecycleRules = append(lifecycleRules, rule)
	}
	return mapError(a.client.SetBucketLifecycle(a.cfg.Bucket, lifecycleRules, oss.WithContext(ctx)))
}

// GetLifecycleRules returns bucket lifecycle rules.
func (a *Adapter) GetLifecycleRules(ctx context.Context) ([]storage.LifecycleRule, error) {
	result, err := a.client.GetBucketLifecycle(a.cfg.Bucket, oss.WithContext(ctx))
	if err != nil {
		if mapError(err) == storage.ErrNotFound {
			return nil, nil
		}
		return nil, mapError(err)
	}
	out := make([]storage.LifecycleRule, 0, len(result.Rules))
	for _, r := range result.Rules {
		rule := storage.LifecycleRule{
			ID:     r.ID,
			Prefix: r.Prefix,
			Status: r.Status,
		}
		if r.Expiration != nil {
			rule.ExpirationDays = r.Expiration.Days
		}
		if r.NonVersionExpiration != nil {
			rule.NoncurrentVersionExpirationDays = r.NonVersionExpiration.NoncurrentDays
		}
		if len(r.Transitions) > 0 {
			rule.TransitionToIADays = r.Transitions[0].Days
		}
		out = append(out, rule)
	}
	return out, nil
}

// DeleteLifecycleRules removes bucket lifecycle configuration.
func (a *Adapter) DeleteLifecycleRules(ctx context.Context) error {
	return mapError(a.client.DeleteBucketLifecycle(a.cfg.Bucket, oss.WithContext(ctx)))
}

// PutBucketNotification configures bucket notifications.
func (a *Adapter) PutBucketNotification(_ context.Context, cfg storage.NotificationDestination) error {
	_ = cfg
	return errors.New("oss: bucket notification configuration not supported in this adapter")
}

// GetBucketNotification returns bucket notification configuration.
func (a *Adapter) GetBucketNotification(_ context.Context) (*storage.NotificationDestination, error) {
	return nil, nil
}

// DeleteBucketNotification removes bucket notification configuration.
func (a *Adapter) DeleteBucketNotification(_ context.Context) error {
	return nil
}
