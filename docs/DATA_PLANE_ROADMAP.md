# 数据平面路线图 (Data Plane Roadmap)

> 与早期实施计划（17 项 To-do）**一一对应**的完成度追踪。  
> 更新时机：数据平面相关 PR merge 后，同步勾选状态、补 `Evidence` 列与 `Last verified` 日期。

**Last verified:** 2026-06-18（Sprint 4.7 / `MIN_COVERAGE=80`）

---

## 状态图例

| 标记 | 含义 |
|------|------|
| ✅ **done** | Port + Adapter 已实现，有契约或集成测试，可在产品中选用 |
| 🟡 **partial** | 核心路径可用，但文档/生产验证/部分云特性仍缺 |
| 🔶 **stub** | 接口存在，实现偏占位或仅 httptest/离线路径 |
| ❌ **missing** | 未实现或不在本仓范围 |
| 📦 **product** | 能力在 kit，接入进度在各产品仓 |

与根 README 图例一致：`✅` 生产可用 · `🟡` Port+Adapter+契约 · `🔶` 部分 · `❌` 缺失。

---

## 总览（17 项）

| # | 计划项 | 状态 | 证据（代码 / 测试） | 文档入口 |
|---|--------|------|---------------------|----------|
| 1 | README/CHANGELOG 与数据平面文档、GCP 对标 | 🟡 | 根 README 模块索引；`CHANGELOG.md`；`backend/{storage,cache,mq,task}/README.md` | [README](../README.md)、本表 |
| 2 | `storage.Config` 扩展（PathStyle、SessionToken 等） | ✅ | `backend/storage/port.go` | [storage/README](../backend/storage/README.md) |
| 3 | W1 · `storage/minio` 全量 ObjectStore（minio-go） | ✅ | `backend/storage/minio/`；`-tags=integration` + testkit MinIO | [storage/README](../backend/storage/README.md) |
| 4 | W1 · `storage/s3` 全量 ObjectStore（aws-sdk-go-v2） | ✅ | `backend/storage/s3/`；gofakes3 契约 | 同上 |
| 5 | W1 · `memory`/`fs` fake + 契约 + MinIO CI | ✅ | `storage/memory`、`storage/fs`、`storage/contract/`；CI `go-integration` | [TESTING_INFRA_PLAN](./TESTING_INFRA_PLAN.md) |
| 6 | W1 · 分片上传 + `UploadLarge`（minio/s3） | ✅ | `MultipartUploader`、`storage/internal/uploadlarge/` | [storage/README](../backend/storage/README.md) |
| 7 | W1.5 · 版本 / 生命周期 / 桶通知 + MQ 桥 | 🟡 | `VersionedObjectStore`、`StorageAdmin` 各 adapter；**无独立 bucket→MQ bridge 包** | `task/README` 组合说明 |
| 8 | W2 · `cache/redis` + `cache/local` + 契约 | ✅ | `backend/cache/redis`、`local`、`cache/contract/` | [cache/README](../backend/cache/README.md) |
| 9 | W2 · `cache/memcached` + 多级缓存（`tiered`） | ✅ | `backend/cache/tiered/`；L1+L2 读穿/写穿；`cache/README` 用法 | [cache/README](../backend/cache/README.md) |
| 10 | W2 · `mq/nats` Pub/Sub + 契约 | ✅ | `backend/mq/nats/`；embedded nats-server 单测 | [mq/README](../backend/mq/README.md) |
| 11 | W2 · `mq/kafka` Pub/Sub | ✅ | `backend/mq/kafka/`；testcontainers 集成 | 同上 |
| 12 | W2 · `mq/rabbitmq` | ✅ | `backend/mq/rabbitmq/`；testcontainers 集成 | 同上 |
| 13 | W3 · `database` postgres / mysql / mongodb + 契约 | ✅ | `backend/database/{postgres,mysql,mongodb}/`；postgres integration | `database/contract/` |
| 14 | W4 · `task` 延迟队列 / Worker / 重试 / 幂等 / DLQ | ✅ | `task/memory`、`task/redis`、`task/redis/worker.go` | [task/README](../backend/task/README.md) |
| 15 | W5 · `secret` Vault/Env + AWS/GCP/Tencent SM | 🟡 | `secret/{env,vault,aws,gcp,tencent}/`；tencent/aws/gcp 有 httptest 契约 | 根 README `secret` 行 |
| 16 | W6 · `storage/oss` + `storage/cos` 原生 SDK | ✅ | `backend/storage/oss`、`cos`；httptest vendor mock | [storage/README](../backend/storage/README.md) |
| 17 | W7 · voxera / finera / pulsera 按需接入 | 📦 | 各产品仓 pin + 替换 stub；kit 提供 Port | [testing.md](./testing.md)、[PR_EXECUTION_CHECKLIST](./PR_EXECUTION_CHECKLIST.md) |

**进度摘要：** 数据平面 **kit 内** 约 **14/16 项 done**（不含 #17 产品接入）；#1、#7、#15 为 partial。

---

## 分项说明

### #1 文档与 GCP 对标

| 子项 | 状态 | 位置 |
|------|------|------|
| 根 README 模块索引 + 对标矩阵 | ✅ | [README.md](../README.md) |
| CHANGELOG 按 Sprint 记录 | ✅ | [CHANGELOG.md](../CHANGELOG.md) |
| storage / cache / mq / task 模块 README | ✅ | `backend/*/README.md` |
| 逐模块 GCP 特性对照表 | ❌ | 未单独成文；storage README 有 GCS/S3/MinIO/OSS/COS 选型 |
| Cookbook（场景配方） | ❌ | 用 [testing.md](./testing.md) + 模块 README 代替 |

**缺口：** 若需要「COS 某 API vs GCS」级对照，建议在对应 `backend/<module>/README.md` 增补，或新建 `docs/gcp-comparison/`（可选）。

---

### #2 `storage.Config`

已实现字段（节选）：`Endpoint`、`Bucket`、`Region`、`PathStyle`、`SessionToken`、`UseSSL`、`DisableSSLVerify`、`MultipartThreshold` 等 → `backend/storage/port.go`。

---

### #3–#6 W1 Storage 核心

```bash
cd backend/storage
go test ./... -race
go test -tags=integration ./minio/...   # 需 Docker
```

| 能力 | minio | s3 | memory/fs |
|------|-------|-----|-----------|
| CRUD / List / Exists | ✅ | ✅ | ✅ |
| 预签名 URL | ✅ | ✅ | ✅ |
| Multipart + UploadLarge | ✅ | ✅ | ✅ |
| 契约测试 | ✅ | ✅ | ✅ |

---

### #7 W1.5 版本 / 生命周期 / 桶通知

| 能力 | Port | 实现概况 |
|------|------|----------|
| 版本（Enable/Get/List/Delete/Restore） | `VersionedObjectStore` | minio/s3/memory 集成测覆盖；gofakes3 对部分 API 有限制 |
| 生命周期规则 | `StorageAdmin.Put/Get/DeleteLifecycleRules` | cos/oss mock + minio integration；s3 受 gofakes3 限制 |
| 桶通知配置 | `NotificationManager` | s3 支持 sqs/sns；minio/cos/oss 部分 stub |
| **事件 → MQ 桥** | — | **未实现独立模块**；产品侧监听 MQ 或自行 webhook；见 `task/README`「与 mq 组合」 |

---

### #8–#9 W2 Cache

```bash
cd backend/cache && go test ./... -race
```

| Adapter | SDK | 测试 |
|---------|-----|------|
| `redis` | go-redis/v9 | miniredis 单测 |
| `local` | ristretto | 单测 |
| `memcached` | gomemcache | 离线 + 不可达 host 单测 |
| `memory` | 内置 | 契约 |
| `tiered` | 组合任意 `Cache` | 契约 + 回填单测 |

**#9 多级缓存：** `cache/tiered` 实现 L1+L2 读穿/写穿；Memorystore 基准文档仍可选。

---

### #10–#12 W2 MQ

```bash
cd backend/mq && go test ./... -race
go test -tags=integration ./kafka/... ./rabbitmq/...  # 需 Docker
```

| Adapter | 契约 | 集成 |
|---------|------|------|
| `nats` | ✅ | embedded server |
| `kafka` | ✅ | testcontainers |
| `rabbitmq` | ✅ | testcontainers |
| `memory` | ✅ | 进程内 |

---

### #13 W3 Database

```bash
cd backend/database
go test ./... -race
go test -tags=integration ./postgres/...
```

| Adapter | 离线单测 | 集成契约 |
|---------|----------|----------|
| `postgres` | Ping/DSN | ✅ testcontainers |
| `mysql` | Ping/DSN/connect 失败路径 | — |
| `mongodb` | Ping/URI/connect 失败路径 | — |

---

### #14 W4 Task / Worker

对标 Cloud Tasks：延迟投递、取消、`IdempotencyKey`、`Retry`、`DeadLetterQueue`。

| 组件 | 路径 |
|------|------|
| Port | `backend/task/port.go` |
| memory 实现 + DLQ | `backend/task/memory/` |
| redis 队列 + Worker | `backend/task/redis/` |
| 契约 | `backend/task/contract/` |

详见 [task/README](../backend/task/README.md)。

---

### #15 W5 Secret

| Adapter | 状态 | 测试 |
|---------|------|------|
| `env` | ✅ | roundtrip |
| `vault` | ✅ | httptest KV v2 mock |
| `aws` | ✅ | httptest Secrets Manager |
| `gcp` | ✅ | bufconn gRPC fake |
| `tencent` | 🟡 | httptest 契约；生产 SDK 行为需云环境验证 |

---

### #16 W6 OSS / COS

原生 SDK 适配器 + `storage/internal/testfixture` 的 cosmock/ossmock（单测不依赖云账号）。

---

### #17 W7 产品接入（📦 不在本仓勾选）

| 产品 | kit 接入方式 | 跟踪 |
|------|--------------|------|
| voxera | `go.work` + `.github/voxera-kit-pin` | 各产品仓 CI |
| finera | 同上 | [PR_EXECUTION_CHECKLIST](./PR_EXECUTION_CHECKLIST.md) Phase 3 |
| pulsera | 同上 | 同上 |

产品侧任务：用真实 adapter 替换 stub、配置环境变量、跑 [testing.md](./testing.md) 中的分层测试。

---

## 相关文档索引

| 文档 | 用途 |
|------|------|
| [README.md](../README.md) | 全局架构、43 模块状态、企业级对标 |
| [backend/storage/README.md](../backend/storage/README.md) | 对象存储选型与测试命令 |
| [backend/cache/README.md](../backend/cache/README.md) | 缓存适配器 |
| [backend/mq/README.md](../backend/mq/README.md) | 消息队列适配器 |
| [backend/task/README.md](../backend/task/README.md) | 延迟任务 vs cron vs mq |
| [TESTING_INFRA_PLAN.md](./TESTING_INFRA_PLAN.md) | Wave T 测试基建 |
| [testing.md](./testing.md) | **产品仓**如何接 testkit / contract / integration |
| [COVERAGE_ROADMAP.md](./COVERAGE_ROADMAP.md) | 覆盖率阶梯（当前 CI **80%**） |
| [development.md](./development.md) | 本地 `make test` / lint |
| [adding-module.md](./adding-module.md) | 新增 Port/Adapter 流程 |
| [CHANGELOG.md](../CHANGELOG.md) | 按 Sprint 的交付记录 |

---

## 维护约定（随 commit 更新）

1. **改数据平面 adapter 或契约时**：更新上表对应行的 `状态` 与 `证据`。
2. **发版 / Sprint 结束时**：改 `Last verified` 日期；在 `CHANGELOG.md` 写条目；必要时同步根 README 模块表。
3. **产品完成 W7 某项接入时**：在产品仓 PR 描述中链接本文件对应行；kit 侧可将 #17 子项拆到各产品 README。
4. **新增计划项**：在「总览」表追加行，勿删历史行（改为 `cancelled` 并注明原因）。

### 快速验证命令

```bash
cd backend
./scripts/coverage.sh                    # 合并覆盖率 ≥80%
go test ./storage/... ./cache/... ./mq/... ./database/... ./task/... ./secret/... -race
go test -tags=integration ./testkit/... ./storage/minio/... ./mq/kafka/...  # 需 Docker
```
