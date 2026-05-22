import type { IAnalyticsProvider } from './provider';
import type { AnalyticsEvent, UserProperties } from './types';

const MAX_RETRIES = 3;
const BASE_DELAY_MS = 500;

export class HttpProvider implements IAnalyticsProvider {
  name = 'http';

  private endpoint: string;
  private apiKey?: string;
  private pendingEvents: AnalyticsEvent[] = [];

  constructor(endpoint: string, apiKey?: string) {
    this.endpoint = endpoint;
    this.apiKey = apiKey;
  }

  async track(event: AnalyticsEvent): Promise<void> {
    this.pendingEvents.push(event);
  }

  async trackBatch(events: AnalyticsEvent[]): Promise<void> {
    await this.send('/batch', { events });
  }

  async identify(userId: string, properties?: UserProperties): Promise<void> {
    await this.send('/identify', { userId, properties });
  }

  async alias(previousId: string, newId: string): Promise<void> {
    await this.send('/alias', { previousId, newId });
  }

  async group(groupId: string, properties?: Record<string, unknown>): Promise<void> {
    await this.send('/group', { groupId, properties });
  }

  async flush(): Promise<void> {
    if (this.pendingEvents.length === 0) return;
    const events = this.pendingEvents.splice(0);
    await this.trackBatch(events);
  }

  reset(): void {
    this.pendingEvents = [];
  }

  private async send(path: string, body: unknown, useKeepalive = false): Promise<void> {
    const url = `${this.endpoint}${path}`;
    const headers: Record<string, string> = { 'Content-Type': 'application/json' };
    if (this.apiKey) headers['Authorization'] = `Bearer ${this.apiKey}`;

    for (let attempt = 0; attempt <= MAX_RETRIES; attempt++) {
      try {
        const res = await fetch(url, {
          method: 'POST',
          headers,
          body: JSON.stringify(body),
          keepalive: useKeepalive,
        });
        if (res.ok) return;
        if (res.status >= 400 && res.status < 500) return; // don't retry client errors
      } catch {
        // network error, will retry
      }

      if (attempt < MAX_RETRIES) {
        await new Promise((resolve) => setTimeout(resolve, BASE_DELAY_MS * 2 ** attempt));
      }
    }
  }
}
