import type { Ad, AdRequest } from './types';

export interface IAdProvider {
  name: string;
  fetchAd(request: AdRequest): Promise<Ad | null>;
  reportImpression(adId: string): Promise<void>;
  reportClick(adId: string): Promise<void>;
  available(): boolean;
}
