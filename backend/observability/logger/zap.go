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
	// TODO: initialise zap.NewProduction() and wrap it
	return &ZapLogger{}, nil
}

func (z *ZapLogger) Debug(msg string, fields ...Field) {
	// TODO: delegate to z.logger.Debug
}

func (z *ZapLogger) Info(msg string, fields ...Field) {
	// TODO: delegate to z.logger.Info
}

func (z *ZapLogger) Warn(msg string, fields ...Field) {
	// TODO: delegate to z.logger.Warn
}

func (z *ZapLogger) Error(msg string, fields ...Field) {
	// TODO: delegate to z.logger.Error
}

func (z *ZapLogger) Fatal(msg string, fields ...Field) {
	// TODO: delegate to z.logger.Fatal
}

func (z *ZapLogger) With(fields ...Field) Logger {
	child := &ZapLogger{fields: append(z.fields, fields...)}
	return child
}

func (z *ZapLogger) WithTraceID(traceID string) Logger {
	return z.With(Field{Key: "trace_id", Value: traceID})
}

var _ Logger = (*ZapLogger)(nil)
