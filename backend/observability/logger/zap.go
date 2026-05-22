// Package logger — Zap adapter.
//
// ZapLogger wraps [go.uber.org/zap] to satisfy the [Logger] interface.
package logger

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Option configures a [ZapLogger] during construction.
type Option func(*options)

type options struct {
	level       Level
	development bool
	outputPaths []string
}

// WithLevel sets the minimum log level for the logger.
func WithLevel(l Level) Option {
	return func(o *options) { o.level = l }
}

// WithDevelopment enables development mode (human-readable output, DPanic
// panics, stack traces on warnings).
func WithDevelopment() Option {
	return func(o *options) { o.development = true }
}

// WithOutputPaths sets the output destinations (e.g. "stdout", file paths).
func WithOutputPaths(paths ...string) Option {
	return func(o *options) { o.outputPaths = paths }
}

// ZapLogger implements [Logger] using Uber's Zap library.
//
// The underlying *zap.Logger is intentionally unexported; callers interact
// exclusively through the [Logger] interface.
type ZapLogger struct {
	logger *zap.Logger
}

func toZapLevel(l Level) zapcore.Level {
	switch l {
	case DebugLevel:
		return zapcore.DebugLevel
	case InfoLevel:
		return zapcore.InfoLevel
	case WarnLevel:
		return zapcore.WarnLevel
	case ErrorLevel:
		return zapcore.ErrorLevel
	case FatalLevel:
		return zapcore.FatalLevel
	default:
		return zapcore.InfoLevel
	}
}

func toZapFields(fields []Field) []zap.Field {
	zf := make([]zap.Field, len(fields))
	for i, f := range fields {
		zf[i] = zap.Any(f.Key, f.Value)
	}
	return zf
}

// NewZapLogger creates a production-ready [ZapLogger].
//
// Use [Option] values such as [WithLevel] and [WithDevelopment] to customize
// the underlying zap configuration.
func NewZapLogger(opts ...Option) (*ZapLogger, error) {
	o := options{
		level:       InfoLevel,
		outputPaths: []string{"stdout"},
	}
	for _, fn := range opts {
		fn(&o)
	}

	var cfg zap.Config
	if o.development {
		cfg = zap.NewDevelopmentConfig()
	} else {
		cfg = zap.NewProductionConfig()
	}
	cfg.Level = zap.NewAtomicLevelAt(toZapLevel(o.level))
	if len(o.outputPaths) > 0 {
		cfg.OutputPaths = o.outputPaths
	}

	l, err := cfg.Build()
	if err != nil {
		return nil, fmt.Errorf("build zap logger: %w", err)
	}
	return &ZapLogger{logger: l}, nil
}

// Debug logs a message at DebugLevel.
func (z *ZapLogger) Debug(msg string, fields ...Field) {
	z.logger.Debug(msg, toZapFields(fields)...)
}

// Info logs a message at InfoLevel.
func (z *ZapLogger) Info(msg string, fields ...Field) {
	z.logger.Info(msg, toZapFields(fields)...)
}

// Warn logs a message at WarnLevel.
func (z *ZapLogger) Warn(msg string, fields ...Field) {
	z.logger.Warn(msg, toZapFields(fields)...)
}

// Error logs a message at ErrorLevel.
func (z *ZapLogger) Error(msg string, fields ...Field) {
	z.logger.Error(msg, toZapFields(fields)...)
}

// Fatal logs a message at FatalLevel and then terminates the process.
func (z *ZapLogger) Fatal(msg string, fields ...Field) {
	z.logger.Fatal(msg, toZapFields(fields)...)
}

// With returns a child Logger that always includes the given fields.
func (z *ZapLogger) With(fields ...Field) Logger {
	return &ZapLogger{logger: z.logger.With(toZapFields(fields)...)}
}

// WithTraceID returns a child Logger tagged with the given trace ID.
func (z *ZapLogger) WithTraceID(traceID string) Logger {
	return z.With(Field{Key: "trace_id", Value: traceID})
}

var _ Logger = (*ZapLogger)(nil)
