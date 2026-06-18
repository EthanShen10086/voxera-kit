package headers_test

import (
	"strings"
	"testing"

	"github.com/EthanShen10086/voxera-kit/security/headers"
)

func TestHSTSString(t *testing.T) {
	h := headers.HSTSConfig{MaxAge: 63072000, IncludeSubDomains: true, Preload: true}
	s := h.String()
	if !strings.Contains(s, "max-age=63072000") || !strings.Contains(s, "includeSubDomains") {
		t.Fatalf("HSTS = %q", s)
	}
}

func TestDefaultConfigs(t *testing.T) {
	strict := headers.DefaultStrict()
	if strict.CSP == "" || strict.XFrameOptions != "DENY" {
		t.Fatal("strict config")
	}
	permissive := headers.DefaultPermissive()
	if permissive.XFrameOptions != "SAMEORIGIN" {
		t.Fatal("permissive config")
	}
}
