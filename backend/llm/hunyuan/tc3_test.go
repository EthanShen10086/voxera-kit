package hunyuan

import (
	"context"
	"net/http"
	"strings"
	"testing"
)

func TestTC3Signer_SetsAuthorization(t *testing.T) {
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, "https://hunyuan.tencentcloudapi.com/hyllm/v1/chat/completions", strings.NewReader(`{"model":"hunyuan-pro"}`))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-TC-Action", "ChatCompletions")

	payload := []byte(`{"model":"hunyuan-pro"}`)
	signer := newTC3Signer("AKIDTEST", "SECRETKEY", "ap-guangzhou")
	signer.sign(req, payload, 1700000000)

	auth := req.Header.Get("Authorization")
	if !strings.HasPrefix(auth, "TC3-HMAC-SHA256 Credential=AKIDTEST/") {
		t.Fatalf("Authorization = %q, want TC3 prefix", auth)
	}
	if req.Header.Get("X-TC-Timestamp") != "1700000000" {
		t.Fatalf("timestamp = %q", req.Header.Get("X-TC-Timestamp"))
	}
}
