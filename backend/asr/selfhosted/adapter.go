// Package selfhosted provides a self-hosted Whisper API implementation of the asr.Recognizer interface.
// It is intended to communicate with a locally deployed Whisper inference server
// (e.g., faster-whisper-server or whisper.cpp HTTP API) via REST.
package selfhosted

import (
	"context"
	"io"

	"github.com/EthanShen10086/voxera-kit/asr"
)

// Adapter implements the asr.Recognizer interface using a self-hosted Whisper API.
// This connects to a locally deployed inference server via HTTP,
// enabling on-premise or air-gapped deployments.
type Adapter struct {
	// httpClient *http.Client // standard library http client
	cfg asr.Config
}

// New creates a new self-hosted Whisper Adapter with the provided configuration.
// The cfg.Endpoint should point to the self-hosted Whisper HTTP API base URL.
func New(cfg asr.Config) *Adapter {
	return &Adapter{cfg: cfg}
}

// Recognize transcribes an audio file by sending it to the self-hosted Whisper API.
func (a *Adapter) Recognize(ctx context.Context, audioURL string, opts *asr.RecognizeOptions) ([]asr.Segment, error) {
	// TODO: implement HTTP POST to self-hosted Whisper API
	return nil, nil
}

// RecognizeStream performs streaming recognition via the self-hosted Whisper API.
// Depending on the server implementation, this may use WebSocket or chunked transfer.
func (a *Adapter) RecognizeStream(ctx context.Context, reader io.Reader, opts *asr.RecognizeOptions) (<-chan asr.Segment, error) {
	// TODO: implement streaming via WebSocket or chunked HTTP
	ch := make(chan asr.Segment)
	close(ch)
	return ch, nil
}

// Close releases all resources held by the HTTP client.
func (a *Adapter) Close() error {
	// TODO: close idle HTTP connections
	return nil
}
