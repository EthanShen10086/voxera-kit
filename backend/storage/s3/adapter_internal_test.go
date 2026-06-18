package s3

import (
	"errors"
	"net/http"
	"testing"

	"github.com/EthanShen10086/voxera-kit/storage"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	smithyhttp "github.com/aws/smithy-go/transport/http"
)

func TestMapError(t *testing.T) {
	if mapError(nil) != nil {
		t.Fatal("nil should stay nil")
	}
	respErr := &smithyhttp.ResponseError{Response: &smithyhttp.Response{Response: &http.Response{StatusCode: http.StatusNotFound}}}
	if !errors.Is(mapError(respErr), storage.ErrNotFound) {
		t.Fatal("expected not found from 404")
	}
	if !errors.Is(mapError(&types.NoSuchKey{}), storage.ErrNotFound) {
		t.Fatal("expected not found from NoSuchKey")
	}
	if !errors.Is(mapError(&types.NotFound{}), storage.ErrNotFound) {
		t.Fatal("expected not found from NotFound")
	}
	if !errors.Is(mapError(errors.New("NoSuchKey")), storage.ErrNotFound) {
		t.Fatal("expected not found from message")
	}
	if mapError(errors.New("other")) == nil {
		t.Fatal("expected original error")
	}
}

func TestSafeInt32(t *testing.T) {
	if safeInt32(1<<30) != 1<<30 {
		t.Fatal("expected in-range value")
	}
	if safeInt32(1<<40) != 1<<31-1 {
		t.Fatalf("expected max int32, got %d", safeInt32(1<<40))
	}
	if safeInt32(-1<<40) != -1<<31 {
		t.Fatalf("expected min int32, got %d", safeInt32(-1<<40))
	}
}

func TestUploadInputMetadata(t *testing.T) {
	a := &Adapter{cfg: storage.Config{Bucket: "b"}}
	in := a.uploadInput("key", &storage.UploadOptions{
		ContentType: "text/plain",
		Metadata:    map[string]string{"x": "y"},
	})
	if aws.ToString(in.Key) != "key" || aws.ToString(in.ContentType) != "text/plain" {
		t.Fatalf("input = %#v", in)
	}
	if in.Metadata["x"] != "y" {
		t.Fatalf("metadata = %#v", in.Metadata)
	}
}
