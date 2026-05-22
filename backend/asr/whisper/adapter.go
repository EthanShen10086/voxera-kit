// Package whisper provides an OpenAI Whisper API implementation of the asr.Recognizer interface.
// It uses only the standard library to call the Whisper REST API directly.
package whisper

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"time"

	"github.com/EthanShen10086/voxera-kit/asr"
)

const defaultEndpoint = "https://api.openai.com"
const defaultModel = "whisper-1"

type whisperResponse struct {
	Text     string           `json:"text"`
	Segments []whisperSegment `json:"segments"`
}

type whisperSegment struct {
	Start float64 `json:"start"`
	End   float64 `json:"end"`
	Text  string  `json:"text"`
}

// Adapter implements the asr.Recognizer interface using the OpenAI Whisper API.
type Adapter struct {
	cfg    asr.Config
	client *http.Client
}

// New creates a new OpenAI Whisper Adapter with the provided configuration.
func New(cfg asr.Config) *Adapter {
	if cfg.Endpoint == "" {
		cfg.Endpoint = defaultEndpoint
	}
	if cfg.Model == "" {
		cfg.Model = defaultModel
	}
	return &Adapter{
		cfg:    cfg,
		client: &http.Client{Timeout: 120 * time.Second},
	}
}

// Recognize transcribes an audio file using the OpenAI Whisper API.
func (a *Adapter) Recognize(ctx context.Context, audioURL string, opts *asr.RecognizeOptions) ([]asr.Segment, error) {
	audioData, err := a.downloadAudio(ctx, audioURL)
	if err != nil {
		return nil, fmt.Errorf("whisper: download audio: %w", err)
	}

	result, err := a.transcribe(ctx, bytes.NewReader(audioData), opts)
	if err != nil {
		return nil, fmt.Errorf("whisper: transcribe: %w", err)
	}

	return result, nil
}

// RecognizeStream performs streaming recognition. The Whisper API does not support
// true streaming, so this buffers all audio and returns segments after full processing.
func (a *Adapter) RecognizeStream(ctx context.Context, reader io.Reader, opts *asr.RecognizeOptions) (<-chan asr.Segment, error) {
	ch := make(chan asr.Segment)

	go func() {
		defer close(ch)

		segments, err := a.transcribe(ctx, reader, opts)
		if err != nil {
			return
		}
		for _, seg := range segments {
			select {
			case <-ctx.Done():
				return
			case ch <- seg:
			}
		}
	}()

	return ch, nil
}

// Close releases all resources held by the Whisper client.
func (a *Adapter) Close() error {
	return nil
}

func (a *Adapter) downloadAudio(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d downloading audio", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

func (a *Adapter) transcribe(ctx context.Context, audio io.Reader, opts *asr.RecognizeOptions) ([]asr.Segment, error) {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	part, err := writer.CreateFormFile("file", "audio.wav")
	if err != nil {
		return nil, err
	}
	if _, err = io.Copy(part, audio); err != nil {
		return nil, err
	}

	_ = writer.WriteField("model", a.cfg.Model)
	_ = writer.WriteField("response_format", "verbose_json")

	if opts != nil && opts.Language != "" {
		_ = writer.WriteField("language", opts.Language)
	}

	if err = writer.Close(); err != nil {
		return nil, err
	}

	url := a.cfg.Endpoint + "/v1/audio/transcriptions"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, &body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	if a.cfg.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+a.cfg.APIKey)
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("whisper API error (%d): %s", resp.StatusCode, string(respBody))
	}

	var result whisperResponse
	if err = json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	segments := make([]asr.Segment, 0, len(result.Segments))
	for _, s := range result.Segments {
		segments = append(segments, asr.Segment{
			StartTime: time.Duration(s.Start * float64(time.Second)),
			EndTime:   time.Duration(s.End * float64(time.Second)),
			Text:      s.Text,
		})
	}

	return segments, nil
}

// compile-time interface check
var _ asr.Recognizer = (*Adapter)(nil)
