# Compatibility Matrix

**voxera-kit** version compatibility with downstream products.

| voxera-kit commit / tag | Go | MsgGuard | voxera | finera | pulsera |
|-------------------------|-----|----------|--------|--------|---------|
| **`ca77f67`** / **v0.3.0** (2026-06-13) | 1.22+ | pin updated | pin updated | pin updated | pin updated |
| **`132286e`** / **v0.2.0** (2026-06-13) | 1.22+ | superseded | superseded | superseded | superseded |
| `58be0db` / **v0.1.0** (2026-06-17) | 1.22+ | **v0.1.0** | superseded | superseded | superseded |
| `c037520` (2026-06) | 1.22+ | historical | historical | historical | historical |

> **v0.2.0** 包含数据平面 W1–W7 与测试基建 T1–T6。`cache/local.New` 签名变更：consumer 须 `_, err := local.New(cfg)`。

## Coordinated release train

1. Tag **voxera-kit** (`v0.x.y` or per-module tag).
2. Update `.github/voxera-kit-pin` in each consumer.
3. `go mod tidy` + CI green.
4. Tag product releases.

## Node / frontend packages

Frontend packages publish to npm independently; see root `CHANGELOG.md` for `@voxera-kit/*` versions.
