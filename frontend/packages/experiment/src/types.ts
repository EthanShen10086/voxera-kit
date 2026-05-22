export interface ExperimentConfig {
  endpoint: string;
  apiKey?: string;
  userId?: string;
  attributes?: Record<string, unknown>;
  /** How long to cache assignments in ms (default 5 min) */
  cacheTimeMs?: number;
}

export interface ExperimentVariant {
  key: string;
  name: string;
  payload?: Record<string, unknown>;
}

export interface ExperimentAssignment {
  experimentKey: string;
  variantKey: string;
  payload?: Record<string, unknown>;
}

export interface ExperimentExposure {
  experimentKey: string;
  variantKey: string;
  timestamp: number;
}
