export {
  LogLevel,
} from "./types.js";

export type {
  LogEntry,
  ILogger,
  WebVitalsMetrics,
  IPerformanceMonitor,
  Breadcrumb,
  IErrorTracker,
  TracingConfig,
  SpanContext,
} from "./types.js";

export { FrontendLogger } from "./logger.js";
export type { FrontendLoggerOptions } from "./logger.js";

export { PerformanceMonitor } from "./performance.js";
export type { ObserveConfig, SlowResource, LongTaskEntry } from "./performance.js";

export { TracingClient } from "./tracing.js";
export type { ExportConfig } from "./tracing.js";

export { ErrorTracker } from "./error.js";
export type {
  ErrorTrackerOptions,
  CapturedEntry,
  ReportConfig,
  ErrorEvent,
  UserContext,
} from "./error.js";

export { AuditClient } from "./audit.js";
export type { AuditAction, AuditClientConfig } from "./audit.js";
