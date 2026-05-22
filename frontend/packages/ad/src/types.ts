export type SlotType = 'banner' | 'interstitial' | 'native' | 'rewarded' | 'sidebar';

export interface AdContent {
  type: 'html' | 'image' | 'script' | 'json';
  payload: string;
}

export interface Ad {
  id: string;
  slotType: SlotType;
  providerName: string;
  content: AdContent;
  clickUrl: string;
  impressionUrl: string;
  expiresAt: number;
  metadata?: Record<string, unknown>;
}

export interface AdRequest {
  slotType: SlotType;
  userId?: string;
  pageContext?: string;
  locale?: string;
  tags?: string[];
}

export interface AdConfig {
  enabled: boolean;
  minorPolicy: 'hide' | 'safe_only';
  frequencyCap: number;
  fallbackHtml: string;
  providers: ProviderConfig[];
}

export interface ProviderConfig {
  name: string;
  priority: number;
  endpoint?: string;
  publisherId?: string;
}

export interface SlotOptions {
  maxWidth?: number;
  maxHeight?: number;
  refreshSec?: number;
  className?: string;
}
