package whisper_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/EthanShen10086/voxera-kit/asr"
	"github.com/EthanShen10086/voxera-kit/asr/whisper"
)

func TestRecognizeFromURL(t *testing.T) {
	audioSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("fake-audio"))
	}))
	defer audioSrv.Close()

	apiSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/v1/audio/transcriptions") {
			t.Fatalf("path = %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"text": "hello world",
			"segments": []map[string]any{
				{"start": 0.0, "end": 1.5, "text": "hello world"},
			},
		})
	}))
	defer apiSrv.Close()

	a := whisper.New(asr.Config{Endpoint: apiSrv.URL, APIKey: "sk-test"})
	segments, err := a.Recognize(context.Background(), audioSrv.URL, &asr.RecognizeOptions{Language: "en"})
	if err != nil {
		t.Fatal(err)
	}
	if len(segments) != 1 || segments[0].Text != "hello world" {
		t.Fatalf("segments = %+v", segments)
	}
	if segments[0].EndTime != time.Duration(1.5*float64(time.Second)) {
		t.Fatalf("end = %v", segments[0].EndTime)
	}
}

func TestRecognizeStream(t *testing.T) {
	apiSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"text": "stream",
			"segments": []map[string]any{
				{"start": 0.0, "end": 0.5, "text": "stream"},
			},
		})
	}))
	defer apiSrv.Close()

	a := whisper.New(asr.Config{Endpoint: apiSrv.URL, APIKey: "k"})
	ch, err := a.RecognizeStream(context.Background(), strings.NewReader("audio"), nil)
	if err != nil {
		t.Fatal(err)
	}
	var count int
	for range ch {
		count++
	}
	if count != 1 {
		t.Fatalf("got %d segments", count)
	}
}

func TestRecognizeAPIError(t *testing.T) {
	audioSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("x"))
	}))
	defer audioSrv.Close()

	apiSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer apiSrv.Close()

	a := whisper.New(asr.Config{Endpoint: apiSrv.URL, APIKey: "bad"})
	_, err := a.Recognize(context.Background(), audioSrv.URL, nil)
	if err == nil {
		t.Fatal("expected error")
	}
}
