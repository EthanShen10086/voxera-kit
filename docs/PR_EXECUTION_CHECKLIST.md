# PR Execution Checklist

Track merge order for the org-wide standards rollout. Check off as PRs land.

## Phase 0 — Org template
- [x] `EthanShen10086/.github` profile + SECURITY template
- [x] `voxera-kit/docs/REPO_STANDARDS.md`
- [x] `voxera-kit` reusable-security / reusable-release workflows

## Phase 1 — MsgGuard
- [x] PR-MG-1 Governance (LICENSE, SECURITY, CODEOWNERS, CHANGELOG, dependabot)
- [x] PR-MG-2 CI tests + verify.sh
- [x] PR-MG-3 Release supply chain (SBOM, cosign, v0.1.0)
- [x] PR-MG-4 ADR + VERSIONING + ONCALL

## Phase 2 — voxera-kit
- [x] PR-KIT-1 Contract tests + vitest CI + codecov
- [x] PR-KIT-2 MIGRATION/DEPRECATION + module READMEs + examples
- [x] PR-KIT-3 Consumer smoke (MsgGuard build)

## Phase 3 — Products
- [x] PR-VX-1 voxera: SECURITY, lint blocking, SRE docs, prod deploy doc
- [x] PR-FN-1 finera: SECURITY, lint blocking, SLO stub
- [x] PR-PL-1 pulsera: CODEOWNERS, CHANGELOG, CD workflows
- [x] PR-CP-1 ciphera: SECURITY, CODEOWNERS, dependabot (npm)
- [x] PR-CV-1 ciphera-vps: LICENSE, tests ≥40%, release SBOM

## Phase 4 — ScrollCap
- [x] PR-SC-1 SECURITY, CHANGELOG, CONTRIBUTING, TestFlight workflow

## Phase 5 — Coordinated release
- [x] `docs/COMPATIBILITY.md` in voxera-kit
- [x] **v0.2.0** tag + consumer pin bump (2026-06-13)
- [x] `docs/RELEASE_TRAIN.md` operational release process
- [x] **v0.3.0** tag + consumer pin bump (2026-06-13)
- [x] Coverage ramp + MIN_COVERAGE enforce (`docs/COVERAGE_ROADMAP.md`)
- [x] `@voxera-kit/faker` pluggable package
- [x] `scripts/monthly-pin-bump.sh` + RELEASE_TRAIN step
- [x] ScrollCap remote: dependabot + SECURITY verified
- [ ] Monthly kit tag → bump consumer pins (operational cadence)
