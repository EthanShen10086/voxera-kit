package gzip_test

import (
	"testing"

	compressgzip "compress/gzip"

	"github.com/EthanShen10086/voxera-kit/compression"
	"github.com/EthanShen10086/voxera-kit/compression/gzip"
)

func TestCompressDecompressRoundtrip(t *testing.T) {
	a, err := gzip.New(compressgzip.DefaultCompression)
	if err != nil {
		t.Fatal(err)
	}
	original := []byte("hello gzip compression test data")
	compressed, err := a.Compress(original)
	if err != nil {
		t.Fatal(err)
	}
	out, err := a.Decompress(compressed)
	if err != nil {
		t.Fatal(err)
	}
	if string(out) != string(original) {
		t.Fatalf("got %q", out)
	}
	if a.Algorithm() != compression.Gzip {
		t.Fatalf("algorithm = %v", a.Algorithm())
	}
	if a.ContentEncoding() != "gzip" {
		t.Fatalf("encoding = %q", a.ContentEncoding())
	}
}

func TestNewInvalidLevel(t *testing.T) {
	_, err := gzip.New(99)
	if err == nil {
		t.Fatal("expected invalid level error")
	}
}
