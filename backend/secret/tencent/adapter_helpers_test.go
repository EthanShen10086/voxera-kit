package tencent

import (
	"errors"
	"testing"

	tcerr "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
)

func TestIsNotFound(t *testing.T) {
	if isNotFound(nil) {
		t.Fatal("nil should not be not found")
	}
	sdkErr := &tcerr.TencentCloudSDKError{Code: "ResourceNotFound.SecretNotFound"}
	if !isNotFound(sdkErr) {
		t.Fatal("expected SDK not found")
	}
	if isNotFound(errors.New("internal error")) {
		t.Fatal("unexpected not found")
	}
	if !isNotFound(errors.New("secret not found in region")) {
		t.Fatal("expected message match")
	}
	sdkErr2 := &tcerr.TencentCloudSDKError{Code: "FailedOperation.ResourceNotFound"}
	if !isNotFound(sdkErr2) {
		t.Fatal("expected FailedOperation not found")
	}
}
