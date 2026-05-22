// Package errors provides a structured application error type with error codes,
// wrapping, and convenience constructors for common failure categories.
package errors

import "fmt"

// ErrorCode is a machine-readable classification of an error.
type ErrorCode int

const (
	// OK means the operation completed successfully (not really an error).
	OK ErrorCode = 0
	// Unknown indicates an error whose cause cannot be classified.
	Unknown ErrorCode = 1
	// InvalidArgument signals that a caller-supplied value is invalid.
	InvalidArgument ErrorCode = 2
	// NotFound means the requested entity does not exist.
	NotFound ErrorCode = 3
	// AlreadyExists means the entity a caller tried to create already exists.
	AlreadyExists ErrorCode = 4
	// PermissionDenied means the caller lacks permission for the operation.
	PermissionDenied ErrorCode = 5
	// Unauthenticated means the request has no valid authentication credentials.
	Unauthenticated ErrorCode = 6
	// Internal indicates an unexpected server-side failure.
	Internal ErrorCode = 7
	// Unavailable means the service is temporarily unable to handle the request.
	Unavailable ErrorCode = 8
	// DeadlineExceeded means the operation timed out.
	DeadlineExceeded ErrorCode = 9
	// ResourceExhausted means a resource quota or limit was reached.
	ResourceExhausted ErrorCode = 10
	// Canceled means the operation was canceled by the caller.
	Canceled ErrorCode = 11
	// Unimplemented means the operation is not implemented or supported.
	Unimplemented ErrorCode = 12
)

// AppError is the canonical error type returned throughout the application.
type AppError struct {
	Code    ErrorCode
	Message string
	Cause   error
}

func (e *AppError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%d] %s: %v", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

// Unwrap returns the underlying cause so errors.Is / errors.As work.
func (e *AppError) Unwrap() error {
	return e.Cause
}

// New creates an [AppError] with the given code and message.
func New(code ErrorCode, message string) *AppError {
	return &AppError{Code: code, Message: message}
}

// Newf creates an [AppError] with a formatted message.
func Newf(code ErrorCode, format string, args ...any) *AppError {
	return &AppError{Code: code, Message: fmt.Sprintf(format, args...)}
}

// Wrap wraps an existing error with an application error code and message.
func Wrap(code ErrorCode, message string, err error) *AppError {
	return &AppError{Code: code, Message: message, Cause: err}
}

// Wrapf wraps an existing error with a formatted message.
func Wrapf(code ErrorCode, err error, format string, args ...any) *AppError {
	return &AppError{Code: code, Message: fmt.Sprintf(format, args...), Cause: err}
}

// Code extracts the [ErrorCode] from err if it is an [*AppError]; otherwise
// it returns [Unknown].
func Code(err error) ErrorCode {
	if err == nil {
		return OK
	}
	if ae, ok := err.(*AppError); ok {
		return ae.Code
	}
	return Unknown
}

// IsNotFound reports whether err has code [NotFound].
func IsNotFound(err error) bool { return Code(err) == NotFound }

// IsUnauthorized reports whether err has code [Unauthenticated].
func IsUnauthorized(err error) bool { return Code(err) == Unauthenticated }

// IsPermissionDenied reports whether err has code [PermissionDenied].
func IsPermissionDenied(err error) bool { return Code(err) == PermissionDenied }

// IsInvalidArgument reports whether err has code [InvalidArgument].
func IsInvalidArgument(err error) bool { return Code(err) == InvalidArgument }

// IsInternal reports whether err has code [Internal].
func IsInternal(err error) bool { return Code(err) == Internal }

// IsAlreadyExists reports whether err has code [AlreadyExists].
func IsAlreadyExists(err error) bool { return Code(err) == AlreadyExists }

// IsUnavailable reports whether err has code [Unavailable].
func IsUnavailable(err error) bool { return Code(err) == Unavailable }

// IsDeadlineExceeded reports whether err has code [DeadlineExceeded].
func IsDeadlineExceeded(err error) bool { return Code(err) == DeadlineExceeded }

// IsCanceled reports whether err has code [Canceled].
func IsCanceled(err error) bool { return Code(err) == Canceled }

// IsUnimplemented reports whether err has code [Unimplemented].
func IsUnimplemented(err error) bool { return Code(err) == Unimplemented }
