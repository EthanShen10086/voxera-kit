// Package hunyuan implements the llm.Provider interface for Tencent's Hunyuan
// models using the Hunyuan REST API.
package hunyuan

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
	defaultEndpoint = "https://hunyuan.tencentcloudapi.com"
	defaultModel    = "hunyuan-pro"
	defaultTimeout  = 60 * time.Second
)

// Adapter implements llm.Provider for the Tencent Hunyuan API.
//
// TODO: Implement full TC3-HMAC-SHA256 signature authentication.
// Currently uses a simplified API-key-in-header approach.
type Adapter struct {
	apiKey   string
	endpoint string
	model    string
	client   *http.Client
	headers  map[string]string
}

// New creates a new Hunyuan adapter from the given configuration.
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

// Name returns "hunyuan".
func (a *Adapter) Name() string { return "hunyuan" }

// Chat performs a synchronous chat completion.
func (a *Adapter) Chat(ctx context.Context, req llm.Request) (*llm.Response, error) {
	start := time.Now()
	model := req.Model
	if model == "" {
		model = a.model
	}

	body := hunyuanRequest{
		Model:       model,
		Messages:    toMessages(req.Messages),
		TopP:        req.TopP,
		Temperature: req.Temperature,
		Stream:      false,
	}

	data, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("hunyuan: marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
		a.endpoint+"/hyllm/v1/chat/completions", bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("hunyuan: create request: %w", err)
	}
	a.setHeaders(httpReq)

	httpResp, err := a.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("hunyuan: send request: %w", err)
	}
	defer func() { _ = httpResp.Body.Close() }()

	if httpResp.StatusCode != http.StatusOK {
		return nil, a.readError(httpResp)
	}

	var resp hunyuanResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("hunyuan: decode response: %w", err)
	}

	var content, finishReason string
	if len(resp.Choices) > 0 {
		content = resp.Choices[0].Message.Content
		finishReason = resp.Choices[0].FinishReason
	}

	return &llm.Response{
		ID:           resp.ID,
		Model:        model,
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

	body := hunyuanRequest{
		Model:       model,
		Messages:    toMessages(req.Messages),
		TopP:        req.TopP,
		Temperature: req.Temperature,
		Stream:      true,
	}

	data, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("hunyuan: marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
		a.endpoint+"/hyllm/v1/chat/completions", bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("hunyuan: create request: %w", err)
	}
	a.setHeaders(httpReq)

	httpResp, err := a.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("hunyuan: send request: %w", err)
	}

	if httpResp.StatusCode != http.StatusOK {
		defer func() { _ = httpResp.Body.Close() }()
		return nil, a.readError(httpResp)
	}

	respBody := httpResp.Body
	httpResp.Body = http.NoBody
	ch := make(chan llm.StreamChunk)
	go func() {
		defer close(ch)
		defer func() { _ = respBody.Close() }()
		a.readSSE(respBody, ch)
	}()
	return ch, nil
}

// Embed returns ErrNotSupported; Hunyuan embedding is not yet implemented.
func (a *Adapter) Embed(_ context.Context, _ llm.EmbeddingRequest) (*llm.EmbeddingResponse, error) {
	return nil, llm.ErrNotSupported
}

// ListModels returns the models available from Hunyuan.
func (a *Adapter) ListModels(_ context.Context) ([]llm.ModelInfo, error) {
	return []llm.ModelInfo{
		{ID: "hunyuan-pro", Provider: "hunyuan", MaxTokens: 32000, InputPrice: 3.00, OutputPrice: 9.00, SupportsStreaming: true},
		{ID: "hunyuan-standard", Provider: "hunyuan", MaxTokens: 32000, InputPrice: 0.45, OutputPrice: 0.80, SupportsStreaming: true},
		{ID: "hunyuan-lite", Provider: "hunyuan", MaxTokens: 16000, InputPrice: 0.0, OutputPrice: 0.0, SupportsStreaming: true},
	}, nil
}

// Available reports whether the Hunyuan API is reachable.
func (a *Adapter) Available(ctx context.Context) bool {
	body := hunyuanRequest{
		Model:    a.model,
		Messages: []message{{Role: "user", Content: "ping"}},
		Stream:   false,
	}
	data, err := json.Marshal(body)
	if err != nil {
		return false
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		a.endpoint+"/hyllm/v1/chat/completions", bytes.NewReader(data))
	if err != nil {
		return false
	}
	a.setHeaders(req)

	resp, err := a.client.Do(req)
	if err != nil {
		return false
	}
	defer func() { _ = resp.Body.Close() }()
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
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("hunyuan: API error %d: read body: %w", resp.StatusCode, err)
	}
	return fmt.Errorf("hunyuan: API error %d: %s", resp.StatusCode, string(body))
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
			ch <- llm.StreamChunk{Err: fmt.Errorf("hunyuan: decode stream event: %w", err)}
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
		ch <- llm.StreamChunk{Err: fmt.Errorf("hunyuan: read stream: %w", err)}
	}
}

// --- Wire types ---

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type hunyuanRequest struct {
	Model       string    `json:"model"`
	Messages    []message `json:"messages"`
	TopP        float64   `json:"top_p,omitempty"`
	Temperature float64   `json:"temperature,omitempty"`
	Stream      bool      `json:"stream"`
}

type hunyuanResponse struct {
	ID      string `json:"id"`
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

func toMessages(msgs []llm.Message) []message {
	out := make([]message, len(msgs))
	for i, m := range msgs {
		out[i] = message{Role: string(m.Role), Content: m.Content}
	}
	return out
}
