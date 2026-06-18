// Package fs provides a filesystem-backed implementation of storage object store interfaces.
package fs

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/EthanShen10086/voxera-kit/storage"
	"github.com/EthanShen10086/voxera-kit/storage/internal/opts"
	"github.com/EthanShen10086/voxera-kit/storage/internal/uploadlarge"
)

const multipartDir = ".multipart"

// Adapter implements ObjectStore, MultipartUploader, and LargeObjectStore on local disk.
type Adapter struct {
	root      string
	cfg       storage.Config
	mu        sync.Mutex
	multipart map[string]*pendingMultipart
}

type pendingMultipart struct {
	key       string
	opts      storage.UploadOptions
	partsDir  string
	parts     map[int]string
	createdAt time.Time
}

// New creates a filesystem-backed adapter rooted at the given directory.
func New(root string, cfg storage.Config) (*Adapter, error) {
	if err := os.MkdirAll(root, 0o755); err != nil {
		return nil, err
	}
	return &Adapter{
		root:      root,
		cfg:       cfg,
		multipart: make(map[string]*pendingMultipart),
	}, nil
}

func (a *Adapter) objectPath(key string) (string, error) {
	key = opts.NormalizeKey(key)
	if key == "" || strings.Contains(key, "..") {
		return "", fmt.Errorf("invalid key %q", key)
	}
	path := filepath.Join(a.root, filepath.FromSlash(key))
	cleanRoot := filepath.Clean(a.root) + string(os.PathSeparator)
	cleanPath := filepath.Clean(path)
	if cleanPath != filepath.Clean(a.root) && !strings.HasPrefix(cleanPath, cleanRoot) {
		return "", fmt.Errorf("invalid key %q", key)
	}
	return path, nil
}

func (a *Adapter) metaPath(objectPath string) string {
	return objectPath + ".meta"
}

func etagForFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := md5.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func (a *Adapter) statObject(path string) (*storage.ObjectMeta, error) {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, storage.ErrNotFound
		}
		return nil, err
	}
	if info.IsDir() {
		return nil, storage.ErrNotFound
	}
	etag, err := etagForFile(path)
	if err != nil {
		return nil, err
	}
	rel, err := filepath.Rel(a.root, path)
	if err != nil {
		return nil, err
	}
	return &storage.ObjectMeta{
		Key:          filepath.ToSlash(rel),
		Size:         info.Size(),
		ETag:         etag,
		LastModified: info.ModTime().UTC(),
	}, nil
}

// Upload stores an object on disk.
func (a *Adapter) Upload(ctx context.Context, key string, reader io.Reader, uploadOpts *storage.UploadOptions) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	path, err := a.objectPath(key)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	tmp, err := os.CreateTemp(filepath.Dir(path), ".upload-*")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	defer func() {
		_ = tmp.Close()
		_ = os.Remove(tmpPath)
	}()

	if _, err := io.Copy(tmp, reader); err != nil {
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmpPath, path); err != nil {
		return err
	}
	_ = opts.MergeUploadOptions(uploadOpts)
	return nil
}

// Download retrieves an object from disk.
func (a *Adapter) Download(_ context.Context, key string) (io.ReadCloser, error) {
	path, err := a.objectPath(key)
	if err != nil {
		return nil, err
	}
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, storage.ErrNotFound
		}
		return nil, err
	}
	return f, nil
}

// Delete removes an object from disk.
func (a *Adapter) Delete(_ context.Context, key string) error {
	path, err := a.objectPath(key)
	if err != nil {
		return err
	}
	if err := os.Remove(path); err != nil {
		if os.IsNotExist(err) {
			return storage.ErrNotFound
		}
		return err
	}
	_ = os.Remove(a.metaPath(path))
	return nil
}

// GetURL returns a file:// URL for the object.
func (a *Adapter) GetURL(_ context.Context, key string, expiry time.Duration) (string, error) {
	path, err := a.objectPath(key)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("file://%s?expiry=%s", path, expiry), nil
}

// List returns metadata for objects matching the prefix.
func (a *Adapter) List(_ context.Context, prefix string) ([]*storage.ObjectMeta, error) {
	prefix = opts.NormalizeKey(prefix)
	searchRoot := a.root
	if prefix != "" {
		var err error
		searchRoot, err = a.objectPath(prefix)
		if err != nil {
			return nil, err
		}
	}

	var out []*storage.ObjectMeta
	err := filepath.Walk(searchRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return err
		}
		if info.IsDir() {
			return nil
		}
		if strings.HasSuffix(path, ".meta") {
			return nil
		}
		rel, err := filepath.Rel(a.root, path)
		if err != nil {
			return err
		}
		key := filepath.ToSlash(rel)
		if prefix != "" && !strings.HasPrefix(key, prefix) {
			return nil
		}
		meta, err := a.statObject(path)
		if err != nil {
			return err
		}
		out = append(out, meta)
		return nil
	})
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	sort.Slice(out, func(i, j int) bool { return out[i].Key < out[j].Key })
	return out, nil
}

// Exists checks whether an object exists on disk.
func (a *Adapter) Exists(_ context.Context, key string) (bool, error) {
	path, err := a.objectPath(key)
	if err != nil {
		return false, err
	}
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return !info.IsDir(), nil
}

// Close releases resources held by the adapter.
func (a *Adapter) Close() error {
	return nil
}

// UploadLarge uploads a large object using multipart when above threshold.
func (a *Adapter) UploadLarge(ctx context.Context, key string, reader io.ReaderAt, size int64, uploadOpts *storage.UploadOptions) error {
	return uploadlarge.Upload(ctx, a, a, a.cfg, key, reader, size, uploadOpts)
}

// InitiateMultipartUpload starts a multipart upload session using temp files.
func (a *Adapter) InitiateMultipartUpload(_ context.Context, key string, uploadOpts *storage.UploadOptions) (string, error) {
	key = opts.NormalizeKey(key)
	uploadID := fmt.Sprintf("mp-%d", time.Now().UnixNano())
	partsDir := filepath.Join(a.root, multipartDir, uploadID)
	if err := os.MkdirAll(partsDir, 0o755); err != nil {
		return "", err
	}

	a.mu.Lock()
	defer a.mu.Unlock()
	a.multipart[uploadID] = &pendingMultipart{
		key:       key,
		opts:      opts.MergeUploadOptions(uploadOpts),
		partsDir:  partsDir,
		parts:     make(map[int]string),
		createdAt: time.Now().UTC(),
	}
	return uploadID, nil
}

// UploadPart stores one multipart upload part in a temp file.
func (a *Adapter) UploadPart(_ context.Context, key, uploadID string, partNumber int, reader io.Reader, size int64) (string, error) {
	key = opts.NormalizeKey(key)

	a.mu.Lock()
	mp, ok := a.multipart[uploadID]
	a.mu.Unlock()
	if !ok || mp.key != key {
		return "", storage.ErrNotFound
	}

	partPath := filepath.Join(mp.partsDir, fmt.Sprintf("part-%05d", partNumber))
	f, err := os.Create(partPath)
	if err != nil {
		return "", err
	}
	written, err := io.Copy(f, reader)
	if closeErr := f.Close(); closeErr != nil && err == nil {
		err = closeErr
	}
	if err != nil {
		_ = os.Remove(partPath)
		return "", err
	}
	if size >= 0 && written != size {
		_ = os.Remove(partPath)
		return "", fmt.Errorf("part size mismatch: expected %d, got %d", size, written)
	}

	a.mu.Lock()
	mp.parts[partNumber] = partPath
	a.mu.Unlock()

	etag, err := etagForFile(partPath)
	if err != nil {
		return "", err
	}
	return etag, nil
}

// CompleteMultipartUpload assembles uploaded parts into a final object.
func (a *Adapter) CompleteMultipartUpload(_ context.Context, key, uploadID string, parts []storage.CompletedPart) error {
	key = opts.NormalizeKey(key)

	a.mu.Lock()
	mp, ok := a.multipart[uploadID]
	if !ok || mp.key != key {
		a.mu.Unlock()
		return storage.ErrNotFound
	}
	delete(a.multipart, uploadID)
	a.mu.Unlock()
	defer os.RemoveAll(mp.partsDir)

	numbers := make([]int, 0, len(parts))
	for _, p := range parts {
		numbers = append(numbers, p.PartNumber)
	}
	sort.Ints(numbers)

	dest, err := a.objectPath(key)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(dest), 0o750); err != nil {
		return err
	}

	out, err := os.Create(dest) // #nosec G304 -- path validated by objectPath
	if err != nil {
		return err
	}
	for _, n := range numbers {
		partPath, ok := mp.parts[n]
		if !ok {
			_ = out.Close()
			_ = os.Remove(dest)
			return fmt.Errorf("missing part %d", n)
		}
		in, err := os.Open(partPath) // #nosec G304 -- part files created under controlled temp dir
		if err != nil {
			_ = out.Close()
			_ = os.Remove(dest)
			return err
		}
		_, copyErr := io.Copy(out, in)
		_ = in.Close()
		if copyErr != nil {
			_ = out.Close()
			_ = os.Remove(dest)
			return copyErr
		}
	}
	return out.Close()
}

// AbortMultipartUpload cancels an in-progress multipart upload.
func (a *Adapter) AbortMultipartUpload(_ context.Context, key, uploadID string) error {
	key = opts.NormalizeKey(key)

	a.mu.Lock()
	mp, ok := a.multipart[uploadID]
	if !ok || mp.key != key {
		a.mu.Unlock()
		return storage.ErrNotFound
	}
	delete(a.multipart, uploadID)
	a.mu.Unlock()
	return os.RemoveAll(mp.partsDir)
}
