import type {
  AIConfig,
  ChatRequest,
  ChatResponse,
  StreamChunk,
  UsageInfo,
} from './types';

const DEFAULT_MAX_RETRIES = 3;
const DEFAULT_TIMEOUT = 30_000;
const RETRY_BASE_DELAY = 1_000;

export class AIClient {
  private config: Required<Pick<AIConfig, 'endpoint' | 'maxRetries' | 'timeout'>>;
  private apiKey?: string;
  private onError?: (error: Error) => void;
  private defaultModel?: string;
  private abortController: AbortController | null = null;

  constructor(config: AIConfig) {
    this.config = {
      endpoint: config.endpoint.replace(/\/+$/, ''),
      maxRetries: config.maxRetries ?? DEFAULT_MAX_RETRIES,
      timeout: config.timeout ?? DEFAULT_TIMEOUT,
    };
    this.apiKey = config.apiKey;
    this.defaultModel = config.defaultModel;
    this.onError = config.onError;
  }

  async chat(request: ChatRequest): Promise<ChatResponse> {
    const body = {
      ...request,
      model: request.model ?? this.defaultModel,
      stream: false,
    };
    return this.fetchWithRetry<ChatResponse>(
      `${this.config.endpoint}/api/v1/ai/chat`,
      { method: 'POST', body: JSON.stringify(body) },
    );
  }

  async chatStream(
    request: ChatRequest,
    onChunk: (chunk: StreamChunk) => void,
  ): Promise<void> {
    this.abortController = new AbortController();
    const body = {
      ...request,
      model: request.model ?? this.defaultModel,
      stream: true,
    };

    const response = await this.fetch(
      `${this.config.endpoint}/api/v1/ai/chat`,
      {
        method: 'POST',
        body: JSON.stringify(body),
        signal: this.abortController.signal,
      },
    );

    if (!response.body) {
      throw new Error('Response body is null — streaming not supported');
    }

    const reader = response.body.getReader();
    const decoder = new TextDecoder();
    let buffer = '';

    try {
      while (true) {
        const { done, value } = await reader.read();
        if (done) break;

        buffer += decoder.decode(value, { stream: true });
        const lines = buffer.split('\n');
        buffer = lines.pop() ?? '';

        for (const line of lines) {
          const trimmed = line.trim();
          if (!trimmed || !trimmed.startsWith('data:')) continue;

          const data = trimmed.slice(5).trim();
          if (data === '[DONE]') {
            onChunk({ content: '', done: true });
            return;
          }

          try {
            const chunk = JSON.parse(data) as StreamChunk;
            onChunk(chunk);
          } catch {
            // skip malformed SSE lines
          }
        }
      }

      onChunk({ content: '', done: true });
    } finally {
      reader.releaseLock();
      this.abortController = null;
    }
  }

  async summarize(
    text: string,
    options?: { language?: string; maxLength?: number },
  ): Promise<string> {
    const body = { text, ...options };
    const res = await this.fetchWithRetry<{ summary: string }>(
      `${this.config.endpoint}/api/v1/ai/summarize`,
      { method: 'POST', body: JSON.stringify(body) },
    );
    return res.summary;
  }

  async translate(
    text: string,
    targetLang: string,
    sourceLang?: string,
  ): Promise<string> {
    const body = { text, targetLang, sourceLang };
    const res = await this.fetchWithRetry<{ translation: string }>(
      `${this.config.endpoint}/api/v1/ai/translate`,
      { method: 'POST', body: JSON.stringify(body) },
    );
    return res.translation;
  }

  async analyze(text: string, analysisType: string): Promise<string> {
    const body = { text, analysisType };
    const res = await this.fetchWithRetry<{ result: string }>(
      `${this.config.endpoint}/api/v1/ai/analyze`,
      { method: 'POST', body: JSON.stringify(body) },
    );
    return res.result;
  }

  async getUsage(): Promise<UsageInfo> {
    return this.fetchWithRetry<UsageInfo>(
      `${this.config.endpoint}/api/v1/ai/usage`,
      { method: 'GET' },
    );
  }

  async getModels(): Promise<Array<{ id: string; provider: string }>> {
    return this.fetchWithRetry<Array<{ id: string; provider: string }>>(
      `${this.config.endpoint}/api/v1/ai/models`,
      { method: 'GET' },
    );
  }

  abort(): void {
    this.abortController?.abort();
    this.abortController = null;
  }

  dispose(): void {
    this.abort();
  }

  // ---------------------------------------------------------------------------

  private async fetch(url: string, init: RequestInit): Promise<Response> {
    const headers: Record<string, string> = {
      'Content-Type': 'application/json',
      ...(this.apiKey ? { Authorization: `Bearer ${this.apiKey}` } : {}),
    };

    const controller = new AbortController();
    const timeoutId = setTimeout(() => controller.abort(), this.config.timeout);

    const mergedSignal = init.signal
      ? anySignal([init.signal, controller.signal])
      : controller.signal;

    try {
      const response = await fetch(url, {
        ...init,
        headers: { ...headers, ...(init.headers as Record<string, string>) },
        signal: mergedSignal,
      });

      if (!response.ok) {
        const text = await response.text().catch(() => '');
        throw new Error(`AI request failed (${response.status}): ${text}`);
      }

      return response;
    } finally {
      clearTimeout(timeoutId);
    }
  }

  private async fetchWithRetry<T>(url: string, init: RequestInit): Promise<T> {
    let lastError: Error | undefined;

    for (let attempt = 0; attempt <= this.config.maxRetries; attempt++) {
      try {
        const response = await this.fetch(url, init);
        return (await response.json()) as T;
      } catch (err) {
        lastError = err instanceof Error ? err : new Error(String(err));
        if (attempt < this.config.maxRetries) {
          const delay = RETRY_BASE_DELAY * 2 ** attempt;
          await sleep(delay);
        }
      }
    }

    this.onError?.(lastError!);
    throw lastError;
  }
}

function sleep(ms: number): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

function anySignal(signals: AbortSignal[]): AbortSignal {
  const controller = new AbortController();
  for (const signal of signals) {
    if (signal.aborted) {
      controller.abort(signal.reason);
      return controller.signal;
    }
    signal.addEventListener('abort', () => controller.abort(signal.reason), {
      once: true,
    });
  }
  return controller.signal;
}
