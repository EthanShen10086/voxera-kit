import { ConnectionStatus } from "./types.js";
import type { IWebSocketClient, WebSocketMessage } from "./types.js";

export interface WebSocketClientOptions {
  reconnect?: boolean;
  reconnectDelay?: number;
  maxReconnectAttempts?: number;
  heartbeatInterval?: number;
  heartbeatMessage?: string;
}

const DEFAULT_OPTIONS: Required<WebSocketClientOptions> = {
  reconnect: true,
  reconnectDelay: 1000,
  maxReconnectAttempts: 10,
  heartbeatInterval: 30_000,
  heartbeatMessage: '{"type":"ping"}',
};

export class WebSocketClient implements IWebSocketClient {
  private ws: WebSocket | null = null;
  private url = "";
  private status: ConnectionStatus = ConnectionStatus.Disconnected;
  private reconnectAttempts = 0;
  private reconnectTimer: ReturnType<typeof setTimeout> | null = null;
  private heartbeatTimer: ReturnType<typeof setInterval> | null = null;
  private readonly options: Required<WebSocketClientOptions>;

  private readonly messageHandlers = new Map<string, Set<(msg: WebSocketMessage<any>) => void>>();
  private readonly statusHandlers = new Set<(status: ConnectionStatus) => void>();
  private readonly messageQueue: string[] = [];

  constructor(options: WebSocketClientOptions = {}) {
    this.options = { ...DEFAULT_OPTIONS, ...options };
  }

  connect(url: string): void {
    if (this.ws && (this.status === ConnectionStatus.Connected || this.status === ConnectionStatus.Connecting)) {
      return;
    }

    this.url = url;
    this.setStatus(ConnectionStatus.Connecting);

    this.ws = new WebSocket(url);

    this.ws.onopen = () => {
      this.reconnectAttempts = 0;
      this.setStatus(ConnectionStatus.Connected);
      this.flushQueue();
      this.startHeartbeat();
    };

    this.ws.onclose = () => {
      this.stopHeartbeat();

      if (this.status === ConnectionStatus.Disconnecting) {
        this.setStatus(ConnectionStatus.Disconnected);
        return;
      }

      this.setStatus(ConnectionStatus.Disconnected);
      this.attemptReconnect();
    };

    this.ws.onerror = () => {
      /* errors surface through onclose */
    };

    this.ws.onmessage = (event: MessageEvent) => {
      this.handleMessage(event);
    };
  }

  disconnect(): void {
    this.setStatus(ConnectionStatus.Disconnecting);
    this.clearReconnect();
    this.stopHeartbeat();

    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }
  }

  send<T>(type: string, payload: T): void {
    const message: WebSocketMessage<T> = {
      type,
      payload,
      timestamp: Date.now(),
    };

    const serialized = JSON.stringify(message);

    if (this.status === ConnectionStatus.Connected && this.ws?.readyState === WebSocket.OPEN) {
      this.ws.send(serialized);
    } else {
      this.messageQueue.push(serialized);
    }
  }

  on<T>(type: string, handler: (msg: WebSocketMessage<T>) => void): () => void {
    let handlers = this.messageHandlers.get(type);
    if (!handlers) {
      handlers = new Set();
      this.messageHandlers.set(type, handlers);
    }
    handlers.add(handler as (msg: WebSocketMessage<any>) => void);

    return () => {
      handlers!.delete(handler as (msg: WebSocketMessage<any>) => void);
      if (handlers!.size === 0) {
        this.messageHandlers.delete(type);
      }
    };
  }

  onStatusChange(handler: (status: ConnectionStatus) => void): () => void {
    this.statusHandlers.add(handler);
    return () => {
      this.statusHandlers.delete(handler);
    };
  }

  private handleMessage(event: MessageEvent): void {
    let parsed: WebSocketMessage<unknown>;
    try {
      parsed = JSON.parse(String(event.data)) as WebSocketMessage<unknown>;
    } catch {
      return;
    }

    if (parsed.type === "pong") return;

    const handlers = this.messageHandlers.get(parsed.type);
    if (handlers) {
      for (const handler of handlers) {
        handler(parsed);
      }
    }
  }

  private setStatus(status: ConnectionStatus): void {
    if (this.status === status) return;
    this.status = status;
    for (const handler of this.statusHandlers) {
      handler(status);
    }
  }

  private flushQueue(): void {
    while (this.messageQueue.length > 0 && this.ws?.readyState === WebSocket.OPEN) {
      const msg = this.messageQueue.shift()!;
      this.ws.send(msg);
    }
  }

  private attemptReconnect(): void {
    if (
      !this.options.reconnect ||
      this.reconnectAttempts >= this.options.maxReconnectAttempts
    ) {
      return;
    }

    this.setStatus(ConnectionStatus.Reconnecting);
    this.reconnectAttempts++;

    const delay = this.options.reconnectDelay * Math.pow(2, this.reconnectAttempts - 1);

    this.reconnectTimer = setTimeout(() => {
      this.connect(this.url);
    }, delay);
  }

  private clearReconnect(): void {
    if (this.reconnectTimer !== null) {
      clearTimeout(this.reconnectTimer);
      this.reconnectTimer = null;
    }
    this.reconnectAttempts = 0;
  }

  private startHeartbeat(): void {
    this.stopHeartbeat();
    this.heartbeatTimer = setInterval(() => {
      if (this.ws?.readyState === WebSocket.OPEN) {
        this.ws.send(this.options.heartbeatMessage);
      }
    }, this.options.heartbeatInterval);
  }

  private stopHeartbeat(): void {
    if (this.heartbeatTimer !== null) {
      clearInterval(this.heartbeatTimer);
      this.heartbeatTimer = null;
    }
  }
}
