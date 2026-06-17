# auth

**Status:** Production (ports) · **Beta** (`jwt` adapter — stub)

Authentication and authorization **ports** (`Authenticator`, `Authorizer`) and domain types.

## Adapters

| Adapter | Status | Notes |
|---------|--------|-------|
| `jwt/` | Stub | TODO: sign/verify JWT |
| `oidc/` | Beta | OIDC client helpers |
| `oauth2/` | Beta | OAuth2 flows |

## Usage

```go
import "github.com/EthanShen10086/voxera-kit/auth"
```

Implement `Authenticator` in your app or use a memory/HMAC adapter (see MsgGuard `pkg/adapters/memory`).

## Tests

`port_test.go` — interface contract compile checks.
