# Changelog - voxera-kit

可插拔基础设施 SDK 的完整功能变更记录。  
此文件独立于 git log，用于更直观地追踪每次功能变更、架构决策和版本迭代。

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

## 待办 (Planned)

### [0.2.0] - 计划中
- [ ] 完善所有 Go adapter 的真实实现 (连接池、错误处理、重试逻辑)
- [ ] 添加单元测试 (目标覆盖率 80%+)
- [ ] 添加集成测试 (使用 testcontainers-go)
- [ ] 发布到 npm + Go proxy
- [ ] 添加 Benchmark 性能测试

## [0.2.0] - 2026-05-17

### 🔧 Feature Gap Fill — 全面补齐可插拔基础设施接口

#### 后端新增 Go 模块 (Phase 1)

| 模块 | 端口接口 | 内置适配器 | 说明 |
|------|----------|-----------|------|
| `ratelimiter` | `RateLimiter` | `MemoryRateLimiter` (滑动窗口) | 请求限流，支持按 key 配置 |
| `security` | `IPFilter` | `MemoryIPFilter` | IP 黑白名单过滤，whitelist/blacklist/both 模式 |
| `concurrency` | `Semaphore`, `WorkerPool` | Channel-based | 并发控制，信号量 + 工作池 |
| `circuitbreaker` | `CircuitBreaker` | `MemoryCircuitBreaker` | 熔断器，closed/open/half-open 状态机 |
| `crypto` | `Encryptor`, `Hasher`, `Signer` | AES-GCM + bcrypt | 加解密/哈希/签名统一抽象 |
| `compression` | `Compressor` | `GzipCompressor` | HTTP 压缩/解压 |
| `scheduler` | `Scheduler`, `Job` | `MemoryScheduler` | 定时任务调度 |
| `shorturl` | `ShortURLGenerator` | `MemoryShortURL` (base62) | 短链生成与解析 |
| `messaging` | `Channel`, `Presence` | `MemoryChannel` | IM 消息通道抽象 |
| `registry` | `ServiceRegistry`, `ConfigCenter` | `MemoryRegistry` | 服务注册/发现 + 配置中心 |

#### 前端新增 Node.js 框架抽象 (Phase 2)

| 包名 | 说明 |
|------|------|
| `@voxera-kit/server` | Node.js HTTP 服务器抽象: `IHttpServer`/`Router`/`Middleware` 接口 |
| — `RawHttpServer` | 基于 `node:http` 的完整实现 |
| — Koa/Express/Fastify stubs | 可插拔框架适配器存根 |
| — 7 个内置中间件 | cors, bodyParser, requestId, responseTime, helmet, compress, rateLimit |

#### 后端 ORM 扩展 (Phase 3)

- `database` 模块新增: `QueryBuilder[T]` 泛型查询构建器
- `QueryCondition`, `SortOrder`, `OrderBy`, `Pagination` 类型
- `Migration` + `Migrator` 接口 (版本化数据库迁移)
- `DBCluster` + `DBClusterConfig` (主从读写分离)

#### 前端增强 (Phase 4)

| 包名 | 新增功能 |
|------|----------|
| `@voxera-kit/cache` | **新包** — `ICache` 接口 + `MemoryCache` (LRU 淘汰策略) |
| `@voxera-kit/seo` | **新包** — `SEOManager` (meta/OG/twitter) + `SitemapGenerator` (XML) |
| `@voxera-kit/api-client` | 新增 `SSEClient` (EventSource + 自动重连 + 指数退避) |
| `@voxera-kit/di` | Container 新增 `registerClass()` 装饰器自动装配 |
| `@voxera-kit/observability` | `TracingClient.activeSpan()` + `FrontendLogger.withTracing()` TraceId 自动注入 |

**编译验证**: 22 个 Go 模块 `go build` + `go vet` ✅ | 11 个 TS 包 `tsc --noEmit` ✅ | `gofmt` 零问题 ✅

---

### [0.3.0] - 计划中
- [ ] 添加更多 ASR 适配器 (百度、讯飞)
- [ ] 添加更多支付适配器 (Apple Pay、Google Pay)
- [ ] framework 模块添加 go-zero 适配器
- [ ] 前端包添加 Vitest 单元测试
