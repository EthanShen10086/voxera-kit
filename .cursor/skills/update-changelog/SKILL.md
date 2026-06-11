---
name: update-changelog
description: 维护本仓库 CHANGELOG.md。post-commit 自动刷新 [Unreleased] 并 amend 入同一 commit；发版用 pnpm release 或 CI changelog-release workflow。
---

# 维护 CHANGELOG.md

## 做什么

- **自动**：`post-commit` 在 commit 落盘后刷新 `## [Unreleased]`（git log 已含 HEAD），再 amend 并入同一 commit
- **质量门禁**：`pre-commit` 仅 lint；`pre-push` 轻量复检（可用 `SKIP_PUSH_CHECK=1` 跳过）
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
CHANGELOG_SKIP=1 git commit ...          # 跳过 post-commit changelog
SKIP_PUSH_CHECK=1 git push               # 跳过 pre-push 复检（紧急）
```

## Hook 说明

| Hook | 职责 |
|------|------|
| pre-commit | lint-staged / gofmt，不含 CHANGELOG |
| commit-msg | Conventional Commits（commitlint） |
| post-commit | refresh Unreleased + amend（`CHANGELOG_AMENDING=1` 防递归） |
| pre-push | lint + typecheck / go vet 复检 |

关闭 CI：仓库 Variable `CI_RELEASE_ENABLED=false` 或删除 `changelog-release.yml`。
