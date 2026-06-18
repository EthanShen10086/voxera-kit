# Changelog - voxera-kit

可插拔基础设施 SDK 的完整功能变更记录。  
此文件独立于 git log，用于更直观地追踪每次功能变更、架构决策和版本迭代。

---

## [Unreleased]

---

## [0.2.0] - 2026-06-13

数据平面 W1–W7 与测试基建 Wave T1–T6 首版打包；产品仓可 pin 本 tag 获取 storage/cache/mq/database/task/secret 实现与 testkit。

### Added
- **testkit (Wave T1)**：`backend/testkit/containers`（Redis/NATS/Postgres/MinIO testcontainers）+ `integration` smoke tests
- **CI**：`go-integration` job 运行 testkit 集成测试；`storage`（minio/s3/oss/cos/memory/fs）、`cache`（redis/local/memcached）、`mq`（nats/kafka/rabbitmq/memory）、`database`（postgres/mysql/mongodb）、`task`（memory/redis）、`secret`（vault/aws/gcp）
- **storage 高级能力**：分片上传、版本控制、生命周期、桶通知（port + memory/minio/s3 等）
- **契约测试**：`storage/contract`、`cache/contract`、`mq/contract`、`database/contract`、`task/contract`、`secret/contract`
- **文档**：`backend/storage/README.md`、数据层 README 状态诚实化；**`docs/testing.md`** 产品接入指南
- **testing (Wave T4)**：`@voxera-kit/testing`（Vitest setup、MSW handlers、renderWithProviders）
- **fixture (Wave T3)**：`backend/fixture`、`@voxera-kit/fixture`
- **E2E 模板 (Wave T5)**：`templates/e2e-playwright`（Playwright 登录 smoke + CI 片段）
- **CI (Wave T6)**：PR 跑契约测试；main 跑 integration + E2E smoke；合并 Go coverage + Codecov；`nightly.yml`

### Changed
- `cache/local.New` 现返回 `(*Adapter, error)`（ristretto 初始化）— **破坏性变更**，finera 等 consumer 需处理 error

### Known gaps (post-0.2.0)
- `secret/tencent` 仍为 stub
- task 高级能力（DLQ、幂等 key、JetStream）未纳入 port
- 覆盖率 80% 门禁未 enforce

---

## [1.0.0] - 2026-05-22

### Added
- **大模型集成模块 (`llm/`)**: 统一多 Provider LLM 接口
  - `llm.Provider` 接口 — Chat / ChatStream / Embed / ListModels
  - `llm.Router` — 多模型优先级路由、自动降级、成本感知选择
  - `llm/openai` — OpenAI GPT-4o/GPT-4/GPT-3.5 适配器（含流式 SSE）
  - `llm/deepseek` — DeepSeek V3/Coder/Chat 适配器（OpenAI 兼容格式）
  - `llm/qwen` — 通义千问 DashScope API 适配器
  - `llm/claude` — Anthropic Claude Messages API 适配器
  - `llm/hunyuan` — 腾讯混元 API 适配器
  - `llm/prompt` — 模板引擎 + 7 个预置模板 (Summarize/Translate/Analyze/QA/Sentiment/Classify/Extract)
  - `llm/token` — Token 估算 + 成本计算工具
- **AI 配额计费模块 (`aiquota/`)**: 用量管理与成本控制
  - 五级 Tier 体系: Free / Pro / Enterprise / VIP / Admin
  - 每日/每月 Token + 请求次数配额
  - 模型白名单（Free 用户只能用 DeepSeek）
  - 并发控制（AcquireConcurrency）
  - 管理员白名单（无限额）
  - 超额策略: Reject / Degrade / Queue / Notify
  - 成本报告（按模型/用户聚合）
  - `aiquota/memory` — 内存实现（自动日/月重置）
  - `aiquota/noop` — 无限放行
- **前端 AI SDK (`@voxera-kit/ai`)**:
  - `AIClient` — 流式 SSE 对话 + 重试 + AbortController
  - `createChatHook` / `createCompletionHook` / `createSummaryHook` — 框架无关 hooks
- **ASR Whisper 真实实现**: HTTP multipart 调用 OpenAI Whisper API
- **Translation OpenAI 真实实现**: 基于 chat completions 的翻译/检测

### AI 配额方案
| Tier | 每日 Token | 每月 Token | 每日请求 | 可用模型 |
|------|-----------|-----------|---------|---------|
| free | 10,000 | 100,000 | 20 | deepseek-chat, hunyuan-lite |
| pro | 500,000 | 5,000,000 | 500 | gpt-4o-mini, deepseek, qwen |
| enterprise | 5,000,000 | 50,000,000 | 5,000 | 全部模型 |
| vip/admin | 无限 | 无限 | 无限 | 全部模型 |

---

## [0.9.0] - 2026-05-22

### Added
- **产品分析模块 (`analytics/`)**: 用户行为追踪与数据分析引擎
  - `analytics.Collector` — 事件采集接口 (Track/TrackBatch/Identify/Alias)
  - `analytics.Querier` — 分析查询接口 (Funnel/Retention/Path/Events/Profile)
  - `analytics/engine` — 自研内存分析引擎（漏斗/留存/路径计算）
  - `analytics/posthog` — PostHog API 适配器
  - `analytics/noop` — 空操作适配器
- **A/B 实验模块 (`experiment/`)**: 实验分组与统计显著性分析
  - `experiment.Manager` — 实验生命周期管理接口
  - `experiment/memory` — 内存实现（SHA-256 确定性分桶 + z-test/t-test 显著性）
  - `experiment/posthog` — PostHog Experiments 适配器
  - `experiment/noop` — 空操作适配器
- **前端行为采集 SDK (`@voxera-kit/analytics`)**:
  - `AnalyticsClient` — 批量缓冲 + 自动上下文丰富
  - `AutoTracker` — 页面浏览/点击/滚动深度/表单/外链/停留时长自动采集
  - `SessionManager` — 30 分钟不活跃自动轮转
  - `AttributionTracker` — UTM 首次/末次触达归因
  - `HttpProvider` / `PostHogProvider` — 可插拔后端
- **前端实验客户端 (`@voxera-kit/experiment`)**:
  - `ExperimentClient` — 分组缓存 + localStorage 持久化
  - `createExperimentHook` / `createVariantHook` — 框架无关 hooks
  - `HttpProvider` / `PostHogProvider` — 可插拔后端

### 产品分析能力矩阵
| 能力 | Kit 模块 | 对标产品 |
|------|----------|----------|
| 行为追踪 | analytics/ + @voxera-kit/analytics | Amplitude, Mixpanel |
| 漏斗分析 | analytics.QueryFunnel | Amplitude Funnels |
| 留存分析 | analytics.QueryRetention | PostHog Lifecycle |
| 路径分析 | analytics.QueryPath | Amplitude Pathfinder |
| A/B 实验 | experiment/ + @voxera-kit/experiment | LaunchDarkly, Statsig |
| 渠道归因 | AttributionTracker (UTM) | AppsFlyer, Branch |

---

## [0.8.0] - 2026-05-22

### Added
- **广告变现模块 (`ad/`)**: 可插拔广告 Provider 架构
  - `ad.Provider` / `ad.Router` / `ad.EventTracker` 接口
  - `ad/google` — Google Ads 适配器（桩）
  - `ad/selfhosted` — 本地库存加权随机适配器
  - `ad/noop` — 空操作适配器（付费用户/禁用时使用）
  - `DefaultRouter` — 优先级路由、付费/未成年人策略、Fallback
- **部署授权模块 (`license/`)**: 私有部署许可证管理
  - `license.Manager` 接口
  - `license/offline` — RSA 离线签名验证
  - `license/online` — 远程许可证服务器验证
- **前端广告包 (`@voxera-kit/ad`)**: 浏览器端广告 SDK
  - `AdRouter` 优先级调度
  - `AdTracker` 批量上报（keepalive）
  - `IAdProvider` 接口
- **PRINCIPLES.md**: 9 大核心设计原则文档
- **变现三路径**: SaaS 订阅 + 私有部署授权 + 广告变现

### Changed
- Gateway 路由配置补全 (subscriptions/usage/callbacks/scheduler/shorturl/share/im)

---

## [0.7.0] - 2026-05-22

### 定时任务 + 文档体系完善

**Commit**: `aa8b8a2`

#### 定时任务 (Cron) 真实实现
- `scheduler/cron/` — 基于 robfig/cron/v3 的生产级 cron 调度器
  - 支持标准 cron 表达式（`0 */5 * * *`）+ 扩展语法（`@every 30s`、`@hourly`）
  - 信号量并发控制 + panic 恢复 + Location 时区
- `scheduler/memory/` — 修复为可用的轻量 fallback（解析 `@every` 格式）
- 对标：Kubernetes CronJob / Quartz / Airflow

#### 文档
- README 全面重写（304 行）：Mermaid 架构图、38 模块索引、企业级特性对标矩阵
- 企业级特性对标表：24 项能力 vs Google/Netflix/Alibaba/AWS 方案

---

## [0.6.0] - 2026-05-22

### 企业级基础设施全面加固

**Commit**: `52e1332`  
**分支**: `master`  
**变更类型**: feat (重大功能增强)

#### Phase 1: 可观测性落地（真实实现替代 stub）

| 模块 | 变更 |
|------|------|
| `observability/logger/zap.go` | 真实 ZapLogger 实现（functional options, 结构化 Field 转换, WithTraceID） |
| `observability/logger/slog.go` | 新增 SlogAdapter（标准库 slog.Handler 适配） |
| `observability/tracing/otel.go` | 真实 OTelTracer（OTLP HTTP exporter, BatchSpanProcessor, 采样率） |
| `observability/metrics/prometheus.go` | 真实 PrometheusRecorder（sync.Map 缓存 metric vectors） |
| `observability/metrics/handler.go` | 新增 HTTPHandler() 返回 promhttp.Handler |
| `observability/profiling/pprof.go` | 新增 pprof 端点注册（可配置开关） |

#### Phase 2: 性能优化模块（全新）

| 模块 | 接口 | 适配器 |
|------|------|--------|
| `singleflight/` | `Deduplicator` | `sync/adapter.go`（封装 x/sync/singleflight） |
| `retry/` | `Retrier` + `Policy` | `exponential/adapter.go`（退避+jitter+crypto/rand） |
| `bulkhead/` | `Bulkhead` + `ErrBulkheadFull` | `semaphore/adapter.go`（buffered channel） |
| `loadshed/` | `Shedder` + `Token` + `ErrOverloaded` | `adaptive/adapter.go`（AIMD 算法） |

#### Phase 3: 安全合规模块（全新）

| 模块 | 接口 | 适配器 |
|------|------|--------|
| `audit/` | `Writer` + `Reader` + `Entry` + `Filter` | memory + noop |
| `secret/` | `Manager` + `ErrNotFound` | env adapter（前缀环境变量） |
| `pii/` | `Redactor` + `Rule` | regex adapter（email/phone/CC/SSN/IP） |
| `featureflag/` | `Store` + `Flag` + `EvalContext` | memory（SHA-256 确定性百分比） |
| `security/headers/` | `Config` + `HSTSConfig` | `DefaultStrict()` / `DefaultPermissive()` |
| `crypto/tls/` | `Config` | `NewServerTLSConfig` / `NewClientTLSConfig`（mTLS） |

#### Phase 4: 可复用 HTTP 中间件模块（全新）

`middleware/` 模块提供 12 个开箱即用的 `net/http` 中间件：

| 中间件 | 功能 |
|--------|------|
| `Chain` | 中间件组合器 |
| `RequestID` | crypto/rand X-Request-ID 生成/传播 |
| `Logging` | 结构化请求日志（via kit Logger） |
| `Tracing` | OTel span + X-Trace-ID 传播 |
| `Metrics` | RED 指标（请求数/延迟/在途） |
| `Audit` | 审计日志（mutating + 敏感路径） |
| `Recovery` | panic 恢复 + 栈追踪 |
| `HealthCheck` | /health + /ready（依赖探测） |
| `Timeout` | 请求级超时 |
| `LoadShed` | 过载保护（503 + Retry-After） |
| `SecurityHeaders` | HSTS/CSP/X-Frame 等 |
| `PIIRedact` | 错误响应 PII 脱敏 |

#### Phase 5: 前端可观测性增强

| 能力 | 变更 |
|------|------|
| Web Vitals | 完整 LCP/CLS/INP/FID（PerformanceObserver） |
| Long Task | longtask observer（>50ms） |
| Resource Timing | 慢资源检测（可配置阈值） |
| TracingClient | 真实 OTLP JSON batch export via fetch |
| ErrorTracker | 远程上报（采样率 + beforeSend + debounce） |
| AuditClient | 用户操作审计（batch + visibilitychange flush） |

#### Phase 7: 部署配套

| 文件 | 说明 |
|------|------|
| `deploy/otel-collector.yaml` | OTLP 接收 → Jaeger + Prometheus 导出 |
| `deploy/prometheus.yml` | 3 组服务抓取配置 |
| `deploy/alertmanager.yml` | 告警路由 |
| `deploy/alert-rules.yml` | 5 条告警规则 |
| `deploy/docker-compose.observability.yml` | 一键本地可观测性栈 |
| `deploy/grafana/` | Dashboard + Provisioning 配置 |

---

## [0.1.0] - 2026-05-17

### 🏗️ 初始化 - 可插拔基础设施 SDK 脚手架

**Commit**: `7ee65ba`  
**分支**: `master`  
**变更类型**: feat (新功能)

#### 前端 SDK (TypeScript Monorepo)

| 包名 | 状态 | 说明 |
|------|------|------|
| `@voxera-kit/di` | ✅ 完成 | IoC/DI 容器，Map-based，支持 Singleton/Transient 生命周期，层级子容器 |
| `@voxera-kit/plugin` | ✅ 完成 | 插件系统，生命周期管理 (onInit/onMount/onDestroy)，依赖拓扑排序 |
| `@voxera-kit/player` | ✅ 完成 | 播放器抽象层 `IPlayerAdapter`，适配器: xgplayer / video.js / hls.js |
| `@voxera-kit/theme` | ✅ 完成 | CSS Variables 主题引擎，预设: light / dark / high-contrast，Design Tokens 定义 |
| `@voxera-kit/i18n` | ✅ 完成 | 框架无关国际化引擎，支持嵌套键、{param} 插值、locale 切换回调 |
| `@voxera-kit/federation` | ✅ 完成 | Module Federation 工具，`IRemoteModule` 契约，host 加载器 + remote 定义器 |
| `@voxera-kit/api-client` | ✅ 完成 | 类型安全 HTTP/WebSocket 客户端，拦截器链、自动重试、超时控制 |
| `@voxera-kit/observability` | ✅ 完成 | 前端可观测性: 结构化日志(批量上报)、Web Vitals、OTel 追踪、错误追踪 |
| `@voxera-kit/config` | ✅ 完成 | 共享 TSConfig (strict) / ESLint (flat config) / Prettier 配置 |

**前端架构决策**:
- Turborepo + pnpm workspace 管理 monorepo
- 所有配置文件使用 `.ts` 格式，零 JS
- TypeScript strict 模式，目标 ES2022，`"type": "module"`
- 每个包独立 `package.json` + `tsconfig.json`

#### 后端 Go Modules (Port + Adapter 模式)

| 模块 | Port (接口) | Adapter (实现) | 状态 |
|------|------------|----------------|------|
| `database` | `Repository[T]`, `Transaction`, `Database` | PostgreSQL, MongoDB, MySQL | ✅ 完成 |
| `cache` | `Cache` (Get/Set/Delete/SetWithTTL) | Redis, Memcached, Local (ristretto) | ✅ 完成 |
| `mq` | `Publisher`, `Subscriber`, `Message` | NATS, Kafka, RabbitMQ | ✅ 完成 |
| `storage` | `ObjectStore` (Upload/Download/Delete/GetURL) | S3, MinIO, 阿里云 OSS | ✅ 完成 |
| `payment` | `PaymentGateway` (CreateOrder/QueryOrder/Refund/HandleCallback) | Stripe, 微信支付, 支付宝, PayPal | ✅ 完成 |
| `asr` | `Recognizer` (Recognize/RecognizeStream), `Segment` | OpenAI Whisper, Azure Speech, 阿里云, 自部署 | ✅ 完成 |
| `translation` | `Translator` (Translate/TranslateBatch) | OpenAI, DeepL, Google Translate | ✅ 完成 |
| `auth` | `Authenticator`, `Authorizer` | JWT, OAuth2, OIDC | ✅ 完成 |
| `observability` | `Logger`, `MetricsRecorder`, `Tracer` | Zap, Prometheus, OpenTelemetry | ✅ 完成 |
| `framework` | `Server`, `HTTPServer`, `RPCServer` | go-kratos (HTTP/gRPC), CloudWeGo (Hertz/Kitex) | ✅ 完成 |
| `config` | 配置加载接口 | Viper-based | ✅ 完成 |
| `errors` | `AppError`, `ErrorCode` | 统一错误类型 + 常用错误码 | ✅ 完成 |

**后端架构决策**:
- Go 1.22+ 特性，Go workspace 管理多模块
- 六边形架构 (Hexagonal / Ports & Adapters)
- 每个模块独立 `go.mod`，可单独引用
- 模块路径: `github.com/EthanShen10086/voxera-kit/{module}`
- 适配器为 stub 实现，包含完整接口签名和 TODO 标注

#### 基础设施

- `.github/workflows/ci.yml` - PR/push 触发 lint + test (前后端)
- `.github/workflows/release.yml` - semantic-release 自动发版
- `.vscode/settings.json` + `extensions.json` - TypeScript + Go 开发配置
- `.editorconfig` - 编辑器统一配置
- `LICENSE` - MIT 开源协议

---

## [0.3.0] - 2026-05-17

### Finera 驱动的新增模块 -- 数据分析能力扩展

为支持 Finera 金融数据分析平台，kit 新增 5 个可插拔模块：

#### 前端新增 (3 个 TS 包)

- **@voxera-kit/spreadsheet**: 电子表格引擎抽象
  - `ISpreadsheetAdapter` 接口 (mount/destroy/setData/getData/setCellValue/exportXlsx 等)
  - `UniverAdapter` 适配器 (Univer Canvas2D 电子表格)
  - `AGGridAdapter` 适配器 (AG Grid 数据网格)
  - `SpreadsheetFactory` 工厂方法
  - 支持条件格式、冻结窗格、XLSX 导入导出

- **@voxera-kit/chart**: 图表引擎抽象
  - `IChartAdapter` 接口 (mount/destroy/render/resize/exportImage)
  - 7 种图表类型: Line, Bar, Pie, Sankey, Area, Waterfall, Radar
  - `EChartsAdapter` 适配器 (百度 ECharts)
  - `VisxAdapter` 适配器 (Airbnb visx)
  - `ChartFactory` 工厂方法

- **@voxera-kit/clipboard**: 剪贴板表格解析
  - `IClipboardParser` 接口 (parse/parseText/parseHtml/detectSource)
  - 自动识别来源: Excel, Google Sheets, 飞书, 腾讯文档, 钉钉
  - 智能类型推断: 数字、百分比、日期、布尔值
  - 支持会计格式负数 `(1,234)` 解析

#### 后端新增 (2 个 Go 模块)

- **dataprovider**: 数据源抽象
  - `DataProvider` 通用数据源接口
  - `FinancialProvider` 金融数据特化接口 (GetQuote/GetFinancials/ListMarkets)
  - `StubAdapter` 内存模拟适配器 (含 A 股/港股/美股 mock 数据)
  - 支持三大市场: US, HK, SH/SZ

- **dataparser**: 文档解析抽象
  - `DocumentParser` 接口 (Parse/ExtractTables/SupportedFormats)
  - `CSVAdapter` 完整 CSV 解析实现 (encoding/csv)
  - `StubAdapter` 测试用模拟适配器
  - 支持格式: PDF, CSV, XLSX, HTML, JSON

**Go 模块总数**: 25 个 (新增 2 个)
**前端包总数**: 15 个 (新增 3 个)

---

## 待办 (Planned)

### [0.4.0] - 计划中
- [ ] 完善所有 Go adapter 的真实实现 (连接池、错误处理、重试逻辑)
- [ ] 添加单元测试 (目标覆盖率 80%+)
- [ ] 添加集成测试 (使用 testcontainers-go)
- [ ] 发布到 npm + Go proxy
- [ ] 添加 Benchmark 性能测试

### [0.3.0] - 计划中
- [ ] 添加更多 ASR 适配器 (百度、讯飞)
- [ ] 添加更多支付适配器 (Apple Pay、Google Pay)
- [ ] framework 模块添加 go-zero 适配器
- [ ] 前端包添加 Vitest 单元测试

---

## [0.4.0] - 2026-05-17

### 🔧 DevOps / DX 工程化全面加固

**变更类型**: chore, ci, docs

#### 新增配置

| 文件 | 说明 |
|------|------|
| `backend/.golangci.yml` | golangci-lint 共享配置，启用 15 个 linter，0-warning 策略 |
| `frontend/.prettierrc.json` | Prettier 根配置 (singleQuote, trailingComma, 100 字符) |
| `commitlint.config.ts` | Conventional Commits 规范校验 |
| `.husky/pre-commit` | lint-staged (TS) + gofmt 检查 (Go) |
| `.husky/commit-msg` | commitlint 校验提交信息 |
| `Makefile` | 12 个 Make 目标: lint/fmt/test/build/vet/typecheck/clean |
| `package.json` (root) | husky + commitlint 依赖 |

#### CI/CD 完善

| Workflow | 触发 | 内容 |
|----------|------|------|
| `ci.yml` | push/PR | Go lint+vet+build+test (25 模块) + TS lint+typecheck (15 包) |
| `reusable-go-ci.yml` | workflow_call | Go 可复用模板 (供 voxera/finera 调用) |
| `reusable-ts-ci.yml` | workflow_call | TS 可复用模板 |
| `reusable-docker-build.yml` | workflow_call | Docker build+push 可复用模板 |

#### 文档体系

| 文档 | 内容 |
|------|------|
| `CONTRIBUTING.md` | 贡献指南: 环境要求, 代码风格, commit 规范, PR 流程 |
| `docs/architecture.md` | 模块关系图, 接口设计原则, 适配器模式 |
| `docs/development.md` | 本地开发环境搭建, 调试方法 |
| `docs/adding-module.md` | 如何添加新的 kit 模块 (Go + TS) |
| `.github/pull_request_template.md` | PR 模板 |

#### DX 增强

- `.editorconfig` 增加 TS/Go/YAML/Makefile 分段规则
- `.vscode/settings.json` 增加 rulers=[100], goimports, eslint validate
- `.vscode/extensions.json` 补齐推荐扩展
- `frontend/package.json` 新增 `format` 脚本

---

## [Unreleased]

**变更摘要**（自动生成）

- （相对上一版本无有效提交，或均为工程类提交）

## [0.5.0] - 2026-05-22

### 新增模块 — Pulsera 支撑

**变更类型**: feat

#### backend/scraper — 抓取引擎抽象

| 接口/类型 | 说明 |
|-----------|------|
| `Fetcher` | HTTP 请求抽象（代理轮换、UA 池、重试退避） |
| `Parser` | 内容解析（HTML/JSON → Post） |
| `ProxyPool` | 代理池管理（添加/移除/健康检查/轮换） |
| `RateLimiter` | 平台级限流策略 |
| `Tracker` | 定时追踪编排器 |
| `Post` | 统一帖子模型 |

适配器: `http/`（标准库 + 代理 + UA 随机化）、`memory/`（内存代理池）

#### backend/notification — 通知渠道抽象

| 接口/类型 | 说明 |
|-----------|------|
| `Notifier` | 通知发送接口 |
| `Router` | 多渠道消息路由 |
| `Message` | 统一消息模型 |
| `DeliveryResult` | 投递结果 |

适配器: `wecom/`（企业微信 Webhook）、`feishu/`（飞书 Webhook）、`email/`（SMTP）、`stub/`（测试桩）

#### frontend/packages/feed — Feed 时间线抽象

| 接口 | 说明 |
|------|------|
| `IFeedAdapter` | Feed 数据源接口 |
| `IFeedRenderer` | Feed 渲染器接口 |
| `FeedItem` | 统一 Feed 项模型 |
