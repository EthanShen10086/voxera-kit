package shorturl_test

import (
	"testing"
	"time"

	"github.com/EthanShen10086/voxera-kit/shorturl"
)

func TestResolveOptions(t *testing.T) {
	params := shorturl.ResolveOptions([]shorturl.GenerateOption{
		shorturl.WithExpiry(time.Hour),
		shorturl.WithCustomCode("abc"),
		shorturl.WithCreator("user-1"),
		shorturl.WithMetadata(map[string]string{"k": "v"}),
	})
	if params.Expiry != time.Hour || params.CustomCode != "abc" || params.Creator != "user-1" {
		t.Fatalf("params = %+v", params)
	}
	if params.Metadata["k"] != "v" {
		t.Fatalf("metadata = %+v", params.Metadata)
	}
}
