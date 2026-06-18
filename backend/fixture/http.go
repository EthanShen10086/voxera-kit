package fixture

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// HTTPRequestBuilder constructs *http.Request values for handler and middleware tests.
type HTTPRequestBuilder struct {
	method  string
	target  string
	header  http.Header
	body    io.Reader
	context map[string]string
}

// NewHTTPRequest starts building an HTTP request for the given method and URL.
func NewHTTPRequest(method, target string) *HTTPRequestBuilder {
	return &HTTPRequestBuilder{
		method: strings.ToUpper(method),
		target: target,
		header: make(http.Header),
	}
}

// WithHeader sets a request header.
func (b *HTTPRequestBuilder) WithHeader(key, value string) *HTTPRequestBuilder {
	b.header.Set(key, value)
	return b
}

// WithBearerToken sets the Authorization header to a bearer token.
func (b *HTTPRequestBuilder) WithBearerToken(token string) *HTTPRequestBuilder {
	return b.WithHeader("Authorization", "Bearer "+token)
}

// WithJSONBody serializes payload as JSON and sets Content-Type.
func (b *HTTPRequestBuilder) WithJSONBody(payload any) (*HTTPRequestBuilder, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return b, fmt.Errorf("fixture: marshal json body: %w", err)
	}
	b.body = bytes.NewReader(data)
	b.header.Set("Content-Type", "application/json")
	return b, nil
}

// WithBody sets a raw request body and optional content type.
func (b *HTTPRequestBuilder) WithBody(body io.Reader, contentType string) *HTTPRequestBuilder {
	b.body = body
	if contentType != "" {
		b.header.Set("Content-Type", contentType)
	}
	return b
}

// Build returns the constructed request.
func (b *HTTPRequestBuilder) Build() (*http.Request, error) {
	req, err := http.NewRequest(b.method, b.target, b.body)
	if err != nil {
		return nil, fmt.Errorf("fixture: build request: %w", err)
	}
	for key, values := range b.header {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}
	return req, nil
}
