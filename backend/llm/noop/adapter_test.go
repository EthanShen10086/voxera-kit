package noop

import (
	"context"
	"testing"

	llm "github.com/EthanShen10086/voxera-kit/llm"
)

func TestNoopProviderContract(t *testing.T) {
	p := New()
	if p.Name() != "noop" {
		t.Fatal("name")
	}
	if !p.Available(context.Background()) {
		t.Fatal("available")
	}
	resp, err := p.Chat(context.Background(), llm.Request{Model: "test"})
	if err != nil || resp == nil {
		t.Fatalf("chat: %v", err)
	}
	ch, err := p.ChatStream(context.Background(), llm.Request{})
	if err != nil {
		t.Fatal(err)
	}
	if _, open := <-ch; open {
		t.Fatal("stream should close immediately")
	}

	emb, err := p.Embed(context.Background(), llm.EmbeddingRequest{})
	if err != nil || emb == nil {
		t.Fatalf("Embed: %v", err)
	}
	models, err := p.ListModels(context.Background())
	if err != nil {
		t.Fatalf("ListModels: %v", err)
	}
	if models != nil {
		t.Fatalf("expected nil models, got %#v", models)
	}
}
