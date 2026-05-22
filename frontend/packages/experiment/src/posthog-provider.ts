import type { IExperimentProvider } from './provider';
import type { ExperimentAssignment, ExperimentExposure } from './types';

export class PostHogProvider implements IExperimentProvider {
  name = 'posthog';

  private endpoint: string;
  private apiKey: string;

  constructor(endpoint: string, apiKey: string) {
    this.endpoint = endpoint.replace(/\/$/, '');
    this.apiKey = apiKey;
  }

  async fetchAssignment(
    experimentKey: string,
    userId: string,
    attributes?: Record<string, unknown>,
  ): Promise<ExperimentAssignment | null> {
    const res = await fetch(`${this.endpoint}/decide/?v=3`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        api_key: this.apiKey,
        distinct_id: userId,
        person_properties: attributes,
      }),
    });
    if (!res.ok) return null;

    const data = (await res.json()) as Record<string, unknown>;
    const featureFlags = data.featureFlags as Record<string, string | boolean> | undefined;
    const flagPayloads = data.featureFlagPayloads as Record<string, string> | undefined;

    if (!featureFlags || !(experimentKey in featureFlags)) return null;

    const variantKey = String(featureFlags[experimentKey]);
    let payload: Record<string, unknown> | undefined;
    if (flagPayloads?.[experimentKey]) {
      try {
        payload = JSON.parse(flagPayloads[experimentKey]) as Record<string, unknown>;
      } catch {
        // non-JSON payload
      }
    }

    return { experimentKey, variantKey, payload };
  }

  async fetchAllAssignments(
    userId: string,
    attributes?: Record<string, unknown>,
  ): Promise<ExperimentAssignment[]> {
    const res = await fetch(`${this.endpoint}/decide/?v=3`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        api_key: this.apiKey,
        distinct_id: userId,
        person_properties: attributes,
      }),
    });
    if (!res.ok) return [];

    const data = (await res.json()) as Record<string, unknown>;
    const featureFlags = data.featureFlags as Record<string, string | boolean> | undefined;
    const flagPayloads = data.featureFlagPayloads as Record<string, string> | undefined;

    if (!featureFlags) return [];

    return Object.entries(featureFlags).map(([key, value]) => {
      let payload: Record<string, unknown> | undefined;
      if (flagPayloads?.[key]) {
        try {
          payload = JSON.parse(flagPayloads[key]) as Record<string, unknown>;
        } catch {
          // non-JSON payload
        }
      }
      return { experimentKey: key, variantKey: String(value), payload };
    });
  }

  async reportExposure(exposure: ExperimentExposure, userId: string): Promise<void> {
    await fetch(`${this.endpoint}/capture`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        api_key: this.apiKey,
        distinct_id: userId,
        event: '$feature_flag_called',
        properties: {
          $feature_flag: exposure.experimentKey,
          $feature_flag_response: exposure.variantKey,
        },
        timestamp: new Date(exposure.timestamp).toISOString(),
      }),
    });
  }

  async reportMetric(
    experimentKey: string,
    userId: string,
    metricKey: string,
    value: number,
  ): Promise<void> {
    await fetch(`${this.endpoint}/capture`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        api_key: this.apiKey,
        distinct_id: userId,
        event: metricKey,
        properties: {
          $experiment_key: experimentKey,
          value,
        },
      }),
    });
  }
}
