package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestChainOrder(t *testing.T) {
	var order []string
	h := Chain(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		order = append(order, "handler")
	}),
		func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				order = append(order, "outer")
				next.ServeHTTP(w, r)
			})
		},
		func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				order = append(order, "inner")
				next.ServeHTTP(w, r)
			})
		},
	)
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil))
	if len(order) != 3 || order[0] != "outer" || order[1] != "inner" || order[2] != "handler" {
		t.Fatalf("order %v", order)
	}
}

func TestRequestIDPropagates(t *testing.T) {
	var got string
	h := Chain(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = RequestIDFromContext(r.Context())
	}), RequestID())
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Request-ID", "test-req-id")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if got != "test-req-id" {
		t.Fatalf("context id %q", got)
	}
	if rec.Header().Get("X-Request-ID") != "test-req-id" {
		t.Fatal("response header missing id")
	}
}
