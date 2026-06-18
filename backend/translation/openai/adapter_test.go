package openai_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/EthanShen10086/voxera-kit/translation"
	"github.com/EthanShen10086/voxera-kit/translation/openai"
)

func chatHandler(t *testing.T, content string) http.HandlerFunc {
	t.Helper()
	return func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.Header.Get("Authorization"), "Bearer ") {
			t.Errorf("missing auth header")
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{
				{"message": map[string]string{"content": content}},
			},
		})
	}
}

func TestTranslateAndDetect(t *testing.T) {
	srv := httptest.NewServer(chatHandler(t, "hello"))
	defer srv.Close()

	a := openai.New(translation.Config{Endpoint: srv.URL, APIKey: "sk-test"})
	res, err := a.Translate(context.Background(), "你好", &translation.TranslateOptions{
		TargetLang: "en", SourceLang: "zh", Formality: "formal",
	})
	if err != nil || res.Text != "hello" {
		t.Fatalf("Translate: %+v err=%v", res, err)
	}

	lang, conf, err := a.DetectLanguage(context.Background(), "bonjour")
	if err != nil || lang != "hello" || conf != 0.9 {
		t.Fatalf("DetectLanguage: lang=%q conf=%v err=%v", lang, conf, err)
	}
}

func TestTranslateBatchAndSupportedLanguages(t *testing.T) {
	srv := httptest.NewServer(chatHandler(t, "1. hola\n2. salut"))
	defer srv.Close()

	a := openai.New(translation.Config{Endpoint: srv.URL, APIKey: "k"})
	batch, err := a.TranslateBatch(context.Background(), []translation.BatchItem{
		{Text: "hi", TargetLang: "es"},
		{Text: "bye", TargetLang: "fr"},
	})
	if err != nil || len(batch) != 2 || batch[0].Text != "hola" {
		t.Fatalf("batch: %+v err=%v", batch, err)
	}

	langs, err := a.SupportedLanguages(context.Background())
	if err != nil || len(langs) == 0 {
		t.Fatalf("langs: %v err=%v", langs, err)
	}
	if err := a.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestTranslate_APIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte("rate limited"))
	}))
	defer srv.Close()

	a := openai.New(translation.Config{Endpoint: srv.URL, APIKey: "k"})
	_, err := a.Translate(context.Background(), "x", &translation.TranslateOptions{TargetLang: "en"})
	if err == nil || !strings.Contains(err.Error(), "429") {
		t.Fatalf("expected API error, got %v", err)
	}
}

func TestTranslateBatch_Empty(t *testing.T) {
	a := openai.New(translation.Config{})
	batch, err := a.TranslateBatch(context.Background(), nil)
	if err != nil || batch != nil {
		t.Fatalf("empty batch: %v err=%v", batch, err)
	}
}
