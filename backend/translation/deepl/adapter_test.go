package deepl_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/EthanShen10086/voxera-kit/translation"
	"github.com/EthanShen10086/voxera-kit/translation/deepl"
)

func TestTranslate_Validation(t *testing.T) {
	a := deepl.New(translation.Config{})
	_, err := a.Translate(context.Background(), "hi", nil)
	if err == nil || !strings.Contains(err.Error(), "TargetLang") {
		t.Fatalf("Translate(nil opts): %v", err)
	}
	_, err = a.Translate(context.Background(), "hi", &translation.TranslateOptions{TargetLang: "en"})
	if err == nil || !strings.Contains(err.Error(), "APIKey") {
		t.Fatalf("Translate(no key): %v", err)
	}
}

func TestTranslate_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if auth := r.Header.Get("Authorization"); !strings.HasPrefix(auth, "DeepL-Auth-Key ") {
			t.Fatalf("Authorization = %q", auth)
		}
		if r.URL.Path != "/v2/translate" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"translations": []map[string]string{
				{"text": "hello", "detected_source_language": "ZH"},
			},
		})
	}))
	defer srv.Close()

	a := deepl.New(translation.Config{APIKey: "secret", Endpoint: srv.URL})
	res, err := a.Translate(context.Background(), "你好", &translation.TranslateOptions{
		TargetLang: "en",
		Formality:  "default",
		SourceLang: "zh",
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.Text != "hello" || res.DetectedSourceLang != "zh" {
		t.Fatalf("result = %+v", res)
	}
}

func TestTranslateBatchAndDetect(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"translations": []map[string]string{
				{"text": "done", "detected_source_language": "FR"},
			},
		})
	}))
	defer srv.Close()

	a := deepl.New(translation.Config{APIKey: "k", Endpoint: srv.URL})
	batch, err := a.TranslateBatch(context.Background(), []translation.BatchItem{
		{Text: "bonjour", TargetLang: "en"},
	})
	if err != nil || len(batch) != 1 || batch[0].Text != "done" {
		t.Fatalf("batch: %+v err=%v", batch, err)
	}

	lang, conf, err := a.DetectLanguage(context.Background(), "bonjour")
	if err != nil || lang != "fr" || conf != 0.95 {
		t.Fatalf("detect: lang=%q conf=%v err=%v", lang, conf, err)
	}
}

func TestSupportedLanguagesAndClose(t *testing.T) {
	a := deepl.New(translation.Config{})
	langs, err := a.SupportedLanguages(context.Background())
	if err != nil || len(langs) == 0 {
		t.Fatalf("SupportedLanguages: %v %v", langs, err)
	}
	if err := a.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestTranslate_APIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte("quota exceeded"))
	}))
	defer srv.Close()

	a := deepl.New(translation.Config{APIKey: "k", Endpoint: srv.URL})
	_, err := a.Translate(context.Background(), "x", &translation.TranslateOptions{TargetLang: "EN"})
	if err == nil || !strings.Contains(err.Error(), "403") {
		t.Fatalf("expected API error, got %v", err)
	}
}
