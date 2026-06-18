# Playwright E2E 模板（Wave T5）

可复制到 voxera / finera / pulsera 等产品仓的 E2E 起步模板。包含：

- `playwright.config.ts` — 本地 demo 静态站 + Chromium
- `pages/LoginPage.ts` — Page Object 示例
- `tests/login.spec.ts` — 登录 / 登出 smoke flow
- `demo/` — 最小静态页面（无需产品前端构建）

## 本地运行

```bash
cd templates/e2e-playwright
npm install
npm run install:browsers
npm test
```

## 复制到产品仓

1. 复制整个 `templates/e2e-playwright/` 到产品仓 `e2e/`（或 `apps/web/e2e/`）
2. 将 `demo/` 替换为产品 dev server URL，或保留 demo 做 smoke
3. 在 `playwright.config.ts` 中修改 `webServer.command` / `baseURL`
4. 复制 `.github/e2e-smoke.yml.example` 到产品 `.github/workflows/`

## 与 kit 测试基建的关系

| 层级 | 包 / 模板 | 用途 |
|------|-----------|------|
| 单测 | `@voxera-kit/testing` + MSW | 组件 / API client |
| 造数 | `@voxera-kit/fixture` | JSON 工厂 |
| E2E | 本模板 | 浏览器 smoke / 登录流 |

## CI

见 `.github/e2e-smoke.yml.example`。voxera-kit 主仓在 push 到 `master`/`main` 时运行模板 smoke（`e2e-smoke` job）。
