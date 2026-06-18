package claude_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	llm "github.com/EthanShen10086/voxera-kit/llm"
	"github.com/EthanShen10086/voxera-kit/llm/claude"
)

func TestChatAndName(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("x-api-key") != "key" {
			t.Errorf("x-api-key = %q", r.Header.Get("x-api-key"))
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id": "msg_1", "model": "claude-sonnet-4-20250514", "stop_reason": "end_turn",
			"content": []map[string]string{{"type": "text", "text": "hello"}},
			"usage":   map[string]int{"input_tokens": 10, "output_tokens": 5},
		})
	}))
	defer srv.Close()

	a := claude.New(llm.Config{Endpoint: srv.URL, APIKey: "key"})
	if a.Name() != "claude" {
		t.Fatalf("Name = %q", a.Name())
	}

	resp, err := a.Chat(context.Background(), llm.Request{
		Messages: []llm.Message{
			{Role: llm.RoleSystem, Content: "be helpful"},
			{Role: llm.RoleUser, Content: "hi"},
		},
	})
	if err != nil || resp.Content != "hello" {
		t.Fatalf("Chat: %+v err=%v", resp, err)
	}
}

func TestEmbedListModelsAvailable(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id": "msg_ping", "model": "claude-sonnet-4-20250514", "stop_reason": "end_turn",
			"content": []map[string]string{{"type": "text", "text": "pong"}},
			"usage":   map[string]int{"input_tokens": 1, "output_tokens": 1},
		})
	}))
	defer srv.Close()

	a := claude.New(llm.Config{Endpoint: srv.URL, APIKey: "k"})
	_, err := a.Embed(context.Background(), llm.EmbeddingRequest{})
	if !errors.Is(err, llm.ErrNotSupported) {
		t.Fatalf("Embed: %v", err)
	}
	models, err := a.ListModels(context.Background())
	if err != nil || len(models) == 0 {
		t.Fatalf("models: %v err=%v", models, err)
	}
	if !a.Available(context.Background()) {
		t.Fatal("expected available")
	}
}

func TestChatStream(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = w.Write([]byte("event: message_start\n"))
		_, _ = w.Write([]byte("data: {\"type\":\"content_block_delta\",\"delta\":{\"type\":\"text_delta\",\"text\":\"hi\"}}\n\n"))
		_, _ = w.Write([]byte("event: message_stop\n"))
		_, _ = w.Write([]byte("data: {\"type\":\"message_delta\",\"delta\":{\"stop_reason\":\"end_turn\"}}\n\n"))
	}))
	defer srv.Close()

	a := claude.New(llm.Config{Endpoint: srv.URL, APIKey: "key"})
	ch, err := a.ChatStream(context.Background(), llm.Request{
		Messages: []llm.Message{{Role: llm.RoleUser, Content: "ping"}},
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
	if content != "hi" {
		t.Fatalf("content = %q", content)
	}
}

func TestChat_APIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer srv.Close()

	a := claude.New(llm.Config{Endpoint: srv.URL, APIKey: "bad"})
	_, err := a.Chat(context.Background(), llm.Request{
		Messages: []llm.Message{{Role: llm.RoleUser, Content: "x"}},
	})
	if err == nil || !strings.Contains(err.Error(), "401") {
		t.Fatalf("err = %v", err)
	}
}
