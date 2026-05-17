export enum LogLevel {
  Debug = 0,
  Info = 1,
  Warn = 2,
  Error = 3,
  Fatal = 4,
}

export interface LogEntry {
  level: LogLevel;
  message: string;
  timestamp: number;
  context?: Record<string, unknown>;
  traceId?: string;
  spanId?: string;
}

export interface ILogger {
  debug(message: string, context?: Record<string, unknown>): void;
  info(message: string, context?: Record<string, unknown>): void;
  warn(message: string, context?: Record<string, unknown>): void;
  error(message: string, context?: Record<string, unknown>): void;
  fatal(message: string, context?: Record<string, unknown>): void;
  setLevel(level: LogLevel): void;
  child(context: Record<string, unknown>): ILogger;
}

export interface WebVitalsMetrics {
  fcp?: number;
  lcp?: number;
  fid?: number;
  cls?: number;
  ttfb?: number;
  inp?: number;
}

export interface IPerformanceMonitor {
  collectWebVitals(): Promise<WebVitalsMetrics>;
  markStart(name: string): void;
  markEnd(name: string): number;
  measure<T>(name: string, fn: () => T): T;
  getMetrics(): Record<string, number>;
}

export interface Breadcrumb {
  type: string;
  category: string;
  message: string;
  timestamp: number;
  data?: Record<string, unknown>;
}

export interface IErrorTracker {
  captureException(error: Error, context?: Record<string, unknown>): string;
  captureMessage(message: string, level?: LogLevel): string;
  setUser(user: { id: string; email?: string }): void;
  addBreadcrumb(breadcrumb: Breadcrumb): void;
}

export interface TracingConfig {
  serviceName: string;
  endpoint?: string;
  sampleRate?: number;
  propagateContextHeaders?: boolean;
}

export interface SpanContext {
  traceId: string;
  spanId: string;
  parentSpanId?: string;
  name: string;
  startTime: number;
  attributes: Record<string, string | number | boolean>;
}
