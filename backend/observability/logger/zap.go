// Package logger — Zap adapter.
//
// ZapLogger wraps [go.uber.org/zap] to satisfy the [Logger] interface.
// Import path for the real dependency: go.uber.org/zap
package logger

// ZapLogger implements [Logger] using Uber's Zap library.
//
// The underlying *zap.Logger is intentionally unexported; callers interact
// exclusively through the [Logger] interface.
type ZapLogger struct {
	// logger *zap.Logger  // uncomment when go.uber.org/zap is added to go.mod
	fields []Field
}

// NewZapLogger creates a production-ready [ZapLogger].
func NewZapLogger() (*ZapLogger, error) {
	// TODO: initialize zap.NewProduction() and wrap it
	return &ZapLogger{}, nil
}

// Debug logs a message at DebugLevel.
func (z *ZapLogger) Debug(msg string, fields ...Field) {
	// TODO: delegate to z.logger.Debug
}

// Info logs a message at InfoLevel.
func (z *ZapLogger) Info(msg string, fields ...Field) {
	// TODO: delegate to z.logger.Info
}

// Warn logs a message at WarnLevel.
func (z *ZapLogger) Warn(msg string, fields ...Field) {
	// TODO: delegate to z.logger.Warn
}

// Error logs a message at ErrorLevel.
func (z *ZapLogger) Error(msg string, fields ...Field) {
	// TODO: delegate to z.logger.Error
}

// Fatal logs a message at FatalLevel and then terminates the process.
func (z *ZapLogger) Fatal(msg string, fields ...Field) {
	// TODO: delegate to z.logger.Fatal
}

// With returns a child Logger that always includes the given fields.
func (z *ZapLogger) With(fields ...Field) Logger {
	child := &ZapLogger{fields: append(z.fields, fields...)}
	return child
}

// WithTraceID returns a child Logger tagged with the given trace ID.
func (z *ZapLogger) WithTraceID(traceID string) Logger {
	return z.With(Field{Key: "trace_id", Value: traceID})
}

var _ Logger = (*ZapLogger)(nil)
