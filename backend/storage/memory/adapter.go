// Package memory provides an in-memory implementation of storage object store interfaces.
package memory

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/EthanShen10086/voxera-kit/storage"
	"github.com/EthanShen10086/voxera-kit/storage/internal/opts"
	"github.com/EthanShen10086/voxera-kit/storage/internal/uploadlarge"
)

// EventPublisher receives object lifecycle events for testing.
type EventPublisher func(key, event string)

// Options configures the in-memory adapter.
type Options struct {
	EventPublisher EventPublisher
}

// Adapter implements ObjectStore, MultipartUploader, LargeObjectStore,
// VersionedObjectStore, and StorageAdmin in memory.
type Adapter struct {
	mu                sync.RWMutex
	objects           map[string]*storedObject
	versions          map[string][]*objectVersion
	multipart         map[string]*pendingMultipart
	versioningEnabled bool
	lifecycleRules    []storage.LifecycleRule
	notification      *storage.NotificationDestination
	events            EventPublisher
	cfg               storage.Config
	versionCounter    uint64
}

type storedObject struct {
	data         []byte
	contentType  string
	metadata     map[string]string
	etag         string
	lastModified time.Time
}

type objectVersion struct {
	versionID      string
	data           []byte
	contentType    string
	metadata       map[string]string
	etag           string
	lastModified   time.Time
	isLatest       bool
	isDeleteMarker bool
}

type pendingMultipart struct {
	key       string
	opts      storage.UploadOptions
	parts     map[int][]byte
	createdAt time.Time
}

// New creates a new in-memory storage adapter.
func New(cfg storage.Config, options ...Options) *Adapter {
	a := &Adapter{
		objects:   make(map[string]*storedObject),
		versions:  make(map[string][]*objectVersion),
		multipart: make(map[string]*pendingMultipart),
		cfg:       cfg,
	}
	if len(options) > 0 {
		a.events = options[0].EventPublisher
	}
	return a
}

func (a *Adapter) publish(key string, event storage.NotificationEvent) {
	if a.events != nil {
		a.events(key, string(event))
	}
}

func etagFor(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:16])
}

func cloneMetadata(m map[string]string) map[string]string {
	if len(m) == 0 {
		return nil
	}
	out := make(map[string]string, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}

func (a *Adapter) nextVersionID() string {
	n := atomic.AddUint64(&a.versionCounter, 1)
	return fmt.Sprintf("v%d", n)
}

func (a *Adapter) latestVersion(key string) (*objectVersion, bool) {
	versions := a.versions[key]
	for i := len(versions) - 1; i >= 0; i-- {
		if versions[i].isLatest {
			return versions[i], true
		}
	}
	if len(versions) > 0 {
		return versions[len(versions)-1], true
	}
	return nil, false
}

func (a *Adapter) putObject(key string, data []byte, uploadOpts storage.UploadOptions) {
	now := time.Now().UTC()
	etag := etagFor(data)
	if a.versioningEnabled {
		for _, v := range a.versions[key] {
			v.isLatest = false
		}
		ver := &objectVersion{
			versionID:    a.nextVersionID(),
			data:         append([]byte(nil), data...),
			contentType:  uploadOpts.ContentType,
			metadata:     cloneMetadata(uploadOpts.Metadata),
			etag:         etag,
			lastModified: now,
			isLatest:     true,
		}
		a.versions[key] = append(a.versions[key], ver)
		return
	}
	a.objects[key] = &storedObject{
		data:         append([]byte(nil), data...),
		contentType:  uploadOpts.ContentType,
		metadata:     cloneMetadata(uploadOpts.Metadata),
		etag:         etag,
		lastModified: now,
	}
}

// Upload stores an object in memory.
func (a *Adapter) Upload(ctx context.Context, key string, reader io.Reader, uploadOpts *storage.UploadOptions) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	key = opts.NormalizeKey(key)
	data, err := io.ReadAll(reader)
	if err != nil {
		return err
	}

	a.mu.Lock()
	defer a.mu.Unlock()
	a.putObject(key, data, opts.MergeUploadOptions(uploadOpts))
	a.publish(key, storage.EventObjectCreated)
	return nil
}

// Download retrieves an object from memory.
func (a *Adapter) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	key = opts.NormalizeKey(key)

	a.mu.RLock()
	defer a.mu.RUnlock()

	if a.versioningEnabled {
		ver, ok := a.latestVersion(key)
		if !ok || ver.isDeleteMarker {
			return nil, storage.ErrNotFound
		}
		return io.NopCloser(bytes.NewReader(ver.data)), nil
	}

	obj, ok := a.objects[key]
	if !ok {
		return nil, storage.ErrNotFound
	}
	return io.NopCloser(bytes.NewReader(obj.data)), nil
}

// Delete removes an object from memory.
func (a *Adapter) Delete(ctx context.Context, key string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	key = opts.NormalizeKey(key)

	a.mu.Lock()
	defer a.mu.Unlock()

	if a.versioningEnabled {
		if _, ok := a.latestVersion(key); !ok {
			return storage.ErrNotFound
		}
		for _, v := range a.versions[key] {
			v.isLatest = false
		}
		now := time.Now().UTC()
		a.versions[key] = append(a.versions[key], &objectVersion{
			versionID:      a.nextVersionID(),
			lastModified:   now,
			isLatest:       true,
			isDeleteMarker: true,
		})
		a.publish(key, storage.EventObjectRemoved)
		return nil
	}

	if _, ok := a.objects[key]; !ok {
		return storage.ErrNotFound
	}
	delete(a.objects, key)
	a.publish(key, storage.EventObjectRemoved)
	return nil
}

// GetURL returns a memory:// URL for the object.
func (a *Adapter) GetURL(_ context.Context, key string, expiry time.Duration) (string, error) {
	key = opts.NormalizeKey(key)
	return fmt.Sprintf("memory://%s?expiry=%s", key, expiry), nil
}

// List returns metadata for objects matching the prefix.
func (a *Adapter) List(_ context.Context, prefix string) ([]*storage.ObjectMeta, error) {
	prefix = opts.NormalizeKey(prefix)

	a.mu.RLock()
	defer a.mu.RUnlock()

	var out []*storage.ObjectMeta
	if a.versioningEnabled {
		for key := range a.versions {
			if !strings.HasPrefix(key, prefix) {
				continue
			}
			ver, ok := a.latestVersion(key)
			if !ok || ver.isDeleteMarker {
				continue
			}
			out = append(out, &storage.ObjectMeta{
				Key:          key,
				Size:         int64(len(ver.data)),
				ContentType:  ver.contentType,
				ETag:         ver.etag,
				LastModified: ver.lastModified,
			})
		}
	} else {
		for key, obj := range a.objects {
			if !strings.HasPrefix(key, prefix) {
				continue
			}
			out = append(out, &storage.ObjectMeta{
				Key:          key,
				Size:         int64(len(obj.data)),
				ContentType:  obj.contentType,
				ETag:         obj.etag,
				LastModified: obj.lastModified,
			})
		}
	}

	sort.Slice(out, func(i, j int) bool { return out[i].Key < out[j].Key })
	return out, nil
}

// Exists checks whether an object exists in memory.
func (a *Adapter) Exists(_ context.Context, key string) (bool, error) {
	key = opts.NormalizeKey(key)

	a.mu.RLock()
	defer a.mu.RUnlock()

	if a.versioningEnabled {
		ver, ok := a.latestVersion(key)
		return ok && !ver.isDeleteMarker, nil
	}
	_, ok := a.objects[key]
	return ok, nil
}

// Close releases resources held by the adapter.
func (a *Adapter) Close() error {
	return nil
}

// UploadLarge uploads a large object using multipart when above threshold.
func (a *Adapter) UploadLarge(ctx context.Context, key string, reader io.ReaderAt, size int64, uploadOpts *storage.UploadOptions) error {
	return uploadlarge.Upload(ctx, a, a, a.cfg, key, reader, size, uploadOpts)
}

// InitiateMultipartUpload starts a multipart upload session.
func (a *Adapter) InitiateMultipartUpload(_ context.Context, key string, uploadOpts *storage.UploadOptions) (string, error) {
	key = opts.NormalizeKey(key)
	uploadID := fmt.Sprintf("mp-%d", time.Now().UnixNano())

	a.mu.Lock()
	defer a.mu.Unlock()
	a.multipart[uploadID] = &pendingMultipart{
		key:       key,
		opts:      opts.MergeUploadOptions(uploadOpts),
		parts:     make(map[int][]byte),
		createdAt: time.Now().UTC(),
	}
	return uploadID, nil
}

// UploadPart stores one multipart upload part.
func (a *Adapter) UploadPart(_ context.Context, key, uploadID string, partNumber int, reader io.Reader, size int64) (string, error) {
	key = opts.NormalizeKey(key)
	data, err := io.ReadAll(reader)
	if err != nil {
		return "", err
	}
	if size >= 0 && int64(len(data)) != size {
		return "", fmt.Errorf("part size mismatch: expected %d, got %d", size, len(data))
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	mp, ok := a.multipart[uploadID]
	if !ok || mp.key != key {
		return "", storage.ErrNotFound
	}
	mp.parts[partNumber] = append([]byte(nil), data...)
	return etagFor(data), nil
}

// CompleteMultipartUpload assembles uploaded parts into a final object.
func (a *Adapter) CompleteMultipartUpload(_ context.Context, key, uploadID string, parts []storage.CompletedPart) error {
	key = opts.NormalizeKey(key)

	a.mu.Lock()
	defer a.mu.Unlock()

	mp, ok := a.multipart[uploadID]
	if !ok || mp.key != key {
		return storage.ErrNotFound
	}

	numbers := make([]int, 0, len(parts))
	for _, p := range parts {
		numbers = append(numbers, p.PartNumber)
	}
	sort.Ints(numbers)

	var assembled []byte
	for _, n := range numbers {
		partData, ok := mp.parts[n]
		if !ok {
			return fmt.Errorf("missing part %d", n)
		}
		assembled = append(assembled, partData...)
	}

	a.putObject(key, assembled, mp.opts)
	delete(a.multipart, uploadID)
	a.publish(key, storage.EventObjectCreated)
	return nil
}

// AbortMultipartUpload cancels an in-progress multipart upload.
func (a *Adapter) AbortMultipartUpload(_ context.Context, key, uploadID string) error {
	key = opts.NormalizeKey(key)

	a.mu.Lock()
	defer a.mu.Unlock()

	mp, ok := a.multipart[uploadID]
	if !ok || mp.key != key {
		return storage.ErrNotFound
	}
	delete(a.multipart, uploadID)
	return nil
}

// ListVersions returns all versions for a key.
func (a *Adapter) ListVersions(_ context.Context, key string) ([]*storage.ObjectVersion, error) {
	key = opts.NormalizeKey(key)

	a.mu.RLock()
	defer a.mu.RUnlock()

	if !a.versioningEnabled {
		return nil, storage.ErrVersioningDisabled
	}

	versions := a.versions[key]
	if len(versions) == 0 {
		return nil, storage.ErrNotFound
	}

	out := make([]*storage.ObjectVersion, 0, len(versions))
	for _, v := range versions {
		out = append(out, &storage.ObjectVersion{
			VersionID:      v.versionID,
			Key:            key,
			Size:           int64(len(v.data)),
			IsLatest:       v.isLatest,
			IsDeleteMarker: v.isDeleteMarker,
			LastModified:   v.lastModified,
		})
	}
	return out, nil
}

// DownloadVersion retrieves a specific object version.
func (a *Adapter) DownloadVersion(_ context.Context, key, versionID string) (io.ReadCloser, error) {
	key = opts.NormalizeKey(key)

	a.mu.RLock()
	defer a.mu.RUnlock()

	if !a.versioningEnabled {
		return nil, storage.ErrVersioningDisabled
	}

	for _, v := range a.versions[key] {
		if v.versionID == versionID {
			if v.isDeleteMarker {
				return nil, storage.ErrNotFound
			}
			return io.NopCloser(bytes.NewReader(v.data)), nil
		}
	}
	return nil, storage.ErrNotFound
}

// DeleteVersion removes a specific object version.
func (a *Adapter) DeleteVersion(_ context.Context, key, versionID string) error {
	key = opts.NormalizeKey(key)

	a.mu.Lock()
	defer a.mu.Unlock()

	if !a.versioningEnabled {
		return storage.ErrVersioningDisabled
	}

	versions := a.versions[key]
	for i, v := range versions {
		if v.versionID != versionID {
			continue
		}
		wasLatest := v.isLatest
		versions = append(versions[:i], versions[i+1:]...)
		a.versions[key] = versions
		if wasLatest && len(versions) > 0 {
			versions[len(versions)-1].isLatest = true
		}
		return nil
	}
	return storage.ErrNotFound
}

// RestoreVersion makes a historical version the latest object.
func (a *Adapter) RestoreVersion(_ context.Context, key, versionID string) error {
	key = opts.NormalizeKey(key)

	a.mu.Lock()
	defer a.mu.Unlock()

	if !a.versioningEnabled {
		return storage.ErrVersioningDisabled
	}

	var target *objectVersion
	for _, v := range a.versions[key] {
		if v.versionID == versionID {
			target = v
			break
		}
	}
	if target == nil || target.isDeleteMarker {
		return storage.ErrNotFound
	}

	for _, v := range a.versions[key] {
		v.isLatest = false
	}
	now := time.Now().UTC()
	restored := &objectVersion{
		versionID:    a.nextVersionID(),
		data:         append([]byte(nil), target.data...),
		contentType:  target.contentType,
		metadata:     cloneMetadata(target.metadata),
		etag:         target.etag,
		lastModified: now,
		isLatest:     true,
	}
	a.versions[key] = append(a.versions[key], restored)
	a.publish(key, storage.EventObjectCreated)
	return nil
}

// EnableVersioning toggles object versioning.
func (a *Adapter) EnableVersioning(_ context.Context, enabled bool) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.versioningEnabled = enabled
	return nil
}

// GetVersioning returns whether versioning is enabled.
func (a *Adapter) GetVersioning(_ context.Context) (bool, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.versioningEnabled, nil
}

// PutLifecycleRules stores lifecycle rules in memory.
func (a *Adapter) PutLifecycleRules(_ context.Context, rules []storage.LifecycleRule) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.lifecycleRules = append([]storage.LifecycleRule(nil), rules...)
	return nil
}

// GetLifecycleRules returns stored lifecycle rules.
func (a *Adapter) GetLifecycleRules(_ context.Context) ([]storage.LifecycleRule, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return append([]storage.LifecycleRule(nil), a.lifecycleRules...), nil
}

// DeleteLifecycleRules removes all lifecycle rules.
func (a *Adapter) DeleteLifecycleRules(_ context.Context) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.lifecycleRules = nil
	return nil
}

// PutBucketNotification stores bucket notification configuration.
func (a *Adapter) PutBucketNotification(_ context.Context, cfg storage.NotificationDestination) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	copied := cfg
	if len(cfg.Events) > 0 {
		copied.Events = append([]storage.NotificationEvent(nil), cfg.Events...)
	}
	a.notification = &copied
	return nil
}

// GetBucketNotification returns stored notification configuration.
func (a *Adapter) GetBucketNotification(_ context.Context) (*storage.NotificationDestination, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if a.notification == nil {
		return nil, nil
	}
	copied := *a.notification
	if len(a.notification.Events) > 0 {
		copied.Events = append([]storage.NotificationEvent(nil), a.notification.Events...)
	}
	return &copied, nil
}

// DeleteBucketNotification removes bucket notification configuration.
func (a *Adapter) DeleteBucketNotification(_ context.Context) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.notification = nil
	return nil
}
