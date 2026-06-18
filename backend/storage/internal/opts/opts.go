package opts

import (
	"strings"

	"github.com/EthanShen10086/voxera-kit/storage"
)

// NormalizeKey trims leading slashes from object keys.
func NormalizeKey(key string) string {
	return strings.TrimPrefix(key, "/")
}

// MergeUploadOptions returns defaults merged with optional overrides.
func MergeUploadOptions(opts *storage.UploadOptions) storage.UploadOptions {
	if opts == nil {
		return storage.UploadOptions{}
	}
	return *opts
}

// PartSize returns the configured part size or DefaultPartSize.
func PartSize(cfg storage.Config) int64 {
	if cfg.PartSize > 0 {
		return cfg.PartSize
	}
	return storage.DefaultPartSize
}

// MultipartThreshold returns the configured threshold or DefaultMultipartThreshold.
func MultipartThreshold(cfg storage.Config) int64 {
	if cfg.MultipartThreshold > 0 {
		return cfg.MultipartThreshold
	}
	return storage.DefaultMultipartThreshold
}
