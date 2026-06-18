package gcp

import (
	"context"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestNewManagerMissingProject(t *testing.T) {
	_, err := NewManager(context.Background(), Config{ProjectID: ""})
	if err == nil {
		t.Fatal("expected error for missing project ID")
	}
}

func TestSecretIDFromName(t *testing.T) {
	if got := secretIDFromName("projects/p/secrets/my-key"); got != "my-key" {
		t.Fatalf("secretIDFromName = %q", got)
	}
	if got := secretIDFromName("projects/p/secrets/prefix/a"); got != "prefix/a" {
		t.Fatalf("nested secretIDFromName = %q", got)
	}
	if got := secretIDFromName(""); got != "" {
		t.Fatalf("empty name = %q", got)
	}
}

func TestIsNotFound(t *testing.T) {
	if isNotFound(nil) {
		t.Fatal("nil should not be not found")
	}
	if !isNotFound(status.Error(codes.NotFound, "missing")) {
		t.Fatal("expected not found")
	}
	if isNotFound(status.Error(codes.Internal, "err")) {
		t.Fatal("unexpected not found")
	}
}
