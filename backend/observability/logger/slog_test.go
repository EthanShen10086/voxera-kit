package logger_test

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"

	"github.com/EthanShen10086/voxera-kit/observability/logger"
)

func TestSlogAdapter_LevelsAndWith(t *testing.T) {
	var buf bytes.Buffer
	h := slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	log := logger.NewSlogAdapter(h)

	log.Debug("debug-msg", logger.Field{Key: "k", Value: "v"})
	log.Info("info-msg")
	log.Warn("warn-msg")
	log.Error("error-msg")

	child := log.With(logger.Field{Key: "service", Value: "kit"})
	child.WithTraceID("trace-1").Info("traced")

	out := buf.String()
	for _, want := range []string{"debug-msg", "info-msg", "warn-msg", "error-msg", "trace-1", "service"} {
		if !strings.Contains(out, want) {
			t.Fatalf("output missing %q: %s", want, out)
		}
	}
}

func TestSlogAdapter_Interface(t *testing.T) {
	var buf bytes.Buffer
	var _ logger.Logger = logger.NewSlogAdapter(slog.NewJSONHandler(&buf, nil))
	_ = context.Background()
}
