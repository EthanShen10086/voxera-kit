// Package logger — standard library slog adapter.
//
// SlogAdapter wraps [log/slog.Logger] to satisfy the [Logger] interface,
// allowing callers to use the Go standard library's structured logger as
// a drop-in backend.
package logger

import (
	"context"
	"log/slog"
	"os"
)

// SlogAdapter implements [Logger] by delegating to [slog.Logger].
type SlogAdapter struct {
	logger *slog.Logger
}

// NewSlogAdapter creates a [SlogAdapter] from the given [slog.Handler].
//
// Pass any handler (e.g. [slog.NewJSONHandler]) to control output format.
func NewSlogAdapter(handler slog.Handler) *SlogAdapter {
	return &SlogAdapter{logger: slog.New(handler)}
}

func toSlogAttrs(fields []Field) []slog.Attr {
	attrs := make([]slog.Attr, len(fields))
	for i, f := range fields {
		attrs[i] = slog.Any(f.Key, f.Value)
	}
	return attrs
}

func attrsToArgs(attrs []slog.Attr) []any {
	args := make([]any, len(attrs))
	for i, a := range attrs {
		args[i] = a
	}
	return args
}

// Debug logs a message at DebugLevel.
func (s *SlogAdapter) Debug(msg string, fields ...Field) {
	s.logger.LogAttrs(context.Background(), slog.LevelDebug, msg, toSlogAttrs(fields)...)
}

// Info logs a message at InfoLevel.
func (s *SlogAdapter) Info(msg string, fields ...Field) {
	s.logger.LogAttrs(context.Background(), slog.LevelInfo, msg, toSlogAttrs(fields)...)
}

// Warn logs a message at WarnLevel.
func (s *SlogAdapter) Warn(msg string, fields ...Field) {
	s.logger.LogAttrs(context.Background(), slog.LevelWarn, msg, toSlogAttrs(fields)...)
}

// Error logs a message at ErrorLevel.
func (s *SlogAdapter) Error(msg string, fields ...Field) {
	s.logger.LogAttrs(context.Background(), slog.LevelError, msg, toSlogAttrs(fields)...)
}

// Fatal logs a message at ErrorLevel (slog has no Fatal) and terminates the
// process via [os.Exit].
func (s *SlogAdapter) Fatal(msg string, fields ...Field) {
	s.logger.LogAttrs(context.Background(), slog.LevelError, msg, toSlogAttrs(fields)...)
	os.Exit(1)
}

// With returns a child Logger that always includes the given fields.
func (s *SlogAdapter) With(fields ...Field) Logger {
	return &SlogAdapter{
		logger: s.logger.With(attrsToArgs(toSlogAttrs(fields))...),
	}
}

// WithTraceID returns a child Logger tagged with the given trace ID.
func (s *SlogAdapter) WithTraceID(traceID string) Logger {
	return s.With(Field{Key: "trace_id", Value: traceID})
}

var _ Logger = (*SlogAdapter)(nil)
