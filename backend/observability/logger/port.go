// Package logger defines the structured logging port used across the toolkit.
//
// Adapters (e.g. Zap, Zerolog) implement the [Logger] interface so that
// application code stays decoupled from any specific logging library.
package logger

// Level represents the severity of a log entry.
type Level int

const (
	// DebugLevel is the most verbose level, useful during development.
	DebugLevel Level = iota
	// InfoLevel is the default production level.
	InfoLevel
	// WarnLevel indicates a potentially harmful situation.
	WarnLevel
	// ErrorLevel indicates an error that does not halt the program.
	ErrorLevel
	// FatalLevel logs the message and then calls os.Exit(1).
	FatalLevel
)

// Field is a single key-value pair attached to a log entry.
type Field struct {
	Key   string
	Value any
}

// Logger is the primary structured logging port.
type Logger interface {
	// Debug logs a message at DebugLevel with optional structured fields.
	Debug(msg string, fields ...Field)

	// Info logs a message at InfoLevel with optional structured fields.
	Info(msg string, fields ...Field)

	// Warn logs a message at WarnLevel with optional structured fields.
	Warn(msg string, fields ...Field)

	// Error logs a message at ErrorLevel with optional structured fields.
	Error(msg string, fields ...Field)

	// Fatal logs a message at FatalLevel, then terminates the process.
	Fatal(msg string, fields ...Field)

	// With returns a child Logger that always includes the given fields.
	With(fields ...Field) Logger

	// WithTraceID returns a child Logger that tags every entry with a trace ID.
	WithTraceID(traceID string) Logger
}
