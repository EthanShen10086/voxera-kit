// Package qwen implements the llm.Provider interface for Alibaba's Qwen models
// using the DashScope REST API.
package qwen

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
	defaultEndpoint = "https://dashscope.aliyuncs.com/api/v1/services/aigc/text-generation/generation"
	defaultModel    = "qwen-turbo"
	defaultTimeout  = 60 * time.Second
)

// Adapter implements llm.Provider for the Alibaba DashScope (Qwen) API.
type Adapter struct {
	apiKey   string
	endpoint string
	model    string
	client   *http.Client
	headers  map[string]string
}

// New creates a new Qwen adapter from the given configuration.
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

// Name returns "qwen".
func (a *Adapter) Name() string { return "qwen" }

// Chat performs a synchronous chat completion.
func (a *Adapter) Chat(ctx context.Context, req llm.Request) (*llm.Response, error) {
	start := time.Now()
	model := req.Model
	if model == "" {
		model = a.model
	}

	body := dashScopeRequest{
		Model: model,
		Input: dashScopeInput{
			Messages: toMessages(req.Messages),
		},
		Parameters: dashScopeParams{
			MaxTokens:   req.MaxTokens,
			Temperature: req.Temperature,
			TopP:        req.TopP,
			Stop:        req.Stop,
		},
	}

	data, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("qwen: marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, a.endpoint, bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("qwen: create request: %w", err)
	}
	a.setHeaders(httpReq, false)

	httpResp, err := a.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("qwen: send request: %w", err)
	}
	defer httpResp.Body.Close() //nolint:errcheck

	if httpResp.StatusCode != http.StatusOK {
		return nil, a.readError(httpResp)
	}

	var resp dashScopeResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("qwen: decode response: %w", err)
	}

	return &llm.Response{
		ID:           resp.RequestID,
		Model:        model,
		Content:      resp.Output.Text,
		FinishReason: resp.Output.FinishReason,
		Usage: llm.TokenUsage{
			InputTokens:  resp.Usage.InputTokens,
			OutputTokens: resp.Usage.OutputTokens,
			TotalTokens:  resp.Usage.InputTokens + resp.Usage.OutputTokens,
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

	body := dashScopeRequest{
		Model: model,
		Input: dashScopeInput{
			Messages: toMessages(req.Messages),
		},
		Parameters: dashScopeParams{
			MaxTokens:         req.MaxTokens,
			Temperature:       req.Temperature,
			TopP:              req.TopP,
			Stop:              req.Stop,
			IncrementalOutput: true,
		},
	}

	data, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("qwen: marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, a.endpoint, bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("qwen: create request: %w", err)
	}
	a.setHeaders(httpReq, true)

	httpResp, err := a.client.Do(httpReq) //nolint:bodyclose // closed in streaming goroutine below
	if err != nil {
		return nil, fmt.Errorf("qwen: send request: %w", err)
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

// Embed returns ErrNotSupported; use a dedicated embedding endpoint instead.
func (a *Adapter) Embed(_ context.Context, _ llm.EmbeddingRequest) (*llm.EmbeddingResponse, error) {
	return nil, llm.ErrNotSupported
}

// ListModels returns the models available from Qwen.
func (a *Adapter) ListModels(_ context.Context) ([]llm.ModelInfo, error) {
	return []llm.ModelInfo{
		{ID: "qwen-turbo", Provider: "qwen", MaxTokens: 8000, InputPrice: 0.30, OutputPrice: 0.60, SupportsStreaming: true},
		{ID: "qwen-plus", Provider: "qwen", MaxTokens: 32000, InputPrice: 0.80, OutputPrice: 2.00, SupportsStreaming: true},
		{ID: "qwen-max", Provider: "qwen", MaxTokens: 32000, InputPrice: 2.00, OutputPrice: 6.00, SupportsStreaming: true},
	}, nil
}

// Available reports whether the DashScope API is reachable.
func (a *Adapter) Available(ctx context.Context) bool {
	body := dashScopeRequest{
		Model: a.model,
		Input: dashScopeInput{
			Messages: []message{{Role: "user", Content: "ping"}},
		},
		Parameters: dashScopeParams{MaxTokens: 1},
	}
	data, err := json.Marshal(body)
	if err != nil {
		return false
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.endpoint, bytes.NewReader(data))
	if err != nil {
		return false
	}
	a.setHeaders(req, false)

	resp, err := a.client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close() //nolint:errcheck
	return resp.StatusCode == http.StatusOK
}

func (a *Adapter) setHeaders(req *http.Request, stream bool) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+a.apiKey)
	if stream {
		req.Header.Set("X-DashScope-SSE", "enable")
	}
	for k, v := range a.headers {
		req.Header.Set(k, v)
	}
}

func (a *Adapter) readError(resp *http.Response) error {
	body, _ := io.ReadAll(resp.Body) //nolint:errcheck
	return fmt.Errorf("qwen: API error %d: %s", resp.StatusCode, string(body))
}

func (a *Adapter) readSSE(r io.Reader, ch chan<- llm.StreamChunk) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data:") {
			continue
		}
		payload := strings.TrimPrefix(line, "data:")
		payload = strings.TrimSpace(payload)
		if payload == "" {
			continue
		}
		var event dashScopeResponse
		if err := json.Unmarshal([]byte(payload), &event); err != nil {
			ch <- llm.StreamChunk{Err: fmt.Errorf("qwen: decode stream event: %w", err)}
			return
		}
		ch <- llm.StreamChunk{
			Content:      event.Output.Text,
			FinishReason: event.Output.FinishReason,
		}
	}
	if err := scanner.Err(); err != nil {
		ch <- llm.StreamChunk{Err: fmt.Errorf("qwen: read stream: %w", err)}
	}
}

// --- Wire types ---

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type dashScopeInput struct {
	Messages []message `json:"messages"`
}

type dashScopeParams struct {
	MaxTokens         int      `json:"max_tokens,omitempty"`
	Temperature       float64  `json:"temperature,omitempty"`
	TopP              float64  `json:"top_p,omitempty"`
	Stop              []string `json:"stop,omitempty"`
	IncrementalOutput bool     `json:"incremental_output,omitempty"`
}

type dashScopeRequest struct {
	Model      string          `json:"model"`
	Input      dashScopeInput  `json:"input"`
	Parameters dashScopeParams `json:"parameters"`
}

type dashScopeResponse struct {
	RequestID string `json:"request_id"`
	Output    struct {
		Text         string `json:"text"`
		FinishReason string `json:"finish_reason"`
	} `json:"output"`
	Usage struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

func toMessages(msgs []llm.Message) []message {
	out := make([]message, len(msgs))
	for i, m := range msgs {
		out[i] = message{Role: string(m.Role), Content: m.Content}
	}
	return out
}
