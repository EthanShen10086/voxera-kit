import type { FeedFilter, FeedItem, FeedPage } from './types';

export interface IFeedAdapter {
  fetchFeed(filter?: FeedFilter, cursor?: string, limit?: number): Promise<FeedPage>;
  subscribe(callback: (item: FeedItem) => void): () => void;
}
