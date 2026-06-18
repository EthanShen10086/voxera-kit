package hunyuan_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	llm "github.com/EthanShen10086/voxera-kit/llm"
	"github.com/EthanShen10086/voxera-kit/llm/hunyuan"
)

func TestChatWithAPIKey(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") == "" {
			t.Error("missing Authorization")
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id": "h1",
			"choices": []map[string]any{
				{"message": map[string]string{"content": "reply"}, "finish_reason": "stop"},
			},
			"usage": map[string]int{"prompt_tokens": 2, "completion_tokens": 3, "total_tokens": 5},
		})
	}))
	defer srv.Close()

	a := hunyuan.New(llm.Config{Endpoint: srv.URL, APIKey: "hk-test"})
	if a.Name() != "hunyuan" {
		t.Fatalf("Name = %q", a.Name())
	}
	resp, err := a.Chat(context.Background(), llm.Request{
		Messages: []llm.Message{{Role: llm.RoleUser, Content: "hi"}},
	})
	if err != nil || resp.Content != "reply" || resp.Usage.TotalTokens != 5 {
		t.Fatalf("Chat: %+v err=%v", resp, err)
	}
}

func TestEmbedNotSupported(t *testing.T) {
	a := hunyuan.New(llm.Config{APIKey: "k"})
	_, err := a.Embed(context.Background(), llm.EmbeddingRequest{Texts: []string{"x"}})
	if err == nil {
		t.Fatal("expected embed error")
	}
}

func TestAvailable(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()
	a := hunyuan.New(llm.Config{Endpoint: srv.URL, APIKey: "k"})
	if !a.Available(context.Background()) {
		t.Fatal("expected available")
	}
}

func TestListModels(t *testing.T) {
	a := hunyuan.New(llm.Config{APIKey: "k"})
	models, err := a.ListModels(context.Background())
	if err != nil || len(models) == 0 {
		t.Fatalf("models: %v err=%v", models, err)
	}
}
