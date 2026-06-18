# testkit

Wave T1 集成测试基建：testcontainers 封装 + 数据平面 smoke tests。

## containers

| 函数 | 镜像 | 用途 |
|------|------|------|
| `StartRedis` | redis:7-alpine | cache/redis 契约 |
| `StartNATS` | nats:2.10-alpine | mq/nats 契约 |
| `StartPostgres` | postgres:16-alpine | database/postgres Ping |
| `StartMinIO` | minio/minio | storage/minio 契约 |

## 运行

```bash
# 单元（无外部依赖）
cd backend/testkit && go test ./containers/...

# 集成（需要 Docker）
cd backend/testkit && go test -tags=integration -race -timeout=15m ./...
```

CI job：`go-integration`（testkit）。
