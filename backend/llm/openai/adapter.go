// Package openai implements the llm.Provider interface for OpenAI models
// (GPT-4o, GPT-4, GPT-3.5-turbo) using the official REST API.
package openai

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	llm "github.com/EthanShen10086/voxera-kit/llm"
)

const (
	defaultEndpoint   = "https://api.openai.com/v1"
	defaultModel      = "gpt-4o"
	defaultMaxTokens  = 4096
	defaultTimeout    = 60 * time.Second
)

// Adapter implements llm.Provider for the OpenAI API.
type Adapter struct {
	apiKey   string
	endpoint string
	model    string
	client   *http.Client
	orgID    string
	headers  map[string]string
}

// New creates a new OpenAI adapter from the given configuration.
func New(cfg llm.Config) *Adapter {
	endpoint := cfg.Endpoint
	if endpoint == "" {
		endpoint = defaultEndpoint
	}
	model := cfg.Model
	if model == "" {
		model = defaultModel
	}
	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = defaultTimeout
	}
	return &Adapter{
		apiKey:   cfg.APIKey,
		endpoint: strings.TrimRight(endpoint, "/"),
		model:    model,
		orgID:    cfg.OrgID,
		headers:  cfg.Headers,
		client:   &http.Client{Timeout: timeout},
	}
}

// Name returns "openai".
func (a *Adapter) Name() string { return "openai" }

// Chat performs a synchronous chat completion.
func (a *Adapter) Chat(ctx context.Context, req llm.Request) (*llm.Response, error) {
	start := time.Now()
	model := req.Model
	if model == "" {
		model = a.model
	}

	body := chatRequest{
		Model:       model,
		Messages:    toMessages(req.Messages),
		MaxTokens:   req.MaxTokens,
		Temperature: req.Temperature,
		TopP:        req.TopP,
		Stream:      false,
		Stop:        req.Stop,
	}

	data, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("openai: marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
		a.endpoint+"/chat/completions", bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("openai: create request: %w", err)
	}
	a.setHeaders(httpReq)

	httpResp, err := a.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("openai: send request: %w", err)
	}
	defer httpResp.Body.Close() //nolint:errcheck

	if httpResp.StatusCode != http.StatusOK {
		return nil, a.readError(httpResp)
	}

	var resp chatResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("openai: decode response: %w", err)
	}

	var content, finishReason string
	if len(resp.Choices) > 0 {
		content = resp.Choices[0].Message.Content
		finishReason = resp.Choices[0].FinishReason
	}

	return &llm.Response{
		ID:           resp.ID,
		Model:        resp.Model,
		Content:      content,
		FinishReason: finishReason,
		Usage: llm.TokenUsage{
			InputTokens:  resp.Usage.PromptTokens,
			OutputTokens: resp.Usage.CompletionTokens,
			TotalTokens:  resp.Usage.TotalTokens,
		},
		Latency: time.Since(start),
	}, nil
}

// ChatStream performs a streaming chat completion using SSE.
func (a *Adapter) ChatStream(ctx context.Context, req llm.Request) (<-chan llm.StreamChunk, error) {
	model := req.Model
	if model == "" {
		model = a.model
	}

	body := chatRequest{
		Model:       model,
		Messages:    toMessages(req.Messages),
		MaxTokens:   req.MaxTokens,
		Temperature: req.Temperature,
		TopP:        req.TopP,
		Stream:      true,
		Stop:        req.Stop,
	}

	data, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("openai: marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
		a.endpoint+"/chat/completions", bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("openai: create request: %w", err)
	}
	a.setHeaders(httpReq)

	httpResp, err := a.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("openai: send request: %w", err)
	}

	if httpResp.StatusCode != http.StatusOK {
		defer httpResp.Body.Close() //nolint:errcheck
		return nil, a.readError(httpResp)
	}

	ch := make(chan llm.StreamChunk)
	go func() {
		defer close(ch)
		defer httpResp.Body.Close() //nolint:errcheck
		a.readSSE(httpResp.Body, ch)
	}()
	return ch, nil
}

// Embed generates vector embeddings for the given texts.
func (a *Adapter) Embed(ctx context.Context, req llm.EmbeddingRequest) (*llm.EmbeddingResponse, error) {
	model := req.Model
	if model == "" {
		model = "text-embedding-3-small"
	}

	body := embeddingRequest{
		Model: model,
		Input: req.Texts,
	}

	data, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("openai: marshal embedding request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
		a.endpoint+"/embeddings", bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("openai: create embedding request: %w", err)
	}
	a.setHeaders(httpReq)

	httpResp, err := a.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("openai: send embedding request: %w", err)
	}
	defer httpResp.Body.Close() //nolint:errcheck

	if httpResp.StatusCode != http.StatusOK {
		return nil, a.readError(httpResp)
	}

	var resp embeddingResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("openai: decode embedding response: %w", err)
	}

	embeddings := make([][]float64, len(resp.Data))
	for i, d := range resp.Data {
		embeddings[i] = d.Embedding
	}

	return &llm.EmbeddingResponse{
		Model:      resp.Model,
		Embeddings: embeddings,
		Usage: llm.TokenUsage{
			InputTokens: resp.Usage.PromptTokens,
			TotalTokens: resp.Usage.TotalTokens,
		},
	}, nil
}

// ListModels returns the models available from OpenAI.
func (a *Adapter) ListModels(_ context.Context) ([]llm.ModelInfo, error) {
	return []llm.ModelInfo{
		{ID: "gpt-4o", Provider: "openai", MaxTokens: 128000, InputPrice: 2.50, OutputPrice: 10.0, SupportsVision: true, SupportsStreaming: true},
		{ID: "gpt-4o-mini", Provider: "openai", MaxTokens: 128000, InputPrice: 0.15, OutputPrice: 0.60, SupportsVision: true, SupportsStreaming: true},
		{ID: "gpt-4-turbo", Provider: "openai", MaxTokens: 128000, InputPrice: 10.0, OutputPrice: 30.0, SupportsVision: true, SupportsStreaming: true},
		{ID: "gpt-4", Provider: "openai", MaxTokens: 8192, InputPrice: 30.0, OutputPrice: 60.0, SupportsStreaming: true},
		{ID: "gpt-3.5-turbo", Provider: "openai", MaxTokens: 16385, InputPrice: 0.50, OutputPrice: 1.50, SupportsStreaming: true},
		{ID: "text-embedding-3-small", Provider: "openai", MaxTokens: 8191, InputPrice: 0.02, SupportsEmbedding: true},
		{ID: "text-embedding-3-large", Provider: "openai", MaxTokens: 8191, InputPrice: 0.13, SupportsEmbedding: true},
	}, nil
}

// Available reports whether the OpenAI API is reachable.
func (a *Adapter) Available(ctx context.Context) bool {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, a.endpoint+"/models", nil)
	if err != nil {
		return false
	}
	a.setHeaders(req)
	resp, err := a.client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close() //nolint:errcheck
	return resp.StatusCode == http.StatusOK
}

func (a *Adapter) setHeaders(req *http.Request) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+a.apiKey)
	if a.orgID != "" {
		req.Header.Set("OpenAI-Organization", a.orgID)
	}
	for k, v := range a.headers {
		req.Header.Set(k, v)
	}
}

func (a *Adapter) readError(resp *http.Response) error {
	body, _ := io.ReadAll(resp.Body) //nolint:errcheck
	return fmt.Errorf("openai: API error %d: %s", resp.StatusCode, string(body))
}

func (a *Adapter) readSSE(r io.Reader, ch chan<- llm.StreamChunk) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		payload := strings.TrimPrefix(line, "data: ")
		if payload == "[DONE]" {
			return
		}
		var event streamEvent
		if err := json.Unmarshal([]byte(payload), &event); err != nil {
			ch <- llm.StreamChunk{Err: fmt.Errorf("openai: decode stream event: %w", err)}
			return
		}
		if len(event.Choices) == 0 {
			continue
		}
		ch <- llm.StreamChunk{
			Content:      event.Choices[0].Delta.Content,
			FinishReason: event.Choices[0].FinishReason,
		}
	}
	if err := scanner.Err(); err != nil {
		ch <- llm.StreamChunk{Err: fmt.Errorf("openai: read stream: %w", err)}
	}
}

// --- Wire types ---

type chatRequest struct {
	Model       string    `json:"model"`
	Messages    []message `json:"messages"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
	Temperature float64   `json:"temperature,omitempty"`
	TopP        float64   `json:"top_p,omitempty"`
	Stream      bool      `json:"stream"`
	Stop        []string  `json:"stop,omitempty"`
}

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatResponse struct {
	ID      string `json:"id"`
	Model   string `json:"model"`
	Choices []struct {
		Message      message `json:"message"`
		FinishReason string  `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

type streamEvent struct {
	Choices []struct {
		Delta struct {
			Content string `json:"content"`
		} `json:"delta"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
}

type embeddingRequest struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
}

type embeddingResponse struct {
	Model string `json:"model"`
	Data  []struct {
		Embedding []float64 `json:"embedding"`
	} `json:"data"`
	Usage struct {
		PromptTokens int `json:"prompt_tokens"`
		TotalTokens  int `json:"total_tokens"`
	} `json:"usage"`
}

func toMessages(msgs []llm.Message) []message {
	out := make([]message, len(msgs))
	for i, m := range msgs {
		out[i] = message{Role: string(m.Role), Content: m.Content}
	}
	return out
}
