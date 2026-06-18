package profiling_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/EthanShen10086/voxera-kit/observability/profiling"
)

func TestRegisterPprofDisabled(t *testing.T) {
	mux := http.NewServeMux()
	profiling.RegisterPprof(mux, false)
	req := httptest.NewRequest(http.MethodGet, "/debug/pprof/", nil)
	_, pattern := mux.Handler(req)
	if pattern != "" {
		t.Fatalf("expected no route, got pattern %q", pattern)
	}
}

func TestRegisterPprofEnabled(t *testing.T) {
	mux := http.NewServeMux()
	profiling.RegisterPprof(mux, true)
	for _, path := range []string{
		"/debug/pprof/",
		"/debug/pprof/cmdline",
		"/debug/pprof/profile",
		"/debug/pprof/symbol",
		"/debug/pprof/trace",
	} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		_, pattern := mux.Handler(req)
		if pattern == "" {
			t.Fatalf("missing route for %s", path)
		}
	}
	srv := httptest.NewServer(mux)
	defer srv.Close()
	resp, err := http.Get(srv.URL + "/debug/pprof/")
	if err != nil {
		t.Fatal(err)
	}
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", resp.StatusCode)
	}
	body := resp.Request.URL.Path
	if !strings.HasPrefix(body, "/debug/pprof") {
		t.Fatalf("path = %q", body)
	}
}
