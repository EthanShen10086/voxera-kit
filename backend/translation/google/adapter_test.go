package google_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/EthanShen10086/voxera-kit/translation"
	"github.com/EthanShen10086/voxera-kit/translation/google"
)

func TestTranslate_Validation(t *testing.T) {
	a := google.New(translation.Config{})
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
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "/language/translate/v2") {
			t.Fatalf("path = %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": map[string]any{
				"translations": []map[string]string{
					{"translatedText": "hello", "detectedSourceLanguage": "zh"},
				},
			},
		})
	}))
	defer srv.Close()

	a := google.New(translation.Config{APIKey: "test-key", Endpoint: srv.URL})
	res, err := a.Translate(context.Background(), "你好", &translation.TranslateOptions{TargetLang: "en"})
	if err != nil {
		t.Fatal(err)
	}
	if res.Text != "hello" || res.DetectedSourceLang != "zh" {
		t.Fatalf("result = %+v", res)
	}
}

func TestTranslateBatchAndDetect(t *testing.T) {
	var calls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		switch {
		case strings.HasSuffix(r.URL.Path, "/detect"):
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": map[string]any{
					"detections": [][]map[string]any{
						{{"language": "de", "confidence": 0.91}},
					},
				},
			})
		default:
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": map[string]any{
					"translations": []map[string]string{
						{"translatedText": "ok", "detectedSourceLanguage": "en"},
					},
				},
			})
		}
	}))
	defer srv.Close()

	a := google.New(translation.Config{APIKey: "k", Endpoint: srv.URL})
	batch, err := a.TranslateBatch(context.Background(), []translation.BatchItem{
		{Text: "a", TargetLang: "en"},
		{Text: "b", TargetLang: "fr", SourceLang: "en"},
	})
	if err != nil || len(batch) != 2 || batch[0].Text != "ok" {
		t.Fatalf("batch: %v err=%v", batch, err)
	}

	lang, conf, err := a.DetectLanguage(context.Background(), "hallo")
	if err != nil || lang != "de" || conf != 0.91 {
		t.Fatalf("detect: lang=%q conf=%v err=%v", lang, conf, err)
	}
	if calls < 3 {
		t.Fatalf("expected >=3 API calls, got %d", calls)
	}
}

func TestSupportedLanguagesAndClose(t *testing.T) {
	a := google.New(translation.Config{})
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
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("bad request"))
	}))
	defer srv.Close()

	a := google.New(translation.Config{APIKey: "k", Endpoint: srv.URL})
	_, err := a.Translate(context.Background(), "x", &translation.TranslateOptions{TargetLang: "en"})
	if err == nil || !strings.Contains(err.Error(), "400") {
		t.Fatalf("expected API error, got %v", err)
	}
}
