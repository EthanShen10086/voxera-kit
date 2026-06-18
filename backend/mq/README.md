# mq

消息队列 Port（对标 Pub/Sub）。

## 状态

| Adapter | 状态 |
|---------|------|
| `nats` | 🟡 |
| `kafka` | 🟡 segmentio/kafka-go |
| `rabbitmq` | 🟡 |
| `memory` | ✅ 进程内 bus（单测/本地） |

## 测试

```bash
cd backend/mq && go test ./... -race
```

本地开发无 broker 时使用 `mq/memory`；生产配置 `KAFKA_BROKERS` / `NATS_URL` 等。
