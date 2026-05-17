import { LogLevel } from "./types.js";
import type { ILogger, LogEntry } from "./types.js";
import type { TracingClient } from "./tracing.js";

export interface FrontendLoggerOptions {
  level?: LogLevel;
  context?: Record<string, unknown>;
  /** Batch size threshold – flush when the buffer reaches this count. */
  batchSize?: number;
  /** Interval in ms to auto-flush buffered logs. */
  flushInterval?: number;
  /** Callback invoked on flush with the accumulated entries. */
  onFlush?: (entries: LogEntry[]) => void | Promise<void>;
  /** Print to console in addition to buffering (typically enabled in dev). */
  console?: boolean;
  tracer?: TracingClient;
}

const LOG_LEVEL_LABELS: Record<LogLevel, string> = {
  [LogLevel.Debug]: "DEBUG",
  [LogLevel.Info]: "INFO",
  [LogLevel.Warn]: "WARN",
  [LogLevel.Error]: "ERROR",
  [LogLevel.Fatal]: "FATAL",
};

export class FrontendLogger implements ILogger {
  private level: LogLevel;
  private readonly context: Record<string, unknown>;
  private readonly buffer: LogEntry[] = [];
  private readonly batchSize: number;
  private readonly onFlush?: (entries: LogEntry[]) => void | Promise<void>;
  private readonly useConsole: boolean;
  private readonly tracer?: TracingClient;
  private flushTimer: ReturnType<typeof setInterval> | null = null;

  constructor(options: FrontendLoggerOptions = {}) {
    this.level = options.level ?? LogLevel.Debug;
    this.context = options.context ?? {};
    this.batchSize = options.batchSize ?? 50;
    this.onFlush = options.onFlush;
    this.useConsole = options.console ?? true;
    this.tracer = options.tracer;

    if (options.flushInterval && options.flushInterval > 0) {
      this.flushTimer = setInterval(() => {
        void this.flush();
      }, options.flushInterval);
    }
  }

  debug(message: string, context?: Record<string, unknown>): void {
    this.log(LogLevel.Debug, message, context);
  }

  info(message: string, context?: Record<string, unknown>): void {
    this.log(LogLevel.Info, message, context);
  }

  warn(message: string, context?: Record<string, unknown>): void {
    this.log(LogLevel.Warn, message, context);
  }

  error(message: string, context?: Record<string, unknown>): void {
    this.log(LogLevel.Error, message, context);
  }

  fatal(message: string, context?: Record<string, unknown>): void {
    this.log(LogLevel.Fatal, message, context);
  }

  setLevel(level: LogLevel): void {
    this.level = level;
  }

  withTracing(tracer: TracingClient): FrontendLogger {
    return new FrontendLogger({
      level: this.level,
      context: { ...this.context },
      batchSize: this.batchSize,
      onFlush: this.onFlush,
      console: this.useConsole,
      tracer,
    });
  }

  child(context: Record<string, unknown>): ILogger {
    return new FrontendLogger({
      level: this.level,
      context: { ...this.context, ...context },
      batchSize: this.batchSize,
      onFlush: this.onFlush,
      console: this.useConsole,
      tracer: this.tracer,
    });
  }

  /** Force-flush the current log buffer. */
  async flush(): Promise<void> {
    if (this.buffer.length === 0) return;
    const entries = this.buffer.splice(0, this.buffer.length);
    await this.onFlush?.(entries);
  }

  /** Stop the periodic flush timer. */
  dispose(): void {
    if (this.flushTimer !== null) {
      clearInterval(this.flushTimer);
      this.flushTimer = null;
    }
  }

  private log(level: LogLevel, message: string, context?: Record<string, unknown>): void {
    if (level < this.level) return;

    const mergedContext: Record<string, unknown> = { ...this.context, ...context };

    const entry: LogEntry = {
      level,
      message,
      timestamp: Date.now(),
      context: mergedContext,
    };

    if (this.tracer) {
      const span = this.tracer.activeSpan();
      if (span) {
        entry.traceId = span.traceId;
        entry.spanId = span.spanId;
      }
    }

    this.buffer.push(entry);

    if (this.useConsole) {
      this.printToConsole(entry);
    }

    if (this.buffer.length >= this.batchSize) {
      void this.flush();
    }
  }

  private printToConsole(entry: LogEntry): void {
    const label = LOG_LEVEL_LABELS[entry.level];
    const msg = `[${label}] ${entry.message}`;

    switch (entry.level) {
      case LogLevel.Debug:
        console.debug(msg, entry.context);
        break;
      case LogLevel.Info:
        console.info(msg, entry.context);
        break;
      case LogLevel.Warn:
        console.warn(msg, entry.context);
        break;
      case LogLevel.Error:
      case LogLevel.Fatal:
        console.error(msg, entry.context);
        break;
    }
  }
}
