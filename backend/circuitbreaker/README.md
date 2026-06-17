# circuitbreaker

**Status:** Production

Circuit breaker port with in-memory adapter (`memory/`).

## Config

- `MaxFailures` — open after N consecutive failures
- `Timeout` — time before half-open probe
- `HalfOpenMaxCalls` — concurrent probes in half-open

## Tests

`memory/adapter_test.go` — open/close state transitions.
