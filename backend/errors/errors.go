// Package errors provides a structured, code-aware error type that works
// seamlessly with Go's standard error wrapping/unwrapping conventions.
//
// Every [AppError] carries a machine-readable [ErrorCode], a human-readable
// message, optional detail metadata, and an optional causal chain.
package errors

import (
	"errors"
	"fmt"
	"runtime"
	"strings"
)

// ErrorCode is a machine-readable classification for application errors.
type ErrorCode int

const (
	OK             ErrorCode = 0
	Unknown        ErrorCode = 1
	NotFound       ErrorCode = 2
	AlreadyExists  ErrorCode = 3
	InvalidArgument ErrorCode = 4
	Unauthorized   ErrorCode = 5
	Forbidden      ErrorCode = 6
	Internal       ErrorCode = 7
	Unavailable    ErrorCode = 8
	Timeout        ErrorCode = 9
	Conflict       ErrorCode = 10
	RateLimited    ErrorCode = 11
	Unimplemented  ErrorCode = 12
)

// String returns a human-readable name for the error code.
func (c ErrorCode) String() string {
	switch c {
	case OK:
		return "OK"
	case Unknown:
		return "Unknown"
	case NotFound:
		return "NotFound"
	case AlreadyExists:
		return "AlreadyExists"
	case InvalidArgument:
		return "InvalidArgument"
	case Unauthorized:
		return "Unauthorized"
	case Forbidden:
		return "Forbidden"
	case Internal:
		return "Internal"
	case Unavailable:
		return "Unavailable"
	case Timeout:
		return "Timeout"
	case Conflict:
		return "Conflict"
	case RateLimited:
		return "RateLimited"
	case Unimplemented:
		return "Unimplemented"
	default:
		return fmt.Sprintf("ErrorCode(%d)", int(c))
	}
}

// AppError is the canonical application error type.
type AppError struct {
	Code    ErrorCode
	Message string
	Details map[string]any
	Cause   error
	Stack   string
}

// Error implements the [error] interface.
func (e *AppError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Unwrap returns the underlying cause so that [errors.Is] and [errors.As] work
// across the error chain.
func (e *AppError) Unwrap() error {
	return e.Cause
}

// captureStack returns a compact stack trace string starting from the caller's caller.
func captureStack() string {
	const depth = 16
	var pcs [depth]uintptr
	n := runtime.Callers(3, pcs[:])
	frames := runtime.CallersFrames(pcs[:n])

	var b strings.Builder
	for {
		frame, more := frames.Next()
		fmt.Fprintf(&b, "%s\n\t%s:%d\n", frame.Function, frame.File, frame.Line)
		if !more {
			break
		}
	}
	return b.String()
}

// New creates an [AppError] with the given code and message.
func New(code ErrorCode, msg string) *AppError {
	return &AppError{
		Code:    code,
		Message: msg,
		Stack:   captureStack(),
	}
}

// Newf creates an [AppError] with a formatted message.
func Newf(code ErrorCode, format string, args ...any) *AppError {
	return &AppError{
		Code:    code,
		Message: fmt.Sprintf(format, args...),
		Stack:   captureStack(),
	}
}

// Wrap wraps an existing error with an [AppError], preserving the causal chain.
func Wrap(err error, code ErrorCode, msg string) *AppError {
	return &AppError{
		Code:    code,
		Message: msg,
		Cause:   err,
		Stack:   captureStack(),
	}
}

// Wrapf wraps an existing error with a formatted [AppError].
func Wrapf(err error, code ErrorCode, format string, args ...any) *AppError {
	return &AppError{
		Code:    code,
		Message: fmt.Sprintf(format, args...),
		Cause:   err,
		Stack:   captureStack(),
	}
}

// Code extracts the [ErrorCode] from err's chain.
// If no [AppError] is found, [Unknown] is returned.
func Code(err error) ErrorCode {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Code
	}
	return Unknown
}

// IsNotFound reports whether err's chain contains a [NotFound] error.
func IsNotFound(err error) bool { return Code(err) == NotFound }

// IsUnauthorized reports whether err's chain contains an [Unauthorized] error.
func IsUnauthorized(err error) bool { return Code(err) == Unauthorized }

// IsForbidden reports whether err's chain contains a [Forbidden] error.
func IsForbidden(err error) bool { return Code(err) == Forbidden }

// IsConflict reports whether err's chain contains a [Conflict] error.
func IsConflict(err error) bool { return Code(err) == Conflict }

// IsInvalidArgument reports whether err's chain contains an [InvalidArgument] error.
func IsInvalidArgument(err error) bool { return Code(err) == InvalidArgument }

// IsInternal reports whether err's chain contains an [Internal] error.
func IsInternal(err error) bool { return Code(err) == Internal }

// IsUnavailable reports whether err's chain contains an [Unavailable] error.
func IsUnavailable(err error) bool { return Code(err) == Unavailable }

// IsTimeout reports whether err's chain contains a [Timeout] error.
func IsTimeout(err error) bool { return Code(err) == Timeout }

// IsRateLimited reports whether err's chain contains a [RateLimited] error.
func IsRateLimited(err error) bool { return Code(err) == RateLimited }

// IsAlreadyExists reports whether err's chain contains an [AlreadyExists] error.
func IsAlreadyExists(err error) bool { return Code(err) == AlreadyExists }

// IsUnimplemented reports whether err's chain contains an [Unimplemented] error.
func IsUnimplemented(err error) bool { return Code(err) == Unimplemented }
