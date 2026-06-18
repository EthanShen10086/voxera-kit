package memory_test

import (
	"context"
	"testing"
	"time"

	"github.com/EthanShen10086/voxera-kit/shorturl"
	"github.com/EthanShen10086/voxera-kit/shorturl/memory"
)

func TestGenerateResolveAndDelete(t *testing.T) {
	ctx := context.Background()
	a := memory.New(shorturl.Config{AllowCustomCode: true})

	su, err := a.Generate(ctx, "https://example.com/long", shorturl.WithCustomCode("abc123"), shorturl.WithCreator("u1"))
	if err != nil {
		t.Fatal(err)
	}
	if su.Code != "abc123" || su.OriginalURL != "https://example.com/long" {
		t.Fatalf("su = %+v", su)
	}

	got, err := a.Resolve(ctx, "abc123")
	if err != nil || got.OriginalURL != "https://example.com/long" {
		t.Fatalf("Resolve: %+v err=%v", got, err)
	}
	if err := a.IncrementClick(ctx, "abc123"); err != nil {
		t.Fatal(err)
	}
	got, _ = a.Resolve(ctx, "abc123")
	if got.ClickCount != 1 {
		t.Fatalf("ClickCount = %d", got.ClickCount)
	}
	if err := a.Delete(ctx, "abc123"); err != nil {
		t.Fatal(err)
	}
	if _, err := a.Resolve(ctx, "abc123"); err == nil {
		t.Fatal("expected not found after delete")
	}
}

func TestGenerateExpiryAndList(t *testing.T) {
	ctx := context.Background()
	a := memory.New(shorturl.Config{AllowCustomCode: true})

	su, err := a.Generate(ctx, "https://expired.test", shorturl.WithCustomCode("expired"),
		shorturl.WithExpiry(time.Millisecond), shorturl.WithCreator("creator"))
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(5 * time.Millisecond)
	if _, err := a.Resolve(ctx, su.Code); err == nil {
		t.Fatal("expected expired error")
	}

	list, err := a.ListByCreator(ctx, "creator", 0, 10)
	if err != nil || len(list) != 1 {
		t.Fatalf("ListByCreator: %v err=%v", list, err)
	}
}

func TestGenerateErrors(t *testing.T) {
	ctx := context.Background()
	a := memory.New(shorturl.Config{AllowCustomCode: false})
	if _, err := a.Generate(ctx, "https://x", shorturl.WithCustomCode("nope")); err == nil {
		t.Fatal("expected custom code disallowed")
	}
	a2 := memory.New(shorturl.Config{AllowCustomCode: true})
	_, _ = a2.Generate(ctx, "https://x", shorturl.WithCustomCode("dup"))
	if _, err := a2.Generate(ctx, "https://y", shorturl.WithCustomCode("dup")); err == nil {
		t.Fatal("expected duplicate code error")
	}
}
