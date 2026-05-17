export interface SSEClientConfig {
  url: string;
  withCredentials?: boolean;
  headers?: Record<string, string>;
  reconnectDelay?: number;
  maxRetries?: number;
}

export interface SSEEvent {
  id?: string;
  event?: string;
  data: string;
  retry?: number;
}

export type SSEClientStatus = "connecting" | "connected" | "disconnected" | "error";

export interface ISSEClient {
  connect(): void;
  close(): void;
  on(event: string, handler: (data: SSEEvent) => void): () => void;
  onStatus(handler: (status: SSEClientStatus) => void): () => void;
  status(): SSEClientStatus;
}

export class SSEClient implements ISSEClient {
  private readonly config: SSEClientConfig;
  private eventSource: EventSource | null = null;
  private currentStatus: SSEClientStatus = "disconnected";
  private retryCount = 0;
  private retryTimer: ReturnType<typeof setTimeout> | null = null;

  private readonly eventHandlers = new Map<string, Set<(data: SSEEvent) => void>>();
  private readonly statusHandlers = new Set<(status: SSEClientStatus) => void>();

  constructor(config: SSEClientConfig) {
    this.config = {
      reconnectDelay: 1000,
      maxRetries: 5,
      ...config,
    };
  }

  connect(): void {
    if (this.eventSource) {
      this.eventSource.close();
    }

    this.setStatus("connecting");

    this.eventSource = new EventSource(this.config.url, {
      withCredentials: this.config.withCredentials,
    });

    this.eventSource.onopen = () => {
      this.retryCount = 0;
      this.setStatus("connected");
    };

    this.eventSource.onerror = () => {
      this.eventSource?.close();
      this.eventSource = null;
      this.setStatus("error");
      this.attemptReconnect();
    };

    this.eventSource.onmessage = (event: MessageEvent) => {
      this.dispatchEvent("message", event);
    };

    for (const eventName of this.eventHandlers.keys()) {
      if (eventName !== "message") {
        this.addNativeListener(eventName);
      }
    }
  }

  close(): void {
    this.clearRetryTimer();
    this.retryCount = 0;

    if (this.eventSource) {
      this.eventSource.close();
      this.eventSource = null;
    }

    this.setStatus("disconnected");
  }

  on(event: string, handler: (data: SSEEvent) => void): () => void {
    let handlers = this.eventHandlers.get(event);
    if (!handlers) {
      handlers = new Set();
      this.eventHandlers.set(event, handlers);

      if (this.eventSource && event !== "message") {
        this.addNativeListener(event);
      }
    }

    handlers.add(handler);

    return () => {
      handlers!.delete(handler);
      if (handlers!.size === 0) {
        this.eventHandlers.delete(event);
      }
    };
  }

  onStatus(handler: (status: SSEClientStatus) => void): () => void {
    this.statusHandlers.add(handler);
    return () => {
      this.statusHandlers.delete(handler);
    };
  }

  status(): SSEClientStatus {
    return this.currentStatus;
  }

  private addNativeListener(eventName: string): void {
    this.eventSource?.addEventListener(eventName, ((event: MessageEvent) => {
      this.dispatchEvent(eventName, event);
    }) as EventListener);
  }

  private dispatchEvent(eventName: string, event: MessageEvent): void {
    const sseEvent: SSEEvent = {
      id: event.lastEventId || undefined,
      event: eventName,
      data: String(event.data),
    };

    const handlers = this.eventHandlers.get(eventName);
    if (handlers) {
      for (const handler of handlers) {
        handler(sseEvent);
      }
    }
  }

  private setStatus(status: SSEClientStatus): void {
    if (this.currentStatus === status) return;
    this.currentStatus = status;
    for (const handler of this.statusHandlers) {
      handler(status);
    }
  }

  private attemptReconnect(): void {
    const maxRetries = this.config.maxRetries ?? 5;
    if (this.retryCount >= maxRetries) {
      this.setStatus("disconnected");
      return;
    }

    this.retryCount++;
    const delay = (this.config.reconnectDelay ?? 1000) * Math.pow(2, this.retryCount - 1);

    this.retryTimer = setTimeout(() => {
      this.connect();
    }, delay);
  }

  private clearRetryTimer(): void {
    if (this.retryTimer !== null) {
      clearTimeout(this.retryTimer);
      this.retryTimer = null;
    }
  }
}
