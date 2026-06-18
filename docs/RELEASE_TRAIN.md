# Release train

Operational process for **voxera-kit → consumer** coordinated releases.

## When to release

- After merging a kit milestone (data-plane wave, testing wave, or breaking change)
- At least **monthly** for security/dependency updates (see `PR_EXECUTION_CHECKLIST.md`)

## Steps

1. **Verify kit CI** — `master` green (unit, contract, integration on main, coverage upload).
2. **Update CHANGELOG** — move `[Unreleased]` to `v0.x.y` with date and breaking notes.
3. **Tag voxera-kit**
   ```bash
   git tag -a v0.x.y -m "v0.x.y summary"
   git push origin v0.x.y
   ```
4. **Record SHA** in `docs/COMPATIBILITY.md` (commit on master after tag; pin = `git rev-parse v0.x.y^{commit}`).
5. **Create GitHub Release** (optional but recommended)
   ```bash
   gh release create v0.x.y --notes-file /tmp/release-notes.md
   ```
6. **Bump consumer pins** — update `.github/voxera-kit-pin` in:
   - voxera, finera, pulsera, MsgGuard, ciphera, ciphera-vps
7. **Consumer CI** — push each consumer; fix breaking changes (e.g. `cache/local.New` error return).
8. **Product tags** (optional) — tag product releases after consumer CI is green.

## Pin file

Single-line full commit SHA:

```
132286ea38c3a3300c43c72c140c8cc2dc34e984
```

CI checks out kit at exactly this SHA — no floating `master`.

## Breaking change policy

- Document in CHANGELOG and COMPATIBILITY matrix row
- Prefer minor bump (`v0.x.0`) for port signature changes
- Consumers must not pin until migration PR is ready

## Related

- [`COMPATIBILITY.md`](./COMPATIBILITY.md)
- [`testing.md`](./testing.md)
- Org profile: [EthanShen10086/.github](https://github.com/EthanShen10086/.github)
