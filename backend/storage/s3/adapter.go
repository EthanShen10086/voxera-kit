// Package s3 provides an Amazon S3 implementation of the storage object store interfaces.
package s3

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/EthanShen10086/voxera-kit/storage"
	"github.com/EthanShen10086/voxera-kit/storage/internal/opts"
	"github.com/EthanShen10086/voxera-kit/storage/internal/uploadlarge"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	smithyhttp "github.com/aws/smithy-go/transport/http"
)

// Adapter implements storage interfaces using Amazon S3 compatible APIs.
type Adapter struct {
	client  *s3.Client
	presign *s3.PresignClient
	cfg     storage.Config
}

// New creates an S3 adapter with optional custom endpoint support.
func New(cfg storage.Config) (*Adapter, error) {
	loadOpts := []func(*config.LoadOptions) error{
		config.WithRegion(cfg.Region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			cfg.AccessKey,
			cfg.SecretKey,
			cfg.SessionToken,
		)),
	}
	if cfg.DisableSSLVerify {
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //nolint:gosec
		}
		loadOpts = append(loadOpts, config.WithHTTPClient(&http.Client{Transport: tr}))
	}

	awsCfg, err := config.LoadDefaultConfig(context.Background(), loadOpts...)
	if err != nil {
		return nil, err
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		if cfg.Endpoint != "" {
			o.BaseEndpoint = aws.String(cfg.Endpoint)
		}
		o.UsePathStyle = cfg.PathStyle
		if cfg.Region != "" {
			o.Region = cfg.Region
		}
	})

	return &Adapter{
		client:  client,
		presign: s3.NewPresignClient(client),
		cfg:     cfg,
	}, nil
}

func mapError(err error) error {
	if err == nil {
		return nil
	}
	var respErr *smithyhttp.ResponseError
	if errors.As(err, &respErr) && respErr.HTTPStatusCode() == http.StatusNotFound {
		return storage.ErrNotFound
	}
	var nsk *types.NoSuchKey
	if errors.As(err, &nsk) {
		return storage.ErrNotFound
	}
	var nb *types.NotFound
	if errors.As(err, &nb) {
		return storage.ErrNotFound
	}
	if strings.Contains(strings.ToLower(err.Error()), "not found") ||
		strings.Contains(strings.ToLower(err.Error()), "nosuchkey") {
		return storage.ErrNotFound
	}
	return err
}

func (a *Adapter) uploadInput(key string, uploadOpts *storage.UploadOptions) *s3.PutObjectInput {
	merged := opts.MergeUploadOptions(uploadOpts)
	input := &s3.PutObjectInput{
		Bucket: a.bucket(),
		Key:    aws.String(key),
	}
	if merged.ContentType != "" {
		input.ContentType = aws.String(merged.ContentType)
	}
	if len(merged.Metadata) > 0 {
		input.Metadata = merged.Metadata
	}
	return input
}

func (a *Adapter) bucket() *string {
	return aws.String(a.cfg.Bucket)
}

// Upload stores an object in the S3 bucket.
func (a *Adapter) Upload(ctx context.Context, key string, reader io.Reader, uploadOpts *storage.UploadOptions) error {
	key = opts.NormalizeKey(key)
	input := a.uploadInput(key, uploadOpts)
	input.Body = reader
	_, err := a.client.PutObject(ctx, input)
	return mapError(err)
}

// Download retrieves an object from the S3 bucket.
func (a *Adapter) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	key = opts.NormalizeKey(key)
	out, err := a.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: a.bucket(),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, mapError(err)
	}
	return out.Body, nil
}

// Delete removes an object from the S3 bucket.
func (a *Adapter) Delete(ctx context.Context, key string) error {
	key = opts.NormalizeKey(key)
	_, err := a.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: a.bucket(),
		Key:    aws.String(key),
	})
	return mapError(err)
}

// GetURL generates a pre-signed URL for temporary access to an S3 object.
func (a *Adapter) GetURL(ctx context.Context, key string, expiry time.Duration) (string, error) {
	key = opts.NormalizeKey(key)
	out, err := a.presign.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: a.bucket(),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(expiry))
	if err != nil {
		return "", mapError(err)
	}
	return out.URL, nil
}

// List returns metadata for all objects matching the given prefix in S3.
func (a *Adapter) List(ctx context.Context, prefix string) ([]*storage.ObjectMeta, error) {
	prefix = opts.NormalizeKey(prefix)
	paginator := s3.NewListObjectsV2Paginator(a.client, &s3.ListObjectsV2Input{
		Bucket: a.bucket(),
		Prefix: aws.String(prefix),
	})
	var out []*storage.ObjectMeta
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, mapError(err)
		}
		for _, obj := range page.Contents {
			out = append(out, &storage.ObjectMeta{
				Key:          aws.ToString(obj.Key),
				Size:         aws.ToInt64(obj.Size),
				ETag:         strings.Trim(aws.ToString(obj.ETag), "\""),
				LastModified: aws.ToTime(obj.LastModified),
			})
		}
	}
	return out, nil
}

// Exists checks whether an object exists in the S3 bucket.
func (a *Adapter) Exists(ctx context.Context, key string) (bool, error) {
	key = opts.NormalizeKey(key)
	_, err := a.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: a.bucket(),
		Key:    aws.String(key),
	})
	if err == nil {
		return true, nil
	}
	if mapError(err) == storage.ErrNotFound {
		return false, nil
	}
	return false, err
}

// Close releases all resources held by the S3 client.
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
	merged := opts.MergeUploadOptions(uploadOpts)
	input := &s3.CreateMultipartUploadInput{
		Bucket: a.bucket(),
		Key:    aws.String(key),
	}
	if merged.ContentType != "" {
		input.ContentType = aws.String(merged.ContentType)
	}
	if len(merged.Metadata) > 0 {
		input.Metadata = merged.Metadata
	}
	out, err := a.client.CreateMultipartUpload(ctx, input)
	if err != nil {
		return "", mapError(err)
	}
	return aws.ToString(out.UploadId), nil
}

// UploadPart uploads one multipart part.
func (a *Adapter) UploadPart(ctx context.Context, key, uploadID string, partNumber int, reader io.Reader, size int64) (string, error) {
	key = opts.NormalizeKey(key)
	out, err := a.client.UploadPart(ctx, &s3.UploadPartInput{
		Bucket:        a.bucket(),
		Key:           aws.String(key),
		UploadId:      aws.String(uploadID),
		PartNumber:    aws.Int32(safeInt32(partNumber)),
		Body:          reader,
		ContentLength: aws.Int64(size),
	})
	if err != nil {
		return "", mapError(err)
	}
	return aws.ToString(out.ETag), nil
}

// CompleteMultipartUpload completes a multipart upload.
func (a *Adapter) CompleteMultipartUpload(ctx context.Context, key, uploadID string, parts []storage.CompletedPart) error {
	key = opts.NormalizeKey(key)
	completed := make([]types.CompletedPart, len(parts))
	for i, p := range parts {
		completed[i] = types.CompletedPart{
			PartNumber: aws.Int32(safeInt32(p.PartNumber)),
			ETag:       aws.String(p.ETag),
		}
	}
	_, err := a.client.CompleteMultipartUpload(ctx, &s3.CompleteMultipartUploadInput{
		Bucket:   a.bucket(),
		Key:      aws.String(key),
		UploadId: aws.String(uploadID),
		MultipartUpload: &types.CompletedMultipartUpload{
			Parts: completed,
		},
	})
	return mapError(err)
}

// AbortMultipartUpload aborts a multipart upload.
func (a *Adapter) AbortMultipartUpload(ctx context.Context, key, uploadID string) error {
	key = opts.NormalizeKey(key)
	_, err := a.client.AbortMultipartUpload(ctx, &s3.AbortMultipartUploadInput{
		Bucket:   a.bucket(),
		Key:      aws.String(key),
		UploadId: aws.String(uploadID),
	})
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

	out, err := a.client.ListObjectVersions(ctx, &s3.ListObjectVersionsInput{
		Bucket: a.bucket(),
		Prefix: aws.String(key),
	})
	if err != nil {
		return nil, mapError(err)
	}

	var versions []*storage.ObjectVersion
	for _, v := range out.Versions {
		if aws.ToString(v.Key) != key {
			continue
		}
		versions = append(versions, &storage.ObjectVersion{
			VersionID:    aws.ToString(v.VersionId),
			Key:          aws.ToString(v.Key),
			Size:         aws.ToInt64(v.Size),
			IsLatest:     aws.ToBool(v.IsLatest),
			LastModified: aws.ToTime(v.LastModified),
		})
	}
	for _, m := range out.DeleteMarkers {
		if aws.ToString(m.Key) != key {
			continue
		}
		versions = append(versions, &storage.ObjectVersion{
			VersionID:      aws.ToString(m.VersionId),
			Key:            aws.ToString(m.Key),
			IsLatest:       aws.ToBool(m.IsLatest),
			IsDeleteMarker: true,
			LastModified:   aws.ToTime(m.LastModified),
		})
	}
	if len(versions) == 0 {
		return nil, storage.ErrNotFound
	}
	return versions, nil
}

// DownloadVersion retrieves a specific object version.
func (a *Adapter) DownloadVersion(ctx context.Context, key, versionID string) (io.ReadCloser, error) {
	key = opts.NormalizeKey(key)
	out, err := a.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket:    a.bucket(),
		Key:       aws.String(key),
		VersionId: aws.String(versionID),
	})
	if err != nil {
		return nil, mapError(err)
	}
	return out.Body, nil
}

// DeleteVersion deletes a specific object version.
func (a *Adapter) DeleteVersion(ctx context.Context, key, versionID string) error {
	key = opts.NormalizeKey(key)
	_, err := a.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket:    a.bucket(),
		Key:       aws.String(key),
		VersionId: aws.String(versionID),
	})
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
	status := types.BucketVersioningStatusSuspended
	if enabled {
		status = types.BucketVersioningStatusEnabled
	}
	_, err := a.client.PutBucketVersioning(ctx, &s3.PutBucketVersioningInput{
		Bucket: a.bucket(),
		VersioningConfiguration: &types.VersioningConfiguration{
			Status: status,
		},
	})
	return mapError(err)
}

// GetVersioning returns whether bucket versioning is enabled.
func (a *Adapter) GetVersioning(ctx context.Context) (bool, error) {
	out, err := a.client.GetBucketVersioning(ctx, &s3.GetBucketVersioningInput{
		Bucket: a.bucket(),
	})
	if err != nil {
		return false, mapError(err)
	}
	return out.Status == types.BucketVersioningStatusEnabled, nil
}

// PutLifecycleRules configures bucket lifecycle rules.
func (a *Adapter) PutLifecycleRules(ctx context.Context, rules []storage.LifecycleRule) error {
	s3Rules := make([]types.LifecycleRule, 0, len(rules))
	for _, r := range rules {
		rule := types.LifecycleRule{
			ID:     aws.String(r.ID),
			Status: types.ExpirationStatus(r.Status),
			Filter: &types.LifecycleRuleFilter{Prefix: aws.String(r.Prefix)},
		}
		if r.ExpirationDays > 0 {
			rule.Expiration = &types.LifecycleExpiration{Days: aws.Int32(safeInt32(r.ExpirationDays))}
		}
		if r.NoncurrentVersionExpirationDays > 0 {
			rule.NoncurrentVersionExpiration = &types.NoncurrentVersionExpiration{
				NoncurrentDays: aws.Int32(safeInt32(r.NoncurrentVersionExpirationDays)),
			}
		}
		if r.TransitionToIADays > 0 {
			rule.Transitions = []types.Transition{{
				Days:         aws.Int32(safeInt32(r.TransitionToIADays)),
				StorageClass: types.TransitionStorageClassStandardIa,
			}}
		}
		s3Rules = append(s3Rules, rule)
	}
	_, err := a.client.PutBucketLifecycleConfiguration(ctx, &s3.PutBucketLifecycleConfigurationInput{
		Bucket: a.bucket(),
		LifecycleConfiguration: &types.BucketLifecycleConfiguration{
			Rules: s3Rules,
		},
	})
	return mapError(err)
}

// GetLifecycleRules returns bucket lifecycle rules.
func (a *Adapter) GetLifecycleRules(ctx context.Context) ([]storage.LifecycleRule, error) {
	out, err := a.client.GetBucketLifecycleConfiguration(ctx, &s3.GetBucketLifecycleConfigurationInput{
		Bucket: a.bucket(),
	})
	if err != nil {
		if mapError(err) == storage.ErrNotFound {
			return nil, nil
		}
		return nil, mapError(err)
	}
	rules := make([]storage.LifecycleRule, 0, len(out.Rules))
	for _, r := range out.Rules {
		rule := storage.LifecycleRule{
			ID:     aws.ToString(r.ID),
			Status: string(r.Status),
			Prefix: aws.ToString(r.Filter.Prefix),
		}
		if r.Expiration != nil && r.Expiration.Days != nil {
			rule.ExpirationDays = int(aws.ToInt32(r.Expiration.Days))
		}
		if r.NoncurrentVersionExpiration != nil && r.NoncurrentVersionExpiration.NoncurrentDays != nil {
			rule.NoncurrentVersionExpirationDays = int(aws.ToInt32(r.NoncurrentVersionExpiration.NoncurrentDays))
		}
		if len(r.Transitions) > 0 && r.Transitions[0].Days != nil {
			rule.TransitionToIADays = int(aws.ToInt32(r.Transitions[0].Days))
		}
		rules = append(rules, rule)
	}
	return rules, nil
}

// DeleteLifecycleRules removes bucket lifecycle configuration.
func (a *Adapter) DeleteLifecycleRules(ctx context.Context) error {
	_, err := a.client.DeleteBucketLifecycle(ctx, &s3.DeleteBucketLifecycleInput{
		Bucket: a.bucket(),
	})
	return mapError(err)
}

// PutBucketNotification configures bucket notifications.
func (a *Adapter) PutBucketNotification(ctx context.Context, cfg storage.NotificationDestination) error {
	if cfg.Type != "sqs" && cfg.Type != "sns" {
		return fmt.Errorf("s3: unsupported notification type %q", cfg.Type)
	}
	events := make([]types.Event, 0, len(cfg.Events))
	for _, ev := range cfg.Events {
		events = append(events, types.Event(ev))
	}
	input := &s3.PutBucketNotificationConfigurationInput{
		Bucket:                    a.bucket(),
		NotificationConfiguration: &types.NotificationConfiguration{},
	}
	switch cfg.Type {
	case "sqs":
		input.NotificationConfiguration.QueueConfigurations = []types.QueueConfiguration{{
			QueueArn: aws.String(cfg.Target),
			Events:   events,
		}}
	case "sns":
		input.NotificationConfiguration.TopicConfigurations = []types.TopicConfiguration{{
			TopicArn: aws.String(cfg.Target),
			Events:   events,
		}}
	}
	_, err := a.client.PutBucketNotificationConfiguration(ctx, input)
	return mapError(err)
}

// GetBucketNotification returns bucket notification configuration.
func (a *Adapter) GetBucketNotification(ctx context.Context) (*storage.NotificationDestination, error) {
	out, err := a.client.GetBucketNotificationConfiguration(ctx, &s3.GetBucketNotificationConfigurationInput{
		Bucket: a.bucket(),
	})
	if err != nil {
		return nil, mapError(err)
	}
	if len(out.QueueConfigurations) > 0 {
		q := out.QueueConfigurations[0]
		dest := &storage.NotificationDestination{Type: "sqs", Target: aws.ToString(q.QueueArn)}
		for _, ev := range q.Events {
			dest.Events = append(dest.Events, storage.NotificationEvent(ev))
		}
		return dest, nil
	}
	if len(out.TopicConfigurations) > 0 {
		t := out.TopicConfigurations[0]
		dest := &storage.NotificationDestination{Type: "sns", Target: aws.ToString(t.TopicArn)}
		for _, ev := range t.Events {
			dest.Events = append(dest.Events, storage.NotificationEvent(ev))
		}
		return dest, nil
	}
	return nil, nil
}

// DeleteBucketNotification removes bucket notification configuration.
func (a *Adapter) DeleteBucketNotification(ctx context.Context) error {
	_, err := a.client.PutBucketNotificationConfiguration(ctx, &s3.PutBucketNotificationConfigurationInput{
		Bucket:                    a.bucket(),
		NotificationConfiguration: &types.NotificationConfiguration{},
	})
	return mapError(err)
}
