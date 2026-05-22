import type { AnalyticsEvent, UserProperties } from './types';

export interface IAnalyticsProvider {
  name: string;
  track(event: AnalyticsEvent): Promise<void>;
  trackBatch(events: AnalyticsEvent[]): Promise<void>;
  identify(userId: string, properties?: UserProperties): Promise<void>;
  alias(previousId: string, newId: string): Promise<void>;
  group(groupId: string, properties?: Record<string, unknown>): Promise<void>;
  flush(): Promise<void>;
  reset(): void;
}
