import type { ExperimentAssignment, ExperimentConfig, ExperimentExposure, ExperimentVariant } from './types';
import type { IExperimentProvider } from './provider';
import { HttpProvider } from './http-provider';

const DEFAULT_CACHE_TIME_MS = 5 * 60 * 1000;
const STORAGE_KEY_PREFIX = 'voxera_exp_';

interface CacheEntry {
  assignment: ExperimentAssignment;
  expiresAt: number;
}

export class ExperimentClient {
  private provider: IExperimentProvider;
  private cache = new Map<string, CacheEntry>();
  private userId?: string;
  private attributes?: Record<string, unknown>;
  private cacheTimeMs: number;

  constructor(config: ExperimentConfig, provider?: IExperimentProvider) {
    this.provider = provider ?? new HttpProvider(config.endpoint, config.apiKey);
    this.userId = config.userId;
    this.attributes = config.attributes;
    this.cacheTimeMs = config.cacheTimeMs ?? DEFAULT_CACHE_TIME_MS;

    this.restoreCache();
  }

  identify(userId: string, attributes?: Record<string, unknown>): void {
    const userChanged = this.userId !== userId;
    this.userId = userId;
    if (attributes) this.attributes = { ...this.attributes, ...attributes };
    if (userChanged) {
      this.cache.clear();
      this.clearStorage();
    }
  }

  async getVariant(experimentKey: string): Promise<ExperimentVariant | null> {
    const cached = this.getCached(experimentKey);
    if (cached) {
      return { key: cached.variantKey, name: cached.variantKey, payload: cached.payload };
    }

    const userId = this.resolveUserId();
    const assignment = await this.provider.fetchAssignment(experimentKey, userId, this.attributes);
    if (!assignment) return null;

    this.setCached(experimentKey, assignment);
    return { key: assignment.variantKey, name: assignment.variantKey, payload: assignment.payload };
  }

  async getAllAssignments(): Promise<ExperimentAssignment[]> {
    const userId = this.resolveUserId();
    const assignments = await this.provider.fetchAllAssignments(userId, this.attributes);
    for (const a of assignments) {
      this.setCached(a.experimentKey, a);
    }
    return assignments;
  }

  exposure(experimentKey: string): void {
    const cached = this.getCached(experimentKey);
    if (!cached) return;

    const exp: ExperimentExposure = {
      experimentKey,
      variantKey: cached.variantKey,
      timestamp: Date.now(),
    };
    void this.provider.reportExposure(exp, this.resolveUserId());
  }

  track(experimentKey: string, metricKey: string, value = 1): void {
    void this.provider.reportMetric(experimentKey, this.resolveUserId(), metricKey, value);
  }

  async prefetch(keys: string[]): Promise<void> {
    const userId = this.resolveUserId();
    await Promise.all(
      keys.map(async (key) => {
        const assignment = await this.provider.fetchAssignment(key, userId, this.attributes);
        if (assignment) this.setCached(key, assignment);
      }),
    );
  }

  private resolveUserId(): string {
    if (this.userId) return this.userId;
    return 'anonymous';
  }

  private getCached(key: string): ExperimentAssignment | null {
    const entry = this.cache.get(key);
    if (!entry) return null;
    if (Date.now() > entry.expiresAt) {
      this.cache.delete(key);
      return null;
    }
    return entry.assignment;
  }

  private setCached(key: string, assignment: ExperimentAssignment): void {
    const entry: CacheEntry = { assignment, expiresAt: Date.now() + this.cacheTimeMs };
    this.cache.set(key, entry);
    this.persistCache();
  }

  private persistCache(): void {
    try {
      const data: Record<string, CacheEntry> = {};
      for (const [k, v] of this.cache) data[k] = v;
      localStorage.setItem(`${STORAGE_KEY_PREFIX}cache`, JSON.stringify(data));
    } catch {
      // localStorage unavailable
    }
  }

  private restoreCache(): void {
    try {
      const raw = localStorage.getItem(`${STORAGE_KEY_PREFIX}cache`);
      if (!raw) return;
      const data = JSON.parse(raw) as Record<string, CacheEntry>;
      for (const [k, v] of Object.entries(data)) {
        if (Date.now() <= v.expiresAt) {
          this.cache.set(k, v);
        }
      }
    } catch {
      // localStorage unavailable or corrupt
    }
  }

  private clearStorage(): void {
    try {
      localStorage.removeItem(`${STORAGE_KEY_PREFIX}cache`);
    } catch {
      // localStorage unavailable
    }
  }
}
