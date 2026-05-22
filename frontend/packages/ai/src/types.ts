export type Role = 'system' | 'user' | 'assistant';

export interface Message {
  role: Role;
  content: string;
  timestamp?: number;
}

export interface ChatRequest {
  model?: string;
  messages: Message[];
  maxTokens?: number;
  temperature?: number;
  stream?: boolean;
  metadata?: Record<string, string>;
}

export interface ChatResponse {
  id: string;
  model: string;
  content: string;
  finishReason: string;
  usage: TokenUsage;
}

export interface TokenUsage {
  inputTokens: number;
  outputTokens: number;
  totalTokens: number;
}

export interface StreamChunk {
  content: string;
  finishReason?: string;
  done: boolean;
}

export interface AIConfig {
  endpoint: string;
  apiKey?: string;
  defaultModel?: string;
  maxRetries?: number;
  timeout?: number;
  onError?: (error: Error) => void;
}

export interface UsageInfo {
  dailyTokens: number;
  monthlyTokens: number;
  dailyRequests: number;
  quotaLimit: QuotaLimit;
}

export interface QuotaLimit {
  dailyTokens: number;
  monthlyTokens: number;
  dailyRequests: number;
  allowedModels: string[];
}
