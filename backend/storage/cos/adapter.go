// Package cos provides a Tencent Cloud COS implementation of the storage object store interfaces.
package cos

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/EthanShen10086/voxera-kit/storage"
	"github.com/EthanShen10086/voxera-kit/storage/internal/opts"
	"github.com/EthanShen10086/voxera-kit/storage/internal/uploadlarge"
	cossdk "github.com/tencentyun/cos-go-sdk-v5"
)

// Adapter implements storage interfaces using Tencent Cloud COS.
type Adapter struct {
	client *cossdk.Client
	cfg    storage.Config
}

// New creates a COS adapter connected to the configured endpoint.
func New(cfg storage.Config) (*Adapter, error) {
	scheme := "https"
	if !cfg.UseSSL {
		scheme = "http"
	}
	bucketURL, err := url.Parse(fmt.Sprintf("%s://%s", scheme, cfg.Endpoint))
	if err != nil {
		return nil, err
	}
	if cfg.Bucket != "" && !strings.Contains(cfg.Endpoint, cfg.Bucket) {
		bucketURL, err = url.Parse(fmt.Sprintf("%s://%s.%s", scheme, cfg.Bucket, cfg.Endpoint))
		if err != nil {
			return nil, err
		}
	}

	baseURL := &cossdk.BaseURL{BucketURL: bucketURL}
	client := cossdk.NewClient(baseURL, &http.Client{
		Transport: &cossdk.AuthorizationTransport{
			SecretID:  cfg.AccessKey,
			SecretKey: cfg.SecretKey,
		},
	})
	return &Adapter{client: client, cfg: cfg}, nil
}

func headerOptions(uploadOpts *storage.UploadOptions) *cossdk.ObjectPutOptions {
	merged := opts.MergeUploadOptions(uploadOpts)
	opt := &cossdk.ObjectPutOptions{ObjectPutHeaderOptions: &cossdk.ObjectPutHeaderOptions{}}
	if merged.ContentType != "" {
		opt.ContentType = merged.ContentType
	}
	if len(merged.Metadata) > 0 {
		opt.XCosMetaXXX = &http.Header{}
		for k, v := range merged.Metadata {
			opt.XCosMetaXXX.Add("x-cos-meta-"+k, v)
		}
	}
	return opt
}

func initiateOptions(uploadOpts *storage.UploadOptions) *cossdk.InitiateMultipartUploadOptions {
	put := headerOptions(uploadOpts)
	return &cossdk.InitiateMultipartUploadOptions{
		ObjectPutHeaderOptions: put.ObjectPutHeaderOptions,
	}
}

func mapError(err error) error {
	if err == nil {
		return nil
	}
	if cossdk.IsNotFoundError(err) {
		return storage.ErrNotFound
	}
	if strings.Contains(strings.ToLower(err.Error()), "nosuchkey") ||
		strings.Contains(strings.ToLower(err.Error()), "not found") ||
		strings.Contains(strings.ToLower(err.Error()), "404") {
		return storage.ErrNotFound
	}
	return err
}

// Upload stores an object in the COS bucket.
func (a *Adapter) Upload(ctx context.Context, key string, reader io.Reader, uploadOpts *storage.UploadOptions) error {
	key = opts.NormalizeKey(key)
	_, err := a.client.Object.Put(ctx, key, reader, headerOptions(uploadOpts))
	return mapError(err)
}

// Download retrieves an object from the COS bucket.
func (a *Adapter) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	key = opts.NormalizeKey(key)
	resp, err := a.client.Object.Get(ctx, key, nil)
	if err != nil {
		return nil, mapError(err)
	}
	return resp.Body, nil
}

// Delete removes an object from the COS bucket.
func (a *Adapter) Delete(ctx context.Context, key string) error {
	key = opts.NormalizeKey(key)
	_, err := a.client.Object.Delete(ctx, key)
	return mapError(err)
}

// GetURL generates a pre-signed URL for temporary access to a COS object.
func (a *Adapter) GetURL(ctx context.Context, key string, expiry time.Duration) (string, error) {
	key = opts.NormalizeKey(key)
	u, err := a.client.Object.GetPresignedURL(ctx, http.MethodGet, key, a.cfg.AccessKey, a.cfg.SecretKey, expiry, nil)
	if err != nil {
		return "", mapError(err)
	}
	return u.String(), nil
}

// List returns metadata for all objects matching the given prefix in COS.
func (a *Adapter) List(ctx context.Context, prefix string) ([]*storage.ObjectMeta, error) {
	prefix = opts.NormalizeKey(prefix)
	opt := &cossdk.BucketGetOptions{Prefix: prefix}
	var out []*storage.ObjectMeta
	for {
		result, _, err := a.client.Bucket.Get(ctx, opt)
		if err != nil {
			return nil, mapError(err)
		}
		for _, obj := range result.Contents {
			lastModified, _ := time.Parse(time.RFC3339, obj.LastModified)
			out = append(out, &storage.ObjectMeta{
				Key:          obj.Key,
				Size:         obj.Size,
				ETag:         strings.Trim(obj.ETag, "\""),
				LastModified: lastModified,
			})
		}
		if !result.IsTruncated {
			break
		}
		opt.Marker = result.NextMarker
	}
	return out, nil
}

// Exists checks whether an object exists in the COS bucket.
func (a *Adapter) Exists(ctx context.Context, key string) (bool, error) {
	key = opts.NormalizeKey(key)
	ok, err := a.client.Object.IsExist(ctx, key)
	if err != nil {
		return false, mapError(err)
	}
	return ok, nil
}

// Close releases all resources held by the COS client.
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
	result, _, err := a.client.Object.InitiateMultipartUpload(ctx, key, initiateOptions(uploadOpts))
	if err != nil {
		return "", mapError(err)
	}
	return result.UploadID, nil
}

// UploadPart uploads one multipart part.
func (a *Adapter) UploadPart(ctx context.Context, key, uploadID string, partNumber int, reader io.Reader, size int64) (string, error) {
	key = opts.NormalizeKey(key)
	partOpt := &cossdk.ObjectUploadPartOptions{ContentLength: size}
	resp, err := a.client.Object.UploadPart(ctx, key, uploadID, partNumber, reader, partOpt)
	if err != nil {
		return "", mapError(err)
	}
	return resp.Header.Get("ETag"), nil
}

// CompleteMultipartUpload completes a multipart upload.
func (a *Adapter) CompleteMultipartUpload(ctx context.Context, key, uploadID string, parts []storage.CompletedPart) error {
	key = opts.NormalizeKey(key)
	cosParts := make([]cossdk.Object, len(parts))
	for i, p := range parts {
		cosParts[i] = cossdk.Object{PartNumber: p.PartNumber, ETag: p.ETag}
	}
	_, _, err := a.client.Object.CompleteMultipartUpload(ctx, key, uploadID, &cossdk.CompleteMultipartUploadOptions{
		Parts: cosParts,
	})
	return mapError(err)
}

// AbortMultipartUpload aborts a multipart upload.
func (a *Adapter) AbortMultipartUpload(ctx context.Context, key, uploadID string) error {
	key = opts.NormalizeKey(key)
	_, err := a.client.Object.AbortMultipartUpload(ctx, key, uploadID)
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

	opt := &cossdk.BucketGetObjectVersionsOptions{Prefix: key}
	result, _, err := a.client.Bucket.GetObjectVersions(ctx, opt)
	if err != nil {
		return nil, mapError(err)
	}
	var out []*storage.ObjectVersion
	for _, v := range result.Version {
		if v.Key != key {
			continue
		}
		lastModified, _ := time.Parse(time.RFC3339, v.LastModified)
		out = append(out, &storage.ObjectVersion{
			VersionID:    v.VersionId,
			Key:          v.Key,
			Size:         v.Size,
			IsLatest:     v.IsLatest,
			LastModified: lastModified,
		})
	}
	for _, m := range result.DeleteMarker {
		if m.Key != key {
			continue
		}
		lastModified, _ := time.Parse(time.RFC3339, m.LastModified)
		out = append(out, &storage.ObjectVersion{
			VersionID:      m.VersionId,
			Key:            m.Key,
			IsLatest:       m.IsLatest,
			IsDeleteMarker: true,
			LastModified:   lastModified,
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
	resp, err := a.client.Object.Get(ctx, key, nil, versionID)
	if err != nil {
		return nil, mapError(err)
	}
	return resp.Body, nil
}

// DeleteVersion deletes a specific object version.
func (a *Adapter) DeleteVersion(ctx context.Context, key, versionID string) error {
	key = opts.NormalizeKey(key)
	_, err := a.client.Object.Delete(ctx, key, &cossdk.ObjectDeleteOptions{VersionId: versionID})
	return mapError(err)
}

// RestoreVersion makes a historical version current by copying it.
func (a *Adapter) RestoreVersion(ctx context.Context, key, versionID string) error {
	rc, err := a.DownloadVersion(ctx, key, versionID)
	if err != nil {
		return err
	}
	defer func() { _ = rc.Close() }()
	return a.Upload(ctx, key, rc, nil)
}

// EnableVersioning toggles bucket versioning.
func (a *Adapter) EnableVersioning(ctx context.Context, enabled bool) error {
	status := "Suspended"
	if enabled {
		status = "Enabled"
	}
	_, err := a.client.Bucket.PutVersioning(ctx, &cossdk.BucketPutVersionOptions{Status: status})
	return mapError(err)
}

// GetVersioning returns whether bucket versioning is enabled.
func (a *Adapter) GetVersioning(ctx context.Context) (bool, error) {
	result, _, err := a.client.Bucket.GetVersioning(ctx)
	if err != nil {
		return false, mapError(err)
	}
	return result.Status == "Enabled", nil
}

// PutLifecycleRules configures bucket lifecycle rules.
func (a *Adapter) PutLifecycleRules(ctx context.Context, rules []storage.LifecycleRule) error {
	cosRules := make([]cossdk.BucketLifecycleRule, 0, len(rules))
	for _, r := range rules {
		rule := cossdk.BucketLifecycleRule{
			ID:     r.ID,
			Status: r.Status,
			Filter: &cossdk.BucketLifecycleFilter{Prefix: r.Prefix},
		}
		if r.ExpirationDays > 0 {
			rule.Expiration = &cossdk.BucketLifecycleExpiration{Days: r.ExpirationDays}
		}
		if r.NoncurrentVersionExpirationDays > 0 {
			rule.NoncurrentVersionExpiration = &cossdk.BucketLifecycleNoncurrentVersion{
				NoncurrentDays: r.NoncurrentVersionExpirationDays,
			}
		}
		if r.TransitionToIADays > 0 {
			rule.Transition = []cossdk.BucketLifecycleTransition{{
				Days:         r.TransitionToIADays,
				StorageClass: "STANDARD_IA",
			}}
		}
		cosRules = append(cosRules, rule)
	}
	_, err := a.client.Bucket.PutLifecycle(ctx, &cossdk.BucketPutLifecycleOptions{Rules: cosRules})
	return mapError(err)
}

// GetLifecycleRules returns bucket lifecycle rules.
func (a *Adapter) GetLifecycleRules(ctx context.Context) ([]storage.LifecycleRule, error) {
	result, _, err := a.client.Bucket.GetLifecycle(ctx)
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
			Status: r.Status,
		}
		if r.Filter != nil {
			rule.Prefix = r.Filter.Prefix
		}
		if r.Expiration != nil {
			rule.ExpirationDays = r.Expiration.Days
		}
		if r.NoncurrentVersionExpiration != nil {
			rule.NoncurrentVersionExpirationDays = r.NoncurrentVersionExpiration.NoncurrentDays
		}
		if len(r.Transition) > 0 {
			rule.TransitionToIADays = r.Transition[0].Days
		}
		out = append(out, rule)
	}
	return out, nil
}

// DeleteLifecycleRules removes bucket lifecycle configuration.
func (a *Adapter) DeleteLifecycleRules(ctx context.Context) error {
	_, err := a.client.Bucket.DeleteLifecycle(ctx)
	return mapError(err)
}

// PutBucketNotification configures bucket notifications.
func (a *Adapter) PutBucketNotification(_ context.Context, _ storage.NotificationDestination) error {
	return errors.New("cos: bucket notification configuration not supported in this adapter")
}

// GetBucketNotification returns bucket notification configuration.
func (a *Adapter) GetBucketNotification(_ context.Context) (*storage.NotificationDestination, error) {
	return nil, nil
}

// DeleteBucketNotification removes bucket notification configuration.
func (a *Adapter) DeleteBucketNotification(_ context.Context) error {
	return nil
}
