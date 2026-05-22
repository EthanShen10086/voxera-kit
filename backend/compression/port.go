// Package compression defines the port interface for data compression and
// decompression operations. It abstracts away the underlying algorithm
// (gzip, deflate, brotli, zstd), allowing different backends to be used
// interchangeably.
package compression

// Algorithm represents a compression algorithm.
type Algorithm int

const (
	// Gzip is the gzip compression algorithm.
	Gzip Algorithm = iota
	// Deflate is the deflate compression algorithm.
	Deflate
	// Brotli is the Brotli compression algorithm.
	Brotli
	// Zstd is the Zstandard compression algorithm.
	Zstd
	// None disables compression.
	None
)

// Config holds general compression parameters.
type Config struct {
	// Enabled controls whether compression is active.
	Enabled bool
	// Algorithm is the compression algorithm to use.
	Algorithm Algorithm
	// Level is the compression level (algorithm-specific).
	Level int
	// MinSize is the minimum payload size in bytes before compression is applied.
	MinSize int
	// ContentTypes lists the MIME types eligible for compression.
	ContentTypes []string
}

// Compressor is the interface for compressing and decompressing byte slices.
// Implementations must be safe for concurrent use.
type Compressor interface {
	// Compress compresses the given data.
	Compress(data []byte) ([]byte, error)
	// Decompress decompresses the given data.
	Decompress(data []byte) ([]byte, error)
	// Algorithm returns the compression algorithm used by this compressor.
	Algorithm() Algorithm
	// ContentEncoding returns the HTTP Content-Encoding header value.
	ContentEncoding() string
}

// CompressorFactory is the interface for creating Compressor instances.
type CompressorFactory interface {
	// Create returns a Compressor for the given algorithm and compression level.
	Create(algorithm Algorithm, level int) (Compressor, error)
}

// HTTPCompressionConfig holds configuration for HTTP-layer compression middleware.
type HTTPCompressionConfig struct {
	// Enabled controls whether HTTP compression is active.
	Enabled bool
	// Algorithms lists the compression algorithms to support, in preference order.
	Algorithms []Algorithm
	// MinSize is the minimum response size in bytes before compression is applied.
	MinSize int
	// ExcludePaths lists URL path prefixes to exclude from compression.
	ExcludePaths []string
	// IncludeTypes lists the MIME types eligible for compression.
	IncludeTypes []string
}
