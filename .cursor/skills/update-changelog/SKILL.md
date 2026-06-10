---
name: update-changelog
description: 维护本仓库 CHANGELOG.md。commit 时自动刷新 [Unreleased] 底稿；发版用 pnpm release 或 CI changelog-release workflow。
---

# 维护 CHANGELOG.md

## 做什么

- **自动**：pre-commit 通过后刷新 `## [Unreleased]`（来自 Conventional Commits）
- **发版本地**：`pnpm release:dry-run` → `pnpm release`（算 semver、固化 CHANGELOG、打 `vX.Y.Z` tag）
- **发版 CI**：默认启用 `.github/workflows/changelog-release.yml`（手动 dispatch 或 main 上含 `[release]` 的 commit）

## 记什么 / 不记什么

**保留**：新功能、行为变更、重要修复、影响流程的重构、接口/配置变更（不限条数）

**可删**：纯样式、chore/ci、无行为变化的重命名、Revert 记录

## 常用命令

```bash
pnpm run changelog:refresh
pnpm run release:dry-run
pnpm release
CHANGELOG_SKIP=1 git commit ...
```

关闭 CI：仓库 Variable `CI_RELEASE_ENABLED=false` 或删除 `changelog-release.yml`。
