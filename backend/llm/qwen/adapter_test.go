package qwen_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	llm "github.com/EthanShen10086/voxera-kit/llm"
	"github.com/EthanShen10086/voxera-kit/llm/qwen"
)

func TestChatAndHeaders(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer sk-qwen" {
			t.Errorf("auth = %q", r.Header.Get("Authorization"))
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"request_id": "req-1",
			"output":     map[string]string{"text": "你好", "finish_reason": "stop"},
			"usage":      map[string]int{"input_tokens": 4, "output_tokens": 2},
		})
	}))
	defer srv.Close()

	a := qwen.New(llm.Config{Endpoint: srv.URL, APIKey: "sk-qwen"})
	if a.Name() != "qwen" {
		t.Fatalf("Name = %q", a.Name())
	}

	resp, err := a.Chat(context.Background(), llm.Request{
		Messages: []llm.Message{{Role: llm.RoleUser, Content: "hi"}},
	})
	if err != nil || resp.Content != "你好" {
		t.Fatalf("Chat: %+v err=%v", resp, err)
	}
}

func TestEmbedListModelsAvailable(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"request_id": "p",
			"output":     map[string]string{"text": "pong"},
			"usage":      map[string]int{"input_tokens": 1, "output_tokens": 1},
		})
	}))
	defer srv.Close()

	a := qwen.New(llm.Config{Endpoint: srv.URL, APIKey: "k"})
	_, err := a.Embed(context.Background(), llm.EmbeddingRequest{})
	if !errors.Is(err, llm.ErrNotSupported) {
		t.Fatalf("Embed: %v", err)
	}
	models, err := a.ListModels(context.Background())
	if err != nil || len(models) == 0 {
		t.Fatalf("models: %v", models)
	}
	if !a.Available(context.Background()) {
		t.Fatal("expected available")
	}
}
