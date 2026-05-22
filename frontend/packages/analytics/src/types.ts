export interface EventContext {
  platform?: string;
  appVersion?: string;
  locale?: string;
  userAgent?: string;
  pageUrl?: string;
  referrer?: string;
  utmSource?: string;
  utmMedium?: string;
  utmCampaign?: string;
  utmTerm?: string;
  utmContent?: string;
  screenWidth?: number;
  screenHeight?: number;
  deviceId?: string;
}

export interface AnalyticsEvent {
  name: string;
  properties?: Record<string, unknown>;
  userId?: string;
  sessionId?: string;
  timestamp: number;
  context?: EventContext;
}

export type UserProperties = Record<string, unknown>;

export interface AutoTrackConfig {
  pageViews?: boolean;
  clicks?: boolean;
  scrollDepth?: boolean;
  formSubmissions?: boolean;
  outboundLinks?: boolean;
  timeOnPage?: boolean;
  errors?: boolean;
}

export interface AnalyticsConfig {
  endpoint: string;
  apiKey?: string;
  autoTrack?: AutoTrackConfig;
  batchSize?: number;
  flushInterval?: number;
  sessionTimeout?: number;
  debug?: boolean;
}
