// Package whisper provides an OpenAI Whisper API implementation of the asr.Recognizer interface.
// It is intended to use the OpenAI API (github.com/sashabaranov/go-openai) for transcription.
package whisper

import (
	"context"
	"io"

	"github.com/EthanShen10086/voxera-kit/asr"
)

// Adapter implements the asr.Recognizer interface using the OpenAI Whisper API.
//
// Intended dependency: github.com/sashabaranov/go-openai
type Adapter struct {
	// client *openai.Client // TODO: uncomment when go-openai dependency is added
	cfg asr.Config
}

// New creates a new OpenAI Whisper Adapter with the provided configuration.
func New(cfg asr.Config) *Adapter {
	return &Adapter{cfg: cfg}
}

// Recognize transcribes an audio file using the OpenAI Whisper API.
func (a *Adapter) Recognize(ctx context.Context, audioURL string, opts *asr.RecognizeOptions) ([]asr.Segment, error) {
	// TODO: implement using go-openai
	return nil, nil
}

// RecognizeStream performs streaming recognition. Note: the OpenAI API may not
// support true streaming; this may buffer and return segments after full processing.
func (a *Adapter) RecognizeStream(ctx context.Context, reader io.Reader, opts *asr.RecognizeOptions) (<-chan asr.Segment, error) {
	// TODO: implement using go-openai
	ch := make(chan asr.Segment)
	close(ch)
	return ch, nil
}

// Close releases all resources held by the Whisper client.
func (a *Adapter) Close() error {
	// TODO: implement using go-openai
	return nil
}
