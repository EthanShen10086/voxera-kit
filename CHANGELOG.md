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
