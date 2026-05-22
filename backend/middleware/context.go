package middleware

import "context"

// contextKey is an unexported type used for context value keys to prevent
// collisions with keys defined outside this package.
type contextKey string

const (
	// CtxRequestID is the context key for the request correlation ID.
	CtxRequestID contextKey = "request_id"
	// CtxTraceID is the context key for the distributed trace ID.
	CtxTraceID contextKey = "trace_id"
	// CtxUserID is the context key for the authenticated user identifier.
	CtxUserID contextKey = "user_id"
	// CtxTenantID is the context key for the tenant identifier.
	CtxTenantID contextKey = "tenant_id"
)

// RequestIDFromContext extracts the request ID stored by the [RequestID]
// middleware. It returns an empty string when no ID is present.
func RequestIDFromContext(ctx context.Context) string {
	v, _ := ctx.Value(CtxRequestID).(string)
	return v
}

// TraceIDFromContext extracts the trace ID stored by the [Tracing] middleware.
// It returns an empty string when no ID is present.
func TraceIDFromContext(ctx context.Context) string {
	v, _ := ctx.Value(CtxTraceID).(string)
	return v
}

// UserIDFromContext extracts the user ID from the context.
// It returns an empty string when no ID is present.
func UserIDFromContext(ctx context.Context) string {
	v, _ := ctx.Value(CtxUserID).(string)
	return v
}

// TenantIDFromContext extracts the tenant ID from the context.
// It returns an empty string when no ID is present.
func TenantIDFromContext(ctx context.Context) string {
	v, _ := ctx.Value(CtxTenantID).(string)
	return v
}
