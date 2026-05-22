# 设计原则 (Design Principles)

> 本文档定义了 voxera-kit 及其派生应用（Voxera/Finera/Pulsera）在架构设计与代码实现中必须遵循的核心原则。
> 每次代码迭代都应回顾这些原则，确保软件演进方向的一致性。

## 1. 可插拔架构 (Pluggable Architecture)

- **六边形架构 (Hexagonal / Ports & Adapters)**：Kit 只定义 Port（接口），Adapter（适配器）携带外部依赖
- **依赖倒置**：应用层依赖抽象而非具体实现
- **适配器可替换**：任何适配器都可以在不修改业务逻辑的前提下替换
- **渐进式采用**：每个模块独立 go.mod，可单独引入

## 2. 云原生 (Cloud Native)

- **12-Factor App**：配置与代码分离、无状态进程、端口绑定、日志流
- **容器优先**：所有服务提供 Dockerfile，支持 K8s 部署
- **声明式基础设施**：Helm Chart / docker-compose 描述期望状态
- **优雅关停**：SIGTERM 处理、连接排空、资源释放

## 3. 微服务 (Microservices)

- **单一职责**：每个服务只做一件事，通过 API 边界协作
- **独立部署**：服务可独立构建、测试、发布
- **API 网关**：统一入口，路由/鉴权/限流/追踪在 Gateway 层完成
- **服务间通信**：REST (同步) + MQ (异步)，追踪 ID 全链路传播

## 4. 可观测性 (Observability)

- **三大支柱**：Logging (结构化 JSON) + Tracing (OpenTelemetry) + Metrics (Prometheus RED)
- **TraceID 全链路**：从前端到数据库，每个请求可追溯
- **性能分析按需开启**：pprof 端点通过配置控制
- **审计日志**：所有变更操作记录 who/what/when/where

## 5. 安全合规 (Security & Compliance)

- **零信任**：mTLS 服务间通信，JWT 用户认证
- **最小权限**：RBAC + 特性开关控制功能可见性
- **数据保护**：PII 脱敏、加密存储、GDPR 合规
- **未成年人保护**：MinorMode 全局开关，限制内容/时长/支付

## 6. 性能优先 (Performance First)

- **过载保护**：Load Shedding (AIMD) + Circuit Breaker + Rate Limiter
- **并发隔离**：Bulkhead 防止级联故障
- **请求去重**：Singleflight 避免缓存击穿
- **智能重试**：指数退避 + 抖动，尊重 context 超时
- **异步优先**：非关键路径走 MQ，不阻塞用户请求

## 7. 代码质量 (Code Quality)

- **严格 Lint**：golangci-lint（revive/misspell/gosec/errcheck），TypeScript strict mode
- **不降规则修代码**：代码适应规则，而非规则适应代码
- **命名不重复**：避免 stutter（`pkg.PkgType` → `pkg.Type`）
- **US English**：注释、变量、常量统一美式英语
- **文档即代码**：每个 exported symbol 必须有 doc comment

## 8. 变现多元化 (Monetization)

- **SaaS 订阅**：Free/Pro/Enterprise 三档，配额限制
- **私有部署授权**：License 离线验证 (RSA) + 在线心跳
- **广告变现**：可插拔广告 Provider，付费用户/未成年人自动隐藏
- **可插拔支付**：Stripe/微信/支付宝/PayPal 四渠道，Kit 定义接口

## 9. 文档与可追溯 (Documentation & Traceability)

- **CHANGELOG**：每次功能迭代更新，独立于 git log
- **README**：反映当前真实架构，含对标矩阵
- **Conventional Commits**：feat/fix/docs/refactor 前缀
- **Git Tag**：语义化版本，每次发布打 tag

## 10. AI-First (AI 原生)

- **统一 Provider 接口**: 所有 LLM 调用通过 `llm.Provider` 接口，一行代码切换模型
- **成本可控**: `aiquota` 模块强制配额，防止 AI 费用失控
- **流式优先**: 所有对话默认 SSE 流式，提升用户体验
- **优雅降级**: 高级模型不可用时自动降级到廉价模型
- **Prompt 工程**: 预置模板标准化常见任务，避免 prompt 散落各处
- **白名单/VIP**: 管理员和特邀用户绕过配额，保障核心用户体验
