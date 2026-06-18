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

func TestChatStream(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-DashScope-SSE") != "enable" {
			t.Errorf("X-DashScope-SSE = %q", r.Header.Get("X-DashScope-SSE"))
		}
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = w.Write([]byte("data: {\"output\":{\"text\":\"流\",\"finish_reason\":\"\"}}\n\n"))
		_, _ = w.Write([]byte("data: {\"output\":{\"text\":\"式\",\"finish_reason\":\"stop\"}}\n\n"))
	}))
	defer srv.Close()

	a := qwen.New(llm.Config{Endpoint: srv.URL, APIKey: "sk-qwen"})
	ch, err := a.ChatStream(context.Background(), llm.Request{
		Messages: []llm.Message{{Role: llm.RoleUser, Content: "hi"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	var content string
	for chunk := range ch {
		if chunk.Err != nil {
			t.Fatal(chunk.Err)
		}
		content += chunk.Content
	}
	if content != "流式" {
		t.Fatalf("content = %q", content)
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
