import type { CacheConfig, CacheEntry, ICache } from "./types.js";

export class MemoryCache implements ICache {
  private readonly store = new Map<string, CacheEntry<unknown>>();
  private readonly maxSize: number;
  private readonly defaultTTL: number | undefined;

  constructor(config: CacheConfig = {}) {
    this.maxSize = config.maxSize ?? Infinity;
    this.defaultTTL = config.defaultTTL;
  }

  get<T>(key: string): T | undefined {
    const entry = this.store.get(key);
    if (!entry) return undefined;

    if (entry.expiresAt !== null && Date.now() > entry.expiresAt) {
      this.store.delete(key);
      return undefined;
    }

    this.store.delete(key);
    this.store.set(key, entry);

    return entry.value as T;
  }

  set<T>(key: string, value: T, ttl?: number): void {
    if (this.store.has(key)) {
      this.store.delete(key);
    }

    while (this.store.size >= this.maxSize) {
      const oldestKey = this.store.keys().next().value;
      if (oldestKey !== undefined) {
        this.store.delete(oldestKey);
      }
    }

    const effectiveTTL = ttl ?? this.defaultTTL;

    const entry: CacheEntry<T> = {
      value,
      expiresAt: effectiveTTL != null ? Date.now() + effectiveTTL : null,
      createdAt: Date.now(),
    };

    this.store.set(key, entry);
  }

  has(key: string): boolean {
    const entry = this.store.get(key);
    if (!entry) return false;

    if (entry.expiresAt !== null && Date.now() > entry.expiresAt) {
      this.store.delete(key);
      return false;
    }

    return true;
  }

  delete(key: string): boolean {
    return this.store.delete(key);
  }

  clear(): void {
    this.store.clear();
  }

  size(): number {
    this.evictExpired();
    return this.store.size;
  }

  keys(): string[] {
    this.evictExpired();
    return [...this.store.keys()];
  }

  private evictExpired(): void {
    const now = Date.now();
    for (const [key, entry] of this.store) {
      if (entry.expiresAt !== null && now > entry.expiresAt) {
        this.store.delete(key);
      }
    }
  }
}
