package uploadlarge

import (
	"context"
	"fmt"
	"io"

	"github.com/EthanShen10086/voxera-kit/storage"
	"github.com/EthanShen10086/voxera-kit/storage/internal/opts"
)

// Upload performs a multipart upload when size exceeds threshold, otherwise a single Upload.
func Upload(
	ctx context.Context,
	store storage.ObjectStore,
	multipart storage.MultipartUploader,
	cfg storage.Config,
	key string,
	reader io.ReaderAt,
	size int64,
	uploadOpts *storage.UploadOptions,
) error {
	key = opts.NormalizeKey(key)
	threshold := opts.MultipartThreshold(cfg)
	if size <= threshold {
		return store.Upload(ctx, key, io.NewSectionReader(reader, 0, size), uploadOpts)
	}

	partSize := opts.PartSize(cfg)
	uploadID, err := multipart.InitiateMultipartUpload(ctx, key, uploadOpts)
	if err != nil {
		return err
	}

	var parts []storage.CompletedPart
	partNumber := 1
	for offset := int64(0); offset < size; offset += partSize {
		end := offset + partSize
		if end > size {
			end = size
		}
		chunkSize := end - offset
		section := io.NewSectionReader(reader, offset, chunkSize)
		etag, err := multipart.UploadPart(ctx, key, uploadID, partNumber, section, chunkSize)
		if err != nil {
			_ = multipart.AbortMultipartUpload(ctx, key, uploadID)
			return err
		}
		parts = append(parts, storage.CompletedPart{PartNumber: partNumber, ETag: etag})
		partNumber++
	}

	if err := multipart.CompleteMultipartUpload(ctx, key, uploadID, parts); err != nil {
		_ = multipart.AbortMultipartUpload(ctx, key, uploadID)
		return fmt.Errorf("complete multipart: %w", err)
	}
	return nil
}
