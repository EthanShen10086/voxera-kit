package fixture

import (
	"strings"
	"testing"
)

func TestNewID(t *testing.T) {
	id := NewID()
	if len(id) != 32 {
		t.Fatalf("NewID() length = %d, want 32", len(id))
	}
	if id == NewID() {
		t.Fatal("expected unique IDs")
	}
}

func TestNewPrefixedID(t *testing.T) {
	id := NewPrefixedID("task")
	if !strings.HasPrefix(id, "task-") {
		t.Fatalf("NewPrefixedID() = %q, want task- prefix", id)
	}
}

func TestNewSequentialID(t *testing.T) {
	ResetSequentialIDs()
	first := NewSequentialID("item")
	second := NewSequentialID("item")
	if first != "item-1" {
		t.Fatalf("first = %q, want item-1", first)
	}
	if second != "item-2" {
		t.Fatalf("second = %q, want item-2", second)
	}
}
