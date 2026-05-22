// Package azure provides an Azure Speech Services implementation of the asr.Recognizer interface.
// It is intended to use the Azure Cognitive Services Speech SDK for Go.
package azure

import (
	"context"
	"io"

	"github.com/EthanShen10086/voxera-kit/asr"
)

// Adapter implements the asr.Recognizer interface using Azure Speech Services.
//
// Intended dependency: github.com/Microsoft/cognitive-services-speech-sdk-go
type Adapter struct {
	// speechConfig *speech.SpeechConfig // TODO: uncomment when Azure SDK dependency is added
	cfg asr.Config
}

// New creates a new Azure Speech Services Adapter with the provided configuration.
func New(cfg asr.Config) *Adapter {
	return &Adapter{cfg: cfg}
}

// Recognize transcribes an audio file using Azure Speech Services.
func (a *Adapter) Recognize(ctx context.Context, audioURL string, opts *asr.RecognizeOptions) ([]asr.Segment, error) {
	// TODO: implement using Azure Speech SDK
	return nil, nil
}

// RecognizeStream performs continuous streaming recognition using Azure Speech Services.
func (a *Adapter) RecognizeStream(ctx context.Context, reader io.Reader, opts *asr.RecognizeOptions) (<-chan asr.Segment, error) {
	// TODO: implement using Azure Speech SDK continuous recognition
	ch := make(chan asr.Segment)
	close(ch)
	return ch, nil
}

// Close releases all resources held by the Azure Speech client.
func (a *Adapter) Close() error {
	// TODO: implement using Azure Speech SDK
	return nil
}
