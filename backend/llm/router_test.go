package llm_test

import (
	"context"
	"errors"
	"testing"

	llm "github.com/EthanShen10086/voxera-kit/llm"
	"github.com/EthanShen10086/voxera-kit/llm/noop"
)

type stubProvider struct {
	name      string
	available bool
	models    []llm.ModelInfo
	chatErr   error
	content   string
}

func (s stubProvider) Name() string { return s.name }
func (s stubProvider) Chat(_ context.Context, _ llm.Request) (*llm.Response, error) {
	if s.chatErr != nil {
		return nil, s.chatErr
	}
	return &llm.Response{Content: s.content}, nil
}
func (s stubProvider) ChatStream(context.Context, llm.Request) (<-chan llm.StreamChunk, error) {
	ch := make(chan llm.StreamChunk)
	close(ch)
	return ch, nil
}
func (s stubProvider) Embed(context.Context, llm.EmbeddingRequest) (*llm.EmbeddingResponse, error) {
	return &llm.EmbeddingResponse{}, nil
}
func (s stubProvider) ListModels(context.Context) ([]llm.ModelInfo, error) {
	return s.models, nil
}
func (s stubProvider) Available(context.Context) bool { return s.available }

func TestRouterRouteAndFallback(t *testing.T) {
	r := llm.NewRouter()
	fail := stubProvider{name: "fail", available: true, chatErr: errors.New("down")}
	ok := stubProvider{name: "ok", available: true, content: "pong", models: []llm.ModelInfo{{ID: "gpt-4o"}}}
	r.Register(fail, 0)
	r.Register(ok, 1)

	resp, err := r.Route(context.Background(), llm.Request{
		Model:    "gpt-4o",
		Messages: []llm.Message{{Role: llm.RoleUser, Content: "hi"}},
	})
	if err != nil || resp.Content != "pong" {
		t.Fatalf("Route: %+v err=%v", resp, err)
	}

	ch, err := r.RouteStream(context.Background(), llm.Request{})
	if err != nil || ch == nil {
		t.Fatalf("RouteStream: %v err=%v", ch, err)
	}

	models := r.ListAllModels(context.Background())
	if len(models) == 0 {
		t.Fatal("expected models")
	}
}

func TestRouterNoProviders(t *testing.T) {
	r := llm.NewRouter()
	_, err := r.Route(context.Background(), llm.Request{})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestRouterUsesNoop(t *testing.T) {
	r := llm.NewRouter()
	r.Register(noop.New(), 0)
	resp, err := r.Route(context.Background(), llm.Request{
		Messages: []llm.Message{{Role: llm.RoleUser, Content: "x"}},
	})
	if err != nil || resp == nil {
		t.Fatalf("Route noop: %+v err=%v", resp, err)
	}
}

func TestEstimateTokensAndCost(t *testing.T) {
	if got := llm.EstimateTokens("hello world"); got < 1 {
		t.Fatalf("latin tokens = %d", got)
	}
	if got := llm.EstimateTokens("你好世界"); got < 1 {
		t.Fatalf("cjk tokens = %d", got)
	}
	if cost := llm.EstimateCost("gpt-4o", 1_000_000, 500_000); cost <= 0 {
		t.Fatalf("cost = %d", cost)
	}
	if cost := llm.EstimateCost("unknown-model", 1000, 500); cost != 0 {
		t.Fatalf("unknown cost = %d", cost)
	}
}
