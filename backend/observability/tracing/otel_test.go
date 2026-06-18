package tracing_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/EthanShen10086/voxera-kit/observability/tracing"
)

func TestNewOTelTracerExport(t *testing.T) {
	var exported bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/traces" {
			exported = true
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	host := strings.TrimPrefix(srv.URL, "http://")
	ctx := context.Background()
	tr, err := tracing.NewOTelTracer(ctx, tracing.OTelConfig{
		ServiceName: "svc",
		Endpoint:    host,
		Insecure:    true,
		SampleRate:  1,
		Headers:     map[string]string{"x-test": "1"},
	})
	if err != nil {
		t.Fatalf("NewOTelTracer: %v", err)
	}

	ctx, span := tr.Start(ctx, "op", tracing.SpanOption{
		Kind:       tracing.SpanKindClient,
		Attributes: map[string]any{"key": "val"},
	})
	span.SetAttributes("extra", 42)
	span.RecordError(errors.New("boom"))
	sc := span.SpanContext()
	if sc.TraceID == "" || sc.SpanID == "" {
		t.Fatalf("span context = %#v", sc)
	}
	span.End()

	if err := tr.Shutdown(ctx); err != nil {
		t.Fatalf("Shutdown: %v", err)
	}
	if !exported {
		t.Fatal("expected OTLP export")
	}
}

func TestNewOTelTracerSampleRates(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()
	host := strings.TrimPrefix(srv.URL, "http://")

	for _, rate := range []float64{0, 0.5} {
		tr, err := tracing.NewOTelTracer(context.Background(), tracing.OTelConfig{
			ServiceName: "svc",
			Endpoint:    host,
			Insecure:    true,
			SampleRate:  rate,
		})
		if err != nil {
			t.Fatalf("rate=%v: %v", rate, err)
		}
		_, span := tr.Start(context.Background(), "x")
		span.End()
		_ = tr.Shutdown(context.Background())
	}
}

func TestSpanKinds(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()
	host := strings.TrimPrefix(srv.URL, "http://")

	tr, err := tracing.NewOTelTracer(context.Background(), tracing.OTelConfig{
		ServiceName: "svc",
		Endpoint:    host,
		Insecure:    true,
	})
	if err != nil {
		t.Fatal(err)
	}
	kinds := []tracing.SpanKind{
		tracing.SpanKindInternal,
		tracing.SpanKindServer,
		tracing.SpanKindProducer,
		tracing.SpanKindConsumer,
	}
	for _, k := range kinds {
		_, span := tr.Start(context.Background(), "k", tracing.SpanOption{Kind: k})
		span.End()
	}
	_ = tr.Shutdown(context.Background())
}
