// Package claude implements the llm.Provider interface for Anthropic's Claude
// models using the Messages API.
package claude

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
	defaultEndpoint = "https://api.anthropic.com/v1"
	defaultModel    = "claude-sonnet-4-20250514"
	defaultVersion  = "2023-06-01"
	defaultTimeout  = 120 * time.Second
)

// Adapter implements llm.Provider for the Anthropic Claude API.
type Adapter struct {
	apiKey   string
	endpoint string
	model    string
	version  string
	client   *http.Client
	headers  map[string]string
}

// New creates a new Claude adapter from the given configuration.
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
		version:  defaultVersion,
		headers:  cfg.Headers,
		client:   &http.Client{Timeout: timeout},
	}
}

// Name returns "claude".
func (a *Adapter) Name() string { return "claude" }

// Chat performs a synchronous chat completion.
func (a *Adapter) Chat(ctx context.Context, req llm.Request) (*llm.Response, error) {
	start := time.Now()
	model := req.Model
	if model == "" {
		model = a.model
	}

	system, msgs := splitSystem(req.Messages)
	maxTokens := req.MaxTokens
	if maxTokens == 0 {
		maxTokens = 4096
	}

	body := messagesRequest{
		Model:       model,
		MaxTokens:   maxTokens,
		Messages:    toContentBlocks(msgs),
		Temperature: req.Temperature,
		TopP:        req.TopP,
		Stop:        req.Stop,
		Stream:      false,
	}
	if system != "" {
		body.System = system
	}

	data, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("claude: marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
		a.endpoint+"/messages", bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("claude: create request: %w", err)
	}
	a.setHeaders(httpReq)

	httpResp, err := a.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("claude: send request: %w", err)
	}
	defer func() { _ = httpResp.Body.Close() }()

	if httpResp.StatusCode != http.StatusOK {
		return nil, a.readError(httpResp)
	}

	var resp messagesResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("claude: decode response: %w", err)
	}

	var content string
	for _, block := range resp.Content {
		if block.Type == "text" {
			content += block.Text
		}
	}

	return &llm.Response{
		ID:           resp.ID,
		Model:        resp.Model,
		Content:      content,
		FinishReason: resp.StopReason,
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

	system, msgs := splitSystem(req.Messages)
	maxTokens := req.MaxTokens
	if maxTokens == 0 {
		maxTokens = 4096
	}

	body := messagesRequest{
		Model:       model,
		MaxTokens:   maxTokens,
		Messages:    toContentBlocks(msgs),
		Temperature: req.Temperature,
		TopP:        req.TopP,
		Stop:        req.Stop,
		Stream:      true,
	}
	if system != "" {
		body.System = system
	}

	data, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("claude: marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
		a.endpoint+"/messages", bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("claude: create request: %w", err)
	}
	a.setHeaders(httpReq)

	httpResp, err := a.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("claude: send request: %w", err)
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

// Embed returns ErrNotSupported; Anthropic does not offer an embedding API.
func (a *Adapter) Embed(_ context.Context, _ llm.EmbeddingRequest) (*llm.EmbeddingResponse, error) {
	return nil, llm.ErrNotSupported
}

// ListModels returns the models available from Anthropic.
func (a *Adapter) ListModels(_ context.Context) ([]llm.ModelInfo, error) {
	return []llm.ModelInfo{
		{ID: "claude-sonnet-4-20250514", Provider: "claude", MaxTokens: 200000, InputPrice: 3.00, OutputPrice: 15.00, SupportsVision: true, SupportsStreaming: true},
		{ID: "claude-3-5-haiku-20241022", Provider: "claude", MaxTokens: 200000, InputPrice: 0.80, OutputPrice: 4.00, SupportsStreaming: true},
	}, nil
}

// Available reports whether the Anthropic API is reachable.
func (a *Adapter) Available(ctx context.Context) bool {
	body := messagesRequest{
		Model:     a.model,
		MaxTokens: 1,
		Messages: []contentMessage{
			{Role: "user", Content: []contentBlock{{Type: "text", Text: "ping"}}},
		},
	}
	data, err := json.Marshal(body)
	if err != nil {
		return false
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		a.endpoint+"/messages", bytes.NewReader(data))
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
	req.Header.Set("x-api-key", a.apiKey)
	req.Header.Set("anthropic-version", a.version)
	for k, v := range a.headers {
		req.Header.Set(k, v)
	}
}

func (a *Adapter) readError(resp *http.Response) error {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("claude: API error %d: read body: %w", resp.StatusCode, err)
	}
	return fmt.Errorf("claude: API error %d: %s", resp.StatusCode, string(body))
}

func (a *Adapter) readSSE(r io.Reader, ch chan<- llm.StreamChunk) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "event: ") {
			eventType := strings.TrimPrefix(line, "event: ")
			if eventType == "message_stop" {
				return
			}
			continue
		}

		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		payload := strings.TrimPrefix(line, "data: ")

		var raw json.RawMessage
		if err := json.Unmarshal([]byte(payload), &raw); err != nil {
			ch <- llm.StreamChunk{Err: fmt.Errorf("claude: decode stream event: %w", err)}
			return
		}

		var typed struct {
			Type  string `json:"type"`
			Delta struct {
				Type       string `json:"type"`
				Text       string `json:"text"`
				StopReason string `json:"stop_reason"`
			} `json:"delta"`
		}
		if err := json.Unmarshal(raw, &typed); err != nil {
			continue
		}

		switch typed.Type {
		case "content_block_delta":
			ch <- llm.StreamChunk{Content: typed.Delta.Text}
		case "message_delta":
			ch <- llm.StreamChunk{FinishReason: typed.Delta.StopReason}
		}
	}
	if err := scanner.Err(); err != nil {
		ch <- llm.StreamChunk{Err: fmt.Errorf("claude: read stream: %w", err)}
	}
}

// splitSystem extracts the system message from the list, returning the system
// text and the remaining non-system messages.
func splitSystem(msgs []llm.Message) (string, []llm.Message) {
	var system string
	var rest []llm.Message
	for _, m := range msgs {
		if m.Role == llm.RoleSystem {
			system = m.Content
		} else {
			rest = append(rest, m)
		}
	}
	return system, rest
}

// --- Wire types ---

type contentBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type contentMessage struct {
	Role    string         `json:"role"`
	Content []contentBlock `json:"content"`
}

type messagesRequest struct {
	Model       string           `json:"model"`
	MaxTokens   int              `json:"max_tokens"`
	System      string           `json:"system,omitempty"`
	Messages    []contentMessage `json:"messages"`
	Temperature float64          `json:"temperature,omitempty"`
	TopP        float64          `json:"top_p,omitempty"`
	Stop        []string         `json:"stop_sequences,omitempty"`
	Stream      bool             `json:"stream,omitempty"`
}

type messagesResponse struct {
	ID         string         `json:"id"`
	Model      string         `json:"model"`
	Content    []contentBlock `json:"content"`
	StopReason string         `json:"stop_reason"`
	Usage      struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

func toContentBlocks(msgs []llm.Message) []contentMessage {
	out := make([]contentMessage, len(msgs))
	for i, m := range msgs {
		out[i] = contentMessage{
			Role:    string(m.Role),
			Content: []contentBlock{{Type: "text", Text: m.Content}},
		}
	}
	return out
}
