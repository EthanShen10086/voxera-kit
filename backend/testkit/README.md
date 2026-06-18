# testkit

Wave T 集成测试基建：testcontainers 封装 + 数据平面 contract smoke tests。

## 结构

| 包 | 职责 |
|----|------|
| `containers/` | testcontainers 封装：Redis、NATS、Postgres、MinIO |
| `contract/` | 重导出各模块 `Run*Contract`；`RunDataPlaneSmoke`（integration tag） |
| `assert/` | 共享断言（`ErrorIs` 等） |
| `integration/` | CI 集成测试入口 |
| `assert/` | 共享断言（`ErrorIs` 等） |

相关：`backend/fixture`（Wave T3 造数）、`frontend/packages/fixture`（前端 JSON 工厂）、`frontend/packages/testing`（Wave T4 Vitest/MSW/React 测试基建）。

## containers

| 函数 | 镜像 | 用途 |
|------|------|------|
| `StartRedis` | redis:7-alpine | cache/redis 契约 |
| `StartNATS` | nats:2.10-alpine | mq/nats 契约 |
| `StartPostgres` | postgres:16-alpine | database/postgres 契约 |
| `StartMinIO` | minio/minio | storage/minio + s3 兼容契约 |

## 运行

```bash
# 单元（无 Docker）
cd backend/testkit && go test ./...

# 各模块契约（无 Docker）
cd backend/database && go test ./contract/...
cd backend/task && go test ./contract/...
cd backend/secret && go test ./contract/...

# 集成（需要 Docker）
cd backend/testkit && go test -tags=integration -race -timeout=15m ./...
```

CI job：`go-integration`（testkit，仅 main/master push）；`go-contract`（PR）；`e2e-smoke`（Playwright 模板，main/master）；`nightly` workflow。
