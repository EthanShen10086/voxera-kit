# Repository Standards (EthanShen10086)

Organization-wide engineering contract. Copy templates from `templates/org-standard/` or [`.github`](https://github.com/EthanShen10086/.github).

## Required files

| File | Purpose |
|------|---------|
| `LICENSE` | MIT unless otherwise noted |
| `SECURITY.md` | Vulnerability reporting + SLA |
| `CODE_OF_CONDUCT.md` | Community standards |
| `CHANGELOG.md` | Semver user-facing changes |
| `CONTRIBUTING.md` | Branch strategy, PR checklist |
| `.github/CODEOWNERS` | Review routing |
| `.github/dependabot.yml` | Weekly dependency PRs |
| `.github/ISSUE_TEMPLATE/` | Bug + feature forms |

## CI gates (minimum)

1. **Lint** — must fail on error (no `\|\| true`)
2. **Test** — `go test -race` / `vitest` / `swift test` with **real test files**
3. **Build** — all deployable artifacts
4. **Security** — gosec + npm audit (blocking on default branch)

## Release

- Tag `v*.*.*` triggers release workflow
- GitHub Release notes from CHANGELOG
- Container images: SBOM (Syft) + cosign sign when publishing to GHCR
- Pin `voxera-kit` to commit SHA in CI (not floating `main`)

## ADR

Significant architecture decisions go in `docs/adr/` or `docs/architecture-decisions.md`.

## Reusable workflows

| Workflow | Use |
|----------|-----|
| `reusable-go-ci.yml` | Go modules |
| `reusable-ts-ci.yml` | TypeScript packages |
| `reusable-security.yml` | gosec, npm audit, helm secrets |
| `reusable-release.yml` | SBOM + optional cosign |

## Product-specific

| Repo | Extra |
|------|-------|
| MsgGuard | `make verify` in CI, ML benchmark gate |
| voxera/finera | Multi-service Docker CD |
| ScrollCap | SwiftLint strict, TestFlight workflow |
| ciphera-vps | Binary release + checksums |

## Compatibility

See [COMPATIBILITY.md](COMPATIBILITY.md) for voxera-kit ↔ product version matrix.
