# @voxera-kit/testing

Wave T4 前端测试基建：Vitest 配置片段、MSW 通用 handler、React provider 渲染辅助。

## Vitest

```typescript
import { defineConfig, mergeConfig } from "vitest/config";
import { setupVitest } from "@voxera-kit/testing/vitest";

export default mergeConfig(
  setupVitest({ test: { environment: "jsdom" } }),
  defineConfig({ test: { include: ["src/**/*.test.ts"] } }),
);
```

## MSW

```typescript
import { beforeAll, afterAll, afterEach } from "vitest";
import { createMockServer } from "@voxera-kit/testing/msw";

const server = createMockServer();
beforeAll(() => server.listen());
afterEach(() => server.resetHandlers());
afterAll(() => server.close());
```

内置 handler：`/api/auth/*`、分页 `/api/items`、错误码 `/api/errors/:code`。

## React（可选）

```tsx
/** @vitest-environment jsdom */
import { renderWithProviders, registerReactTestingCleanup } from "@voxera-kit/testing/react";

registerReactTestingCleanup();

renderWithProviders(<MyComponent />);
```
