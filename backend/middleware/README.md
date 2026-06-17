# middleware

**Status:** Production

Composable HTTP middleware: RequestID, Logging, Tracing, Metrics, Recovery, Timeout, LoadShed, SecurityHeaders, PII redaction.

## Usage

```go
h := middleware.Chain(mux,
    middleware.RequestID(),
    middleware.Logging(log),
)
```

## Tests

`middleware_test.go` — chain order and RequestID propagation.
