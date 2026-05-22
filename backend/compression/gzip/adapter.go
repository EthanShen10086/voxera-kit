// Package gzip provides a gzip implementation of the compression.Compressor interface
// using the compress/gzip standard library package.
package gzip

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"

	"github.com/EthanShen10086/voxera-kit/compression"
)

// Adapter is a gzip-based compressor.
type Adapter struct {
	level int
}

// New creates a new gzip compressor with the given compression level.
// Valid levels range from gzip.HuffmanOnly (-2) to gzip.BestCompression (9).
// Use gzip.DefaultCompression (-1) for the default level.
func New(level int) (*Adapter, error) {
	if level < gzip.HuffmanOnly || level > gzip.BestCompression {
		return nil, fmt.Errorf("gzip: invalid compression level %d", level)
	}
	return &Adapter{level: level}, nil
}

// Compress compresses data using gzip.
func (a *Adapter) Compress(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	w, err := gzip.NewWriterLevel(&buf, a.level)
	if err != nil {
		return nil, fmt.Errorf("gzip: failed to create writer: %w", err)
	}

	if _, err := w.Write(data); err != nil {
		return nil, fmt.Errorf("gzip: failed to write data: %w", err)
	}

	if err := w.Close(); err != nil {
		return nil, fmt.Errorf("gzip: failed to close writer: %w", err)
	}

	return buf.Bytes(), nil
}

// Decompress decompresses gzip data.
func (a *Adapter) Decompress(data []byte) ([]byte, error) {
	r, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("gzip: failed to create reader: %w", err)
	}
	defer func() { _ = r.Close() }()

	result, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("gzip: failed to read data: %w", err)
	}

	return result, nil
}

// Algorithm returns compression.Gzip.
func (a *Adapter) Algorithm() compression.Algorithm {
	return compression.Gzip
}

// ContentEncoding returns "gzip" for use in HTTP Content-Encoding headers.
func (a *Adapter) ContentEncoding() string {
	return "gzip"
}
