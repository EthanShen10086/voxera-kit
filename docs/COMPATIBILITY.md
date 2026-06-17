# Compatibility Matrix

**voxera-kit** version compatibility with downstream products.

| voxera-kit commit / tag | Go | MsgGuard | voxera | finera | pulsera |
|-------------------------|-----|----------|--------|--------|---------|
| `58be0db` / **v0.1.0** (2026-06-17) | 1.22+ | **v0.1.0** | master | master | master |
| `c037520` (2026-06) | 1.22+ | ≥ `v0.1.0` | master | master | master |
| _future `v0.2.0`_ | 1.22+ | TBD | TBD | TBD | TBD |

## Coordinated release train

1. Tag **voxera-kit** (`v0.x.y` or per-module tag).
2. Update `.github/voxera-kit-pin` in each consumer.
3. `go mod tidy` + CI green.
4. Tag product releases.

## Node / frontend packages

Frontend packages publish to npm independently; see root `CHANGELOG.md` for `@voxera-kit/*` versions.
