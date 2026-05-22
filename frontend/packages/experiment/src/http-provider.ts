import type { IExperimentProvider } from './provider';
import type { ExperimentAssignment, ExperimentExposure } from './types';

export class HttpProvider implements IExperimentProvider {
  name = 'http';

  private endpoint: string;
  private apiKey?: string;

  constructor(endpoint: string, apiKey?: string) {
    this.endpoint = endpoint.replace(/\/$/, '');
    this.apiKey = apiKey;
  }

  async fetchAssignment(
    experimentKey: string,
    userId: string,
    attributes?: Record<string, unknown>,
  ): Promise<ExperimentAssignment | null> {
    const params = new URLSearchParams({ key: experimentKey, user_id: userId });
    if (attributes) params.set('attributes', JSON.stringify(attributes));

    const res = await fetch(`${this.endpoint}/api/v1/experiments/assign?${params}`, {
      headers: this.headers(),
    });
    if (!res.ok) return null;
    return (await res.json()) as ExperimentAssignment;
  }

  async fetchAllAssignments(
    userId: string,
    attributes?: Record<string, unknown>,
  ): Promise<ExperimentAssignment[]> {
    const params = new URLSearchParams({ user_id: userId });
    if (attributes) params.set('attributes', JSON.stringify(attributes));

    const res = await fetch(`${this.endpoint}/api/v1/experiments/assign?${params}`, {
      headers: this.headers(),
    });
    if (!res.ok) return [];
    return (await res.json()) as ExperimentAssignment[];
  }

  async reportExposure(exposure: ExperimentExposure, userId: string): Promise<void> {
    await fetch(`${this.endpoint}/api/v1/experiments/exposure`, {
      method: 'POST',
      headers: this.headers(),
      body: JSON.stringify({ ...exposure, userId }),
    });
  }

  async reportMetric(
    experimentKey: string,
    userId: string,
    metricKey: string,
    value: number,
  ): Promise<void> {
    await fetch(`${this.endpoint}/api/v1/experiments/metric`, {
      method: 'POST',
      headers: this.headers(),
      body: JSON.stringify({ experimentKey, userId, metricKey, value }),
    });
  }

  private headers(): Record<string, string> {
    const h: Record<string, string> = { 'Content-Type': 'application/json' };
    if (this.apiKey) h['Authorization'] = `Bearer ${this.apiKey}`;
    return h;
  }
}
