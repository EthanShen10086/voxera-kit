package deepseek_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	llm "github.com/EthanShen10086/voxera-kit/llm"
	"github.com/EthanShen10086/voxera-kit/llm/deepseek"
)

func TestChatAndEmbed(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, "/embeddings"):
			_ = json.NewEncoder(w).Encode(map[string]any{
				"model": "deepseek-embed",
				"data":  []map[string]any{{"embedding": []float64{0.1, 0.2}}},
				"usage": map[string]int{"prompt_tokens": 3, "total_tokens": 3},
			})
		default:
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id": "c1", "model": "deepseek-chat",
				"choices": []map[string]any{
					{"message": map[string]string{"content": "ok"}, "finish_reason": "stop"},
				},
				"usage": map[string]int{"prompt_tokens": 2, "completion_tokens": 1, "total_tokens": 3},
			})
		}
	}))
	defer srv.Close()

	a := deepseek.New(llm.Config{Endpoint: srv.URL, APIKey: "k"})
	if a.Name() != "deepseek" {
		t.Fatalf("Name = %q", a.Name())
	}

	resp, err := a.Chat(context.Background(), llm.Request{
		Messages: []llm.Message{{Role: llm.RoleUser, Content: "ping"}},
	})
	if err != nil || resp.Content != "ok" {
		t.Fatalf("Chat: %+v err=%v", resp, err)
	}

	emb, err := a.Embed(context.Background(), llm.EmbeddingRequest{Texts: []string{"hello"}})
	if err != nil || len(emb.Embeddings) != 1 {
		t.Fatalf("Embed: %+v err=%v", emb, err)
	}
}

func TestListModelsAndAvailable(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/models") {
			w.WriteHeader(http.StatusOK)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	a := deepseek.New(llm.Config{Endpoint: srv.URL, APIKey: "k"})
	if !a.Available(context.Background()) {
		t.Fatal("expected available")
	}
	models, err := a.ListModels(context.Background())
	if err != nil || len(models) == 0 {
		t.Fatalf("models: %v err=%v", models, err)
	}
}
