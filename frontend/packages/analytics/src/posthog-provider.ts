import type { IAnalyticsProvider } from './provider';
import type { AnalyticsEvent, UserProperties } from './types';

export class PostHogProvider implements IAnalyticsProvider {
  name = 'posthog';

  private endpoint: string;
  private apiKey: string;

  constructor(endpoint: string, apiKey: string) {
    this.endpoint = endpoint.replace(/\/$/, '');
    this.apiKey = apiKey;
  }

  async track(event: AnalyticsEvent): Promise<void> {
    await this.capture({
      event: event.name,
      distinct_id: event.userId ?? 'anonymous',
      properties: {
        ...event.properties,
        ...event.context,
        $session_id: event.sessionId,
        timestamp: new Date(event.timestamp).toISOString(),
      },
    });
  }

  async trackBatch(events: AnalyticsEvent[]): Promise<void> {
    const batch = events.map((event) => ({
      event: event.name,
      distinct_id: event.userId ?? 'anonymous',
      properties: {
        ...event.properties,
        ...event.context,
        $session_id: event.sessionId,
        timestamp: new Date(event.timestamp).toISOString(),
      },
    }));
    await this.post('/capture', { api_key: this.apiKey, batch });
  }

  async identify(userId: string, properties?: UserProperties): Promise<void> {
    await this.capture({
      event: '$identify',
      distinct_id: userId,
      $set: properties,
    });
  }

  async alias(previousId: string, newId: string): Promise<void> {
    await this.capture({
      event: '$create_alias',
      distinct_id: newId,
      properties: { alias: previousId },
    });
  }

  async group(groupId: string, properties?: Record<string, unknown>): Promise<void> {
    await this.capture({
      event: '$groupidentify',
      distinct_id: groupId,
      properties: { $group_type: 'company', $group_key: groupId, $group_set: properties },
    });
  }

  async flush(): Promise<void> {
    // PostHog provider sends events immediately
  }

  reset(): void {
    // no local state to clear
  }

  private async capture(payload: Record<string, unknown>): Promise<void> {
    await this.post('/capture', { api_key: this.apiKey, ...payload });
  }

  private async post(path: string, body: unknown): Promise<void> {
    await fetch(`${this.endpoint}${path}`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(body),
    });
  }
}
