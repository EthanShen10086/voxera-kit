# Migration Guide

How to adopt or upgrade **voxera-kit** in application repos (MsgGuard, voxera, finera, pulsera).

## From ad-hoc utilities

1. Identify overlapping concerns (auth, metrics, middleware) in your app.
2. Add `require` for the specific kit module path, e.g. `github.com/EthanShen10086/voxera-kit/auth`.
3. Wire **ports** in application code; inject **adapters** at `main()` or DI container.
4. Remove duplicate middleware once parity is verified.

## Local development (`go.work` / `replace`)

```bash
# In your app's go.mod
replace github.com/EthanShen10086/voxera-kit/auth => ../voxera-kit/backend/auth
```

Pin CI to a commit SHA (see `.github/voxera-kit-pin` in consumer repos).

## Upgrading kit versions

1. Read [`CHANGELOG.md`](./CHANGELOG.md) and [`DEPRECATION.md`](./DEPRECATION.md).
2. Bump `go get` per module or coordinated tag from [`docs/COMPATIBILITY.md`](./docs/COMPATIBILITY.md).
3. Run consumer smoke: `go test` + gateway build in your product repo.
4. Deploy staging before production.

## JWT auth adapter

`auth/jwt` is currently a **stub** (interface only). Products should use their own HMAC/OIDC adapter or wait for JWT implementation before migrating off local auth.

## Breaking change policy

- **Minor** kit releases: additive only.
- **Major** (module path `/v2` or documented breaking ADR): 6-month deprecation window.
