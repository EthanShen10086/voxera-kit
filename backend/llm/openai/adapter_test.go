package openai_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	llm "github.com/EthanShen10086/voxera-kit/llm"
	"github.com/EthanShen10086/voxera-kit/llm/openai"
)

func TestChatAndName(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.Header.Get("Authorization"), "Bearer sk-test") {
			t.Errorf("auth = %q", r.Header.Get("Authorization"))
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id": "cmpl-1", "model": "gpt-4o",
			"choices": []map[string]any{
				{"message": map[string]string{"content": "pong"}, "finish_reason": "stop"},
			},
			"usage": map[string]int{"prompt_tokens": 5, "completion_tokens": 3, "total_tokens": 8},
		})
	}))
	defer srv.Close()

	a := openai.New(llm.Config{Endpoint: srv.URL, APIKey: "sk-test"})
	if a.Name() != "openai" {
		t.Fatalf("Name = %q", a.Name())
	}

	resp, err := a.Chat(context.Background(), llm.Request{
		Messages: []llm.Message{{Role: llm.RoleUser, Content: "ping"}},
	})
	if err != nil || resp.Content != "pong" || resp.Usage.TotalTokens != 8 {
		t.Fatalf("Chat: %+v err=%v", resp, err)
	}
}

func TestAvailableAndListModels(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, "/models"):
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": []map[string]string{{"id": "gpt-4o"}},
			})
		default:
			_ = json.NewEncoder(w).Encode(map[string]any{
				"choices": []map[string]any{{"message": map[string]string{"content": "ok"}}},
				"usage":   map[string]int{"prompt_tokens": 1, "completion_tokens": 1, "total_tokens": 2},
			})
		}
	}))
	defer srv.Close()

	a := openai.New(llm.Config{Endpoint: srv.URL, APIKey: "k"})
	if !a.Available(context.Background()) {
		t.Fatal("expected available")
	}
	models, err := a.ListModels(context.Background())
	if err != nil || len(models) == 0 || models[0].ID != "gpt-4o" {
		t.Fatalf("models: %+v err=%v", models, err)
	}
}

func TestChat_APIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":{"message":"invalid key"}}`))
	}))
	defer srv.Close()

	a := openai.New(llm.Config{Endpoint: srv.URL, APIKey: "bad"})
	_, err := a.Chat(context.Background(), llm.Request{
		Messages: []llm.Message{{Role: llm.RoleUser, Content: "x"}},
	})
	if err == nil {
		t.Fatal("expected error")
	}
}
