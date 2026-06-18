package share_test

import (
	"testing"
	"time"

	"github.com/EthanShen10086/voxera-kit/share"
)

func TestLinkValidity(t *testing.T) {
	future := time.Now().Add(time.Hour)
	link := &share.Link{ExpiresAt: future, MaxUses: 2, UseCount: 1}
	if link.IsExpired() || link.IsExhausted() || !link.IsValid() {
		t.Fatal("expected valid link")
	}
	link.UseCount = 2
	if !link.IsExhausted() || link.IsValid() {
		t.Fatal("expected exhausted")
	}
	link = &share.Link{ExpiresAt: time.Now().Add(-time.Hour)}
	if !link.IsExpired() || link.IsValid() {
		t.Fatal("expected expired")
	}
}
