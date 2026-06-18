package logger_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/EthanShen10086/voxera-kit/observability/logger"
)

func TestZapLogger_WritesToFile(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "app.log")

	log, err := logger.NewZapLogger(
		logger.WithLevel(logger.DebugLevel),
		logger.WithDevelopment(),
		logger.WithOutputPaths(logPath),
	)
	if err != nil {
		t.Fatal(err)
	}

	log.Info("hello zap", logger.Field{Key: "k", Value: "v"})
	log.WithTraceID("trace-xyz").Warn("warn-msg")

	data, err := os.ReadFile(logPath) //nolint:gosec // test reads log file under t.TempDir()
	if err != nil {
		t.Fatal(err)
	}
	out := string(data)
	if !strings.Contains(out, "hello zap") || !strings.Contains(out, "trace-xyz") {
		t.Fatalf("log output: %s", out)
	}
}

func TestZapLogger_Interface(t *testing.T) {
	log, err := logger.NewZapLogger(logger.WithOutputPaths(t.TempDir() + "/z.log"))
	if err != nil {
		t.Fatal(err)
	}
	var _ logger.Logger = log
	child := log.With(logger.Field{Key: "svc", Value: "kit"})
	child.Error("err-msg")
}
