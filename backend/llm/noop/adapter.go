// Package noop provides a no-op llm.Provider implementation for testing.
package noop

import (
	"context"

	llm "github.com/EthanShen10086/voxera-kit/llm"
)

// Adapter implements llm.Provider as a no-op for use in tests.
type Adapter struct{}

// New creates a new no-op adapter.
func New() *Adapter { return &Adapter{} }

// Name returns "noop".
func (a *Adapter) Name() string { return "noop" }

// Chat returns an empty response.
func (a *Adapter) Chat(_ context.Context, _ llm.Request) (*llm.Response, error) {
	return &llm.Response{}, nil
}

// ChatStream returns a channel that closes immediately.
func (a *Adapter) ChatStream(_ context.Context, _ llm.Request) (<-chan llm.StreamChunk, error) {
	ch := make(chan llm.StreamChunk)
	close(ch)
	return ch, nil
}

// Embed returns an empty embedding response.
func (a *Adapter) Embed(_ context.Context, _ llm.EmbeddingRequest) (*llm.EmbeddingResponse, error) {
	return &llm.EmbeddingResponse{}, nil
}

// ListModels returns an empty model list.
func (a *Adapter) ListModels(_ context.Context) ([]llm.ModelInfo, error) {
	return nil, nil
}

// Available always returns true.
func (a *Adapter) Available(_ context.Context) bool { return true }
