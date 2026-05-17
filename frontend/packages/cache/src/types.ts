export interface CacheEntry<T> {
  value: T;
  expiresAt: number | null;
  createdAt: number;
}

export interface ICache {
  get<T>(key: string): T | undefined;
  set<T>(key: string, value: T, ttl?: number): void;
  has(key: string): boolean;
  delete(key: string): boolean;
  clear(): void;
  size(): number;
  keys(): string[];
}

export interface ICacheWithEvents extends ICache {
  onEvict(callback: (key: string, value: unknown) => void): () => void;
}

export interface CacheConfig {
  maxSize?: number;
  defaultTTL?: number;
  evictionPolicy?: "lru" | "lfu" | "fifo";
}
