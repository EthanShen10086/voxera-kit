// Package deepseek implements the llm.Provider interface for DeepSeek models
// using their OpenAI-compatible REST API.
package deepseek

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
	defaultEndpoint = "https://api.deepseek.com/v1"
	defaultModel    = "deepseek-chat"
	defaultTimeout  = 60 * time.Second
)

// Adapter implements llm.Provider for the DeepSeek API.
type Adapter struct {
	apiKey   string
	endpoint string
	model    string
	client   *http.Client
	headers  map[string]string
}

// New creates a new DeepSeek adapter from the given configuration.
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
		headers:  cfg.Headers,
		client:   &http.Client{Timeout: timeout},
	}
}

// Name returns "deepseek".
func (a *Adapter) Name() string { return "deepseek" }

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
		return nil, fmt.Errorf("deepseek: marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
		a.endpoint+"/chat/completions", bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("deepseek: create request: %w", err)
	}
	a.setHeaders(httpReq)

	httpResp, err := a.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("deepseek: send request: %w", err)
	}
	defer httpResp.Body.Close() //nolint:errcheck

	if httpResp.StatusCode != http.StatusOK {
		return nil, a.readError(httpResp)
	}

	var resp chatResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("deepseek: decode response: %w", err)
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
		return nil, fmt.Errorf("deepseek: marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
		a.endpoint+"/chat/completions", bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("deepseek: create request: %w", err)
	}
	a.setHeaders(httpReq)

	httpResp, err := a.client.Do(httpReq) //nolint:bodyclose // closed in streaming goroutine below
	if err != nil {
		return nil, fmt.Errorf("deepseek: send request: %w", err)
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
		model = "deepseek-chat"
	}

	body := embeddingRequest{
		Model: model,
		Input: req.Texts,
	}

	data, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("deepseek: marshal embedding request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
		a.endpoint+"/embeddings", bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("deepseek: create embedding request: %w", err)
	}
	a.setHeaders(httpReq)

	httpResp, err := a.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("deepseek: send embedding request: %w", err)
	}
	defer httpResp.Body.Close() //nolint:errcheck

	if httpResp.StatusCode != http.StatusOK {
		return nil, a.readError(httpResp)
	}

	var resp embeddingResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("deepseek: decode embedding response: %w", err)
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

// ListModels returns the models available from DeepSeek.
func (a *Adapter) ListModels(_ context.Context) ([]llm.ModelInfo, error) {
	return []llm.ModelInfo{
		{ID: "deepseek-chat", Provider: "deepseek", MaxTokens: 64000, InputPrice: 0.14, OutputPrice: 0.28, SupportsStreaming: true},
		{ID: "deepseek-coder", Provider: "deepseek", MaxTokens: 64000, InputPrice: 0.14, OutputPrice: 0.28, SupportsStreaming: true},
		{ID: "deepseek-reasoner", Provider: "deepseek", MaxTokens: 64000, InputPrice: 0.55, OutputPrice: 2.19, SupportsStreaming: true},
	}, nil
}

// Available reports whether the DeepSeek API is reachable.
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
	for k, v := range a.headers {
		req.Header.Set(k, v)
	}
}

func (a *Adapter) readError(resp *http.Response) error {
	body, _ := io.ReadAll(resp.Body) //nolint:errcheck
	return fmt.Errorf("deepseek: API error %d: %s", resp.StatusCode, string(body))
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
			ch <- llm.StreamChunk{Err: fmt.Errorf("deepseek: decode stream event: %w", err)}
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
		ch <- llm.StreamChunk{Err: fmt.Errorf("deepseek: read stream: %w", err)}
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
