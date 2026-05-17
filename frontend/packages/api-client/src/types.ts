export interface RequestConfig {
  url: string;
  method: "GET" | "POST" | "PUT" | "DELETE" | "PATCH";
  headers?: Record<string, string>;
  params?: Record<string, string>;
  data?: unknown;
  timeout?: number;
  signal?: AbortSignal;
  responseType?: "json" | "text" | "blob" | "arrayBuffer";
}

export interface Response<T> {
  data: T;
  status: number;
  headers: Headers;
  config: RequestConfig;
}

export type RequestInterceptor = (
  config: RequestConfig,
) => RequestConfig | Promise<RequestConfig>;

export type ResponseInterceptor<T = unknown> = (
  response: Response<T>,
) => Response<T> | Promise<Response<T>>;

export interface RetryConfig {
  maxRetries: number;
  retryDelay: number;
  retryOn?: number[];
}

export interface IHttpClient {
  get<T>(url: string, config?: Partial<RequestConfig>): Promise<Response<T>>;
  post<T>(url: string, data?: unknown, config?: Partial<RequestConfig>): Promise<Response<T>>;
  put<T>(url: string, data?: unknown, config?: Partial<RequestConfig>): Promise<Response<T>>;
  delete<T>(url: string, config?: Partial<RequestConfig>): Promise<Response<T>>;
  patch<T>(url: string, data?: unknown, config?: Partial<RequestConfig>): Promise<Response<T>>;
  request<T>(config: RequestConfig): Promise<Response<T>>;
  addRequestInterceptor(interceptor: RequestInterceptor): void;
  addResponseInterceptor<T = unknown>(interceptor: ResponseInterceptor<T>): void;
}

export interface WebSocketMessage<T> {
  type: string;
  payload: T;
  timestamp: number;
}

export enum ConnectionStatus {
  Connecting = "connecting",
  Connected = "connected",
  Disconnecting = "disconnecting",
  Disconnected = "disconnected",
  Reconnecting = "reconnecting",
}

export interface IWebSocketClient {
  connect(url: string): void;
  disconnect(): void;
  send<T>(type: string, payload: T): void;
  on<T>(type: string, handler: (msg: WebSocketMessage<T>) => void): () => void;
  onStatusChange(handler: (status: ConnectionStatus) => void): () => void;
}
