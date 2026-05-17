import type {
  IHttpClient,
  RequestConfig,
  RequestInterceptor,
  Response,
  ResponseInterceptor,
  RetryConfig,
} from "./types.js";

export class HttpClientError extends Error {
  constructor(
    message: string,
    public readonly status: number | undefined,
    public readonly config: RequestConfig,
    public readonly response?: Response<unknown>,
  ) {
    super(message);
    this.name = "HttpClientError";
  }
}

export interface HttpClientOptions {
  baseURL?: string;
  defaultHeaders?: Record<string, string>;
  timeout?: number;
  retry?: RetryConfig;
}

export class HttpClient implements IHttpClient {
  private readonly baseURL: string;
  private readonly defaultHeaders: Record<string, string>;
  private readonly defaultTimeout: number;
  private readonly retryConfig: RetryConfig;
  private readonly requestInterceptors: RequestInterceptor[] = [];
  private readonly responseInterceptors: ResponseInterceptor<any>[] = [];

  constructor(options: HttpClientOptions = {}) {
    this.baseURL = options.baseURL?.replace(/\/+$/, "") ?? "";
    this.defaultHeaders = options.defaultHeaders ?? {};
    this.defaultTimeout = options.timeout ?? 30_000;
    this.retryConfig = options.retry ?? { maxRetries: 0, retryDelay: 1000 };
  }

  addRequestInterceptor(interceptor: RequestInterceptor): void {
    this.requestInterceptors.push(interceptor);
  }

  addResponseInterceptor<T = unknown>(interceptor: ResponseInterceptor<T>): void {
    this.responseInterceptors.push(interceptor as ResponseInterceptor<any>);
  }

  async get<T>(url: string, config?: Partial<RequestConfig>): Promise<Response<T>> {
    return this.request<T>({ ...config, url, method: "GET" } as RequestConfig);
  }

  async post<T>(url: string, data?: unknown, config?: Partial<RequestConfig>): Promise<Response<T>> {
    return this.request<T>({ ...config, url, method: "POST", data } as RequestConfig);
  }

  async put<T>(url: string, data?: unknown, config?: Partial<RequestConfig>): Promise<Response<T>> {
    return this.request<T>({ ...config, url, method: "PUT", data } as RequestConfig);
  }

  async delete<T>(url: string, config?: Partial<RequestConfig>): Promise<Response<T>> {
    return this.request<T>({ ...config, url, method: "DELETE" } as RequestConfig);
  }

  async patch<T>(url: string, data?: unknown, config?: Partial<RequestConfig>): Promise<Response<T>> {
    return this.request<T>({ ...config, url, method: "PATCH", data } as RequestConfig);
  }

  async request<T>(config: RequestConfig): Promise<Response<T>> {
    let resolvedConfig = { ...config };

    for (const interceptor of this.requestInterceptors) {
      resolvedConfig = await interceptor(resolvedConfig);
    }

    const fullURL = this.buildURL(resolvedConfig);
    const timeout = resolvedConfig.timeout ?? this.defaultTimeout;

    let lastError: unknown;

    const maxAttempts = this.retryConfig.maxRetries + 1;
    for (let attempt = 0; attempt < maxAttempts; attempt++) {
      try {
        const response = await this.executeRequest<T>(fullURL, resolvedConfig, timeout);

        let processed = response as Response<any>;
        for (const interceptor of this.responseInterceptors) {
          processed = await interceptor(processed);
        }

        return processed as Response<T>;
      } catch (err) {
        lastError = err;

        const isRetryable =
          err instanceof HttpClientError &&
          err.status !== undefined &&
          (this.retryConfig.retryOn?.includes(err.status) ?? err.status >= 500);

        if (!isRetryable || attempt >= this.retryConfig.maxRetries) {
          throw err;
        }

        const delay = this.retryConfig.retryDelay * Math.pow(2, attempt);
        await this.sleep(delay);
      }
    }

    throw lastError;
  }

  private async executeRequest<T>(
    url: string,
    config: RequestConfig,
    timeout: number,
  ): Promise<Response<T>> {
    const controller = new AbortController();
    const externalSignal = config.signal;

    if (externalSignal?.aborted) {
      controller.abort(externalSignal.reason);
    } else {
      externalSignal?.addEventListener("abort", () => controller.abort(externalSignal.reason), {
        once: true,
      });
    }

    const timer = setTimeout(() => controller.abort("Request timeout"), timeout);

    const headers: Record<string, string> = {
      ...this.defaultHeaders,
      ...config.headers,
    };

    const hasBody = config.data !== undefined && config.method !== "GET";
    if (hasBody && !headers["Content-Type"]) {
      headers["Content-Type"] = "application/json";
    }

    try {
      const fetchResponse = await fetch(url, {
        method: config.method,
        headers,
        body: hasBody ? JSON.stringify(config.data) : undefined,
        signal: controller.signal,
      });

      if (!fetchResponse.ok) {
        throw new HttpClientError(
          `Request failed with status ${fetchResponse.status}`,
          fetchResponse.status,
          config,
        );
      }

      const data = await this.parseResponse<T>(fetchResponse, config.responseType);

      return {
        data,
        status: fetchResponse.status,
        headers: fetchResponse.headers,
        config,
      };
    } catch (err) {
      if (err instanceof HttpClientError) throw err;

      const message =
        err instanceof DOMException && err.name === "AbortError"
          ? `Request to ${config.url} timed out after ${timeout}ms`
          : `Network error requesting ${config.url}: ${err instanceof Error ? err.message : String(err)}`;

      throw new HttpClientError(message, undefined, config);
    } finally {
      clearTimeout(timer);
    }
  }

  private async parseResponse<T>(
    response: globalThis.Response,
    responseType?: RequestConfig["responseType"],
  ): Promise<T> {
    switch (responseType) {
      case "text":
        return (await response.text()) as T;
      case "blob":
        return (await response.blob()) as T;
      case "arrayBuffer":
        return (await response.arrayBuffer()) as T;
      case "json":
      default:
        return (await response.json()) as T;
    }
  }

  private buildURL(config: RequestConfig): string {
    const base = config.url.startsWith("http") ? config.url : `${this.baseURL}/${config.url.replace(/^\/+/, "")}`;

    if (!config.params || Object.keys(config.params).length === 0) {
      return base;
    }

    const searchParams = new URLSearchParams(config.params);
    const separator = base.includes("?") ? "&" : "?";
    return `${base}${separator}${searchParams.toString()}`;
  }

  private sleep(ms: number): Promise<void> {
    return new Promise((resolve) => setTimeout(resolve, ms));
  }
}
