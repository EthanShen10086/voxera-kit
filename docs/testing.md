# 产品测试接入指南

本文说明 **voxera / finera / pulsera** 等产品仓如何复用 voxera-kit 的测试基建（Wave T）。  
运行时 kit 模块（storage/cache/mq 等）与测试 kit **分离**：产品只引入 harness、fake、fixture，不替代业务断言。

详细规划见 [`TESTING_INFRA_PLAN.md`](./TESTING_INFRA_PLAN.md)。

---

## 能力矩阵

| 包 | 路径 | 用途 | Docker |
|----|------|------|--------|
| **testkit** | `backend/testkit` | containers + contract smoke | integration 需要 |
| **fixture** | `backend/fixture` | ID/时间/HTTP builder、媒体样例 | 否 |
| **testing** | `frontend/packages/testing` | Vitest setup、MSW、React 辅助 | 否 |
| **fixture (FE)** | `frontend/packages/fixture` | 用户/session/API JSON 工厂 | 否 |
| **E2E 模板** | `templates/e2e-playwright` | Playwright 登录 smoke + CI 片段 | CI 可选 |

各数据平面模块另有 **memory fake** 与 **Run\*Contract**，单测优先用 fake，集成测用 testkit。

---

## CI 分层约定

| 层级 | 触发 | 内容 | kit 入口 |
|------|------|------|----------|
| **unit** | 每次 PR | 业务逻辑 + memory fake | 各模块 `*/memory` |
| **contract** | 每次 PR | 无 Docker 契约 | `go test ./.../contract/...` |
| **integration** | main / nightly | testcontainers | `go test -tags=integration ./testkit/...` |
| **E2E smoke** | main（可选） | Playwright 模板 | 复制 `templates/e2e-playwright` |

kit 自身 CI 参考：`.github/workflows/ci.yml`、`nightly.yml`。

---

## 后端：单测（无 Docker）

### 1. 固定 kit 版本

产品 CI 通过 sibling checkout + pin 文件锁定 kit：

```bash
# .github/voxera-kit-pin — v0.2.0
132286ea38c3a3300c43c72c140c8cc2dc34e984
```

CI 步骤（与 voxera/finera/pulsera 现有一致）：

```yaml
- run: |
    cd ../voxera-kit && git checkout "$(cat $GITHUB_WORKSPACE/.github/voxera-kit-pin)"
```

发版流程：kit tag → 各仓更新 pin → `go mod tidy` → CI 绿。见 [`COMPATIBILITY.md`](./COMPATIBILITY.md)。

### 2. go.work 纳入 kit 子模块

产品 `backend/go.work` 加入所需 kit 模块，例如：

```
use (
    ../voxera-kit/backend/storage
    ../voxera-kit/backend/cache
    ../voxera-kit/backend/testkit
    ../voxera-kit/backend/fixture
)
```

### 3. 用 memory fake 测 adapter 层

产品保留 domain 接口，adapter 委托 kit port；单测注入 fake：

```go
import (
    "testing"

    storagemem "github.com/EthanShen10086/voxera-kit/storage/memory"
    storagecontract "github.com/EthanShen10086/voxera-kit/storage/contract"
)

func TestMediaAdapter(t *testing.T) {
    store := storagemem.New()
    storagecontract.RunObjectStoreContract(t, store)
    // 再测产品 adapter 的 key 规范、元数据等
}
```

常用 contract runner：

| 模块 | 函数 |
|------|------|
| storage | `storage/contract.RunObjectStoreContract` |
| cache | `cache/contract.RunCacheContract` |
| mq | `mq/contract.RunPublisherContract` |
| database | `database/contract.RunDatabaseContract` |
| task | `task/contract.RunTaskContract` |
| secret | `secret/contract.RunSecretContract` |

### 4. fixture 造数

```go
import (
    "github.com/EthanShen10086/voxera-kit/fixture"
    mediafixture "github.com/EthanShen10086/voxera-kit/fixture/media"
)

req, _ := fixture.NewHTTPRequest("POST", "/v1/upload").
    WithBearerToken("test").
    WithJSONBody(map[string]string{"name": "demo"}).
    Build()

key := mediafixture.AudioObjectKey("user-1", "rec-1")
```

领域数据（行情、会话等）仍用产品 `dataprovider/stub`；跨模块通用造数走 `fixture`。

---

## 后端：集成测（Docker）

### 1. 本地运行

```bash
cd backend/testkit
go test -tags=integration -race -timeout=15m ./...
```

需要 Docker。containers 封装：`StartRedis`、`StartMinIO`、`StartNATS`、`StartPostgres`。

### 2. 产品仓 smoke 示例

在 `services/gateway/integration/smoke_test.go`（或等价路径）：

```go
//go:build integration

package integration_test

import (
    "testing"

    "github.com/EthanShen10086/voxera-kit/testkit/contract"
)

func TestDataPlaneSmoke(t *testing.T) {
    contract.RunDataPlaneSmoke(t)
}
```

PR 不跑 integration；main merge 后或 nightly 跑。build tag `integration` 避免 PR 依赖 Docker。

### 3. 单模块深度集成

```go
import (
    "context"
    "testing"

    "github.com/EthanShen10086/voxera-kit/testkit/containers"
    miniostore "github.com/EthanShen10086/voxera-kit/storage/minio"
)

func TestStorageWithMinIO(t *testing.T) {
    ctx := context.Background()
    ep, cleanup := containers.StartMinIO(ctx, t)
    defer cleanup()

    store, err := miniostore.New(/* Config from ep */)
    // ...
}
```

---

## 前端：Vitest + MSW

### 1. 依赖

```json
{
  "devDependencies": {
    "@voxera-kit/testing": "workspace:*"
  }
}
```

monorepo 通过 `go.work` / pnpm workspace 链到 kit `frontend/packages/testing`。

### 2. vitest.config.ts

```typescript
import { defineConfig, mergeConfig } from "vitest/config";
import { setupVitest } from "@voxera-kit/testing/vitest";

export default mergeConfig(
  setupVitest({ test: { environment: "jsdom" } }),
  defineConfig({ test: { include: ["src/**/*.test.ts"] } }),
);
```

### 3. MSW

```typescript
import { beforeAll, afterAll, afterEach } from "vitest";
import { createMockServer } from "@voxera-kit/testing/msw";

const server = createMockServer();
beforeAll(() => server.listen({ onUnhandledRequest: "error" }));
afterEach(() => server.resetHandlers());
afterAll(() => server.close());
```

内置 handler：`/api/auth/*`、分页列表、错误码。产品可 `server.use(...)` 覆盖。

### 4. React 组件测（可选）

```tsx
/** @vitest-environment jsdom */
import { renderWithProviders, registerReactTestingCleanup } from "@voxera-kit/testing/react";

registerReactTestingCleanup();
renderWithProviders(<MyPage />);
```

---

## E2E（Playwright）

复制 kit 模板到产品仓：

```bash
cp -R ../voxera-kit/templates/e2e-playwright ./e2e
```

含 `playwright.config.ts`、登录 Page Object、`.github/workflows` CI 片段。  
产品定制 baseURL、登录凭证（CI secret），不强制把用例放进 kit 主模块。

---

## 与 MsgGuard 的边界

MsgGuard **不依赖 voxera-kit**，保持自有测试栈。本文档面向 voxera / finera / pulsera 等 kit consumer。

---

## 推荐落地顺序

1. 更新 `.github/voxera-kit-pin` 到最新 kit tag
2. 选一个 service 加 memory fake + contract 单测（PR 必过）
3. 加 `integration` smoke（main 必过）
4. 前端接入 `@voxera-kit/testing`（若有 UI 包）
5. 按需复制 E2E 模板

---

## 相关文档

- [`backend/testkit/README.md`](../backend/testkit/README.md)
- [`backend/fixture/README.md`](../backend/fixture/README.md)
- [`frontend/packages/testing/README.md`](../frontend/packages/testing/README.md)
- [`TESTING_INFRA_PLAN.md`](./TESTING_INFRA_PLAN.md)
- [`COMPATIBILITY.md`](./COMPATIBILITY.md)
