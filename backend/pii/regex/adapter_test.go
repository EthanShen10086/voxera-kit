package regex_test

import (
	"testing"

	"github.com/EthanShen10086/voxera-kit/pii"
	"github.com/EthanShen10086/voxera-kit/pii/regex"
)

func TestRedactDefaultRules(t *testing.T) {
	rules := regex.DefaultRules()
	r := regex.NewRedactor(pii.Config{Rules: rules, DefaultMask: "[REDACTED]"})

	input := "contact alice@example.com or 555-123-4567 from 192.168.1.1"
	out := r.Redact(input)
	if out == input {
		t.Fatal("expected redaction")
	}
	for _, s := range []string{"alice@example.com", "555-123-4567", "192.168.1.1"} {
		if contains(out, s) {
			t.Fatalf("still contains %q: %s", s, out)
		}
	}
}

func TestRedactFields(t *testing.T) {
	r := regex.NewRedactor(pii.Config{
		Rules: []pii.Rule{{FieldName: "ssn", Replacement: "***"}},
	})
	out := r.RedactFields(map[string]any{
		"ssn": "123-45-6789",
		"note": map[string]any{"email": "x@y.com"},
	})
	if out["ssn"] != "***" {
		t.Fatalf("ssn = %v", out["ssn"])
	}
}

func contains(s, sub string) bool {
	return len(sub) == 0 || (len(s) >= len(sub) && (s == sub || len(s) > 0 && indexOf(s, sub) >= 0))
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
