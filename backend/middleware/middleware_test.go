package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/EthanShen10086/voxera-kit/audit"
	auditmemory "github.com/EthanShen10086/voxera-kit/audit/memory"
	loadshedadaptive "github.com/EthanShen10086/voxera-kit/loadshed/adaptive"
	"github.com/EthanShen10086/voxera-kit/loadshed"
	"github.com/EthanShen10086/voxera-kit/observability/logger"
	"github.com/EthanShen10086/voxera-kit/observability/metrics"
	"github.com/EthanShen10086/voxera-kit/observability/tracing"
	piiregex "github.com/EthanShen10086/voxera-kit/pii/regex"
	"github.com/EthanShen10086/voxera-kit/pii"
	"github.com/EthanShen10086/voxera-kit/security/headers"
)

func TestChainOrder(t *testing.T) {
	var order []string
	h := Chain(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		order = append(order, "handler")
	}),
		func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				order = append(order, "outer")
				next.ServeHTTP(w, r)
			})
		},
		func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				order = append(order, "inner")
				next.ServeHTTP(w, r)
			})
		},
	)
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil))
	if len(order) != 3 || order[0] != "outer" || order[1] != "inner" || order[2] != "handler" {
		t.Fatalf("order %v", order)
	}
}

func TestRequestIDPropagates(t *testing.T) {
	var got string
	h := Chain(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = RequestIDFromContext(r.Context())
	}), RequestID())
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Request-ID", "test-req-id")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if got != "test-req-id" {
		t.Fatalf("context id %q", got)
	}
	if rec.Header().Get("X-Request-ID") != "test-req-id" {
		t.Fatal("response header missing id")
	}
}

type stubChecker struct {
	err error
}

func (s stubChecker) Check(context.Context) error { return s.err }

type stubLogger struct{}

func (stubLogger) Debug(string, ...logger.Field) {}
func (stubLogger) Info(string, ...logger.Field)  {}
func (stubLogger) Warn(string, ...logger.Field)  {}
func (stubLogger) Error(string, ...logger.Field) {}
func (stubLogger) Fatal(string, ...logger.Field) {}
func (stubLogger) With(...logger.Field) logger.Logger { return stubLogger{} }
func (stubLogger) WithTraceID(string) logger.Logger { return stubLogger{} }

func TestHealthCheckEndpoints(t *testing.T) {
	h := Chain(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	}), HealthCheck(map[string]HealthChecker{
		"db": stubChecker{},
	}))

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/health", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("/health status = %d", rec.Code)
	}

	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/ready", nil))
	var resp healthResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}
	if resp.Status != "ok" || resp.Checks["db"] != "ok" {
		t.Fatalf("ready ok: %+v", resp)
	}

	hDegraded := Chain(http.NotFoundHandler(), HealthCheck(map[string]HealthChecker{
		"cache": stubChecker{err: errors.New("down")},
	}))
	rec = httptest.NewRecorder()
	hDegraded.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/ready", nil))
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("degraded status = %d", rec.Code)
	}
}

func TestSecurityHeaders(t *testing.T) {
	cfg := headers.DefaultStrict()
	h := Chain(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {}), SecurityHeaders(cfg))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Header().Get("Content-Security-Policy") == "" {
		t.Fatal("missing CSP")
	}
	if rec.Header().Get("X-Frame-Options") != "DENY" {
		t.Fatalf("X-Frame-Options = %q", rec.Header().Get("X-Frame-Options"))
	}
}

func TestRecoveryCatchesPanic(t *testing.T) {
	h := Chain(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("boom")
	}), Recovery(stubLogger{}))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/panic", nil))
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestTimeoutReturns503(t *testing.T) {
	h := Chain(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-r.Context().Done():
			return
		case <-time.After(200 * time.Millisecond):
			w.WriteHeader(http.StatusOK)
		}
	}), Timeout(20*time.Millisecond))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/slow", nil))
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestLoggingMiddleware(t *testing.T) {
	var logged bool
	h := Chain(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusCreated)
	}), Logging(stubLoggerWithHook{stubLogger: stubLogger{}, onInfo: func() { logged = true }}))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api", nil))
	if rec.Code != http.StatusCreated || !logged {
		t.Fatalf("status=%d logged=%v", rec.Code, logged)
	}
}

type stubLoggerWithHook struct {
	stubLogger
	onInfo func()
}

func (s stubLoggerWithHook) Info(string, ...logger.Field) {
	if s.onInfo != nil {
		s.onInfo()
	}
}

func TestContextHelpers(t *testing.T) {
	ctx := context.Background()
	ctx = context.WithValue(ctx, CtxTraceID, "trace-1")
	ctx = context.WithValue(ctx, CtxUserID, "user-1")
	ctx = context.WithValue(ctx, CtxTenantID, "tenant-1")
	if TraceIDFromContext(ctx) != "trace-1" {
		t.Fatal("trace id")
	}
	if UserIDFromContext(ctx) != "user-1" || TenantIDFromContext(ctx) != "tenant-1" {
		t.Fatal("user/tenant id")
	}
}

type stubRecorder struct{}

func (stubRecorder) Counter(string, float64, map[string]string)   {}
func (stubRecorder) Gauge(string, float64, map[string]string)     {}
func (stubRecorder) Histogram(string, float64, map[string]string) {}
func (stubRecorder) Timer(string) func()                          { return func() {} }

func TestMetricsMiddleware(t *testing.T) {
	h := Chain(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusAccepted)
	}), Metrics(stubRecorder{}))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/metrics-demo", nil))
	if rec.Code != http.StatusAccepted {
		t.Fatalf("status = %d", rec.Code)
	}
	var _ metrics.Recorder = stubRecorder{}
}

type stubTracer struct{}

func (stubTracer) Start(ctx context.Context, name string, _ ...tracing.SpanOption) (context.Context, tracing.Span) {
	_ = name
	return ctx, stubSpan{}
}

func (stubTracer) Shutdown(context.Context) error { return nil }

type stubSpan struct{}

func (stubSpan) End()                                      {}
func (stubSpan) SetAttributes(string, any)                 {}
func (stubSpan) RecordError(error)                         {}
func (stubSpan) SpanContext() tracing.SpanContext          { return tracing.SpanContext{TraceID: "trace-abc"} }

func TestTracingMiddleware(t *testing.T) {
	h := Chain(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if TraceIDFromContext(r.Context()) != "trace-abc" {
			t.Fatal("trace id not in context")
		}
	}), Tracing(stubTracer{}))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api", nil))
	if rec.Header().Get("X-Trace-ID") != "trace-abc" {
		t.Fatalf("trace header = %q", rec.Header().Get("X-Trace-ID"))
	}
}

func TestAuditMiddleware(t *testing.T) {
	writer := auditmemory.NewAdapter()
	h := Chain(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusCreated)
	}), Audit(writer, stubLogger{}))
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/items", strings.NewReader(`{"x":1}`))
	h.ServeHTTP(rec, req)
	entries, err := writer.Query(context.Background(), audit.Filter{})
	if err != nil || len(entries) != 1 || entries[0].ResponseStatus != http.StatusCreated {
		t.Fatalf("audit entries: %+v err=%v", entries, err)
	}
}

func TestLoadShedMiddleware(t *testing.T) {
	shedder := loadshedadaptive.New(loadshed.Config{MaxLoad: 0})
	h := Chain(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}), LoadShed(shedder))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestPIIRedactMiddleware(t *testing.T) {
	redactor := piiregex.NewRedactor(pii.Config{Rules: piiregex.DefaultRules(), DefaultMask: "[REDACTED]"})
	h := Chain(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "contact user@example.com", http.StatusBadRequest)
	}), PIIRedact(redactor))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
	if strings.Contains(rec.Body.String(), "user@example.com") {
		t.Fatalf("body not redacted: %q", rec.Body.String())
	}
}
