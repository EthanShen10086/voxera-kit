package memory_test

import (
	"context"
	"testing"
	"time"

	"github.com/EthanShen10086/voxera-kit/share"
	"github.com/EthanShen10086/voxera-kit/share/memory"
)

func TestRepositoryCRUD(t *testing.T) {
	ctx := context.Background()
	repo := memory.NewRepository()
	link := &share.Link{
		ID:           "id1",
		Token:        "tok1",
		ResourceType: "doc",
		ResourceID:   "doc-1",
		CreatedBy:    "user1",
		CreatedAt:    time.Now(),
	}

	if err := repo.Save(ctx, link); err != nil {
		t.Fatal(err)
	}
	got, err := repo.FindByToken(ctx, "tok1")
	if err != nil || got.ID != "id1" {
		t.Fatalf("FindByToken: %+v err=%v", got, err)
	}
	got, err = repo.FindByID(ctx, "id1")
	if err != nil || got.Token != "tok1" {
		t.Fatalf("FindByID: %+v err=%v", got, err)
	}
	if err := repo.IncrementUseCount(ctx, "id1"); err != nil {
		t.Fatal(err)
	}
	got, _ = repo.FindByID(ctx, "id1")
	if got.UseCount != 1 {
		t.Fatalf("UseCount = %d", got.UseCount)
	}
	if err := repo.Revoke(ctx, "id1"); err != nil {
		t.Fatal(err)
	}
	got, _ = repo.FindByID(ctx, "id1")
	if !got.ExpiresAt.Before(time.Now()) {
		t.Fatal("expected revoked link to be expired")
	}
}

func TestRepositoryListAndErrors(t *testing.T) {
	ctx := context.Background()
	repo := memory.NewRepository()
	link := &share.Link{ID: "1", Token: "t", CreatedBy: "u", CreatedAt: time.Now()}
	_ = repo.Save(ctx, link)
	list, err := repo.ListByUser(ctx, "u", 10, 0)
	if err != nil || len(list) != 1 {
		t.Fatalf("ListByUser: %v err=%v", list, err)
	}
	if _, err := repo.FindByID(ctx, "missing"); err == nil {
		t.Fatal("expected not found")
	}
}

func TestGenerator(t *testing.T) {
	ctx := context.Background()
	gen := memory.NewGenerator()
	link, err := gen.Generate(ctx, share.CreateShareRequest{
		ResourceType: "file",
		ResourceID:   "f1",
		CreatedBy:    "u1",
		Permissions:  []share.Permission{share.PermissionView},
	})
	if err != nil || link.ID == "" || link.Token == "" {
		t.Fatalf("Generate: %+v err=%v", link, err)
	}
}
