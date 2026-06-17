# Deprecation Policy

## Module lifecycle labels

| Label | Meaning |
|-------|---------|
| **Production** | Used in production workloads; API stable |
| **Beta** | API may change; adapters exist but limited soak |
| **Stub** | Port + no-op or TODO adapter only — **not for production** |

See per-module README for current label.

## Deprecation process

1. Mark API as deprecated in Go doc + CHANGELOG.
2. Emit compile-time or runtime warning for one minor release (when applicable).
3. Remove in next **major** module version or after **6 months**, whichever is later.

## Currently deprecated / stub

| Module | Item | Replacement |
|--------|------|-------------|
| `auth/jwt` | Full JWT implementation | Product-local auth until v0.2.0 |
| `llm/noop` | N/A (test double) | Use real provider adapters in prod |

## Notifications

Breaking removals are announced in:
- `CHANGELOG.md`
- GitHub Release notes
- `docs/COMPATIBILITY.md` matrix update
