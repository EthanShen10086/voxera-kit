// Package asr defines the port interface for Automatic Speech Recognition operations.
// It abstracts transcription across different ASR providers (OpenAI Whisper,
// Azure Speech Services, Alibaba Cloud, self-hosted) using a unified interface.
package asr

import (
	"context"
	"io"
	"time"
)

// Segment represents a single transcribed segment of audio.
type Segment struct {
	// StartTime is the offset from audio start where this segment begins.
	StartTime time.Duration
	// EndTime is the offset from audio start where this segment ends.
	EndTime time.Duration
	// Text is the transcribed text content of this segment.
	Text string
	// Speaker is the identified speaker label (when diarization is enabled).
	Speaker string
	// Confidence is the recognition confidence score between 0.0 and 1.0.
	Confidence float64
}

// RecognizeOptions configures the behavior of a recognition request.
type RecognizeOptions struct {
	// Language is the BCP-47 language code (e.g., "en-US", "zh-CN").
	Language string
	// SpeakerDiarization enables speaker identification when true.
	SpeakerDiarization bool
	// MaxSpeakers is the maximum number of distinct speakers to identify.
	MaxSpeakers int
	// PunctuationEnabled enables automatic punctuation insertion when true.
	PunctuationEnabled bool
}

// Recognizer is the interface for performing speech-to-text recognition.
type Recognizer interface {
	// Recognize transcribes an audio file at the given URL and returns segments.
	Recognize(ctx context.Context, audioURL string, opts *RecognizeOptions) ([]Segment, error)
	// RecognizeStream performs streaming recognition from an audio reader,
	// returning segments as they are identified via a channel.
	RecognizeStream(ctx context.Context, reader io.Reader, opts *RecognizeOptions) (<-chan Segment, error)
	// Close releases all resources held by the recognizer.
	Close() error
}

// ASRConfig holds the configuration parameters for an ASR provider.
type ASRConfig struct {
	// Endpoint is the API endpoint URL for the ASR service.
	Endpoint string
	// APIKey is the authentication key for the ASR service.
	APIKey string
	// Region is the service region (applicable to cloud providers).
	Region string
	// Model specifies the recognition model to use (e.g., "whisper-1", "large-v3").
	Model string
}
