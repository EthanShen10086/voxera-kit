package fixture

import (
	"io"
	"net/http"
	"testing"
)

func TestHTTPRequestBuilder_JSON(t *testing.T) {
	builder, err := NewHTTPRequest(http.MethodPost, "https://example.com/v1/items").
		WithBearerToken("token-123").
		WithJSONBody(map[string]string{"name": "demo"})
	if err != nil {
		t.Fatalf("WithJSONBody() = %v", err)
	}

	req, err := builder.Build()
	if err != nil {
		t.Fatalf("Build() = %v", err)
	}
	if req.Method != http.MethodPost {
		t.Fatalf("method = %q, want POST", req.Method)
	}
	if got := req.Header.Get("Authorization"); got != "Bearer token-123" {
		t.Fatalf("Authorization = %q", got)
	}
	if got := req.Header.Get("Content-Type"); got != "application/json" {
		t.Fatalf("Content-Type = %q", got)
	}

	body, err := io.ReadAll(req.Body)
	if err != nil {
		t.Fatalf("ReadAll() = %v", err)
	}
	if string(body) != `{"name":"demo"}` {
		t.Fatalf("body = %q", body)
	}
}

func TestHTTPRequestBuilder_RawBody(t *testing.T) {
	req, err := NewHTTPRequest(http.MethodPut, "https://example.com/raw").
		WithBody(stringsReader("hello"), "text/plain").
		Build()
	if err != nil {
		t.Fatalf("Build() = %v", err)
	}
	body, err := io.ReadAll(req.Body)
	if err != nil {
		t.Fatalf("ReadAll() = %v", err)
	}
	if string(body) != "hello" {
		t.Fatalf("body = %q", body)
	}
}

type stringsReader string

func (s stringsReader) Read(p []byte) (int, error) {
	n := copy(p, []byte(s))
	return n, io.EOF
}
