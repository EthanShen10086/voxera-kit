import type { FeedItem, FeedViewMode } from './types';

export interface FeedRenderOptions {
  viewMode: FeedViewMode;
  showMedia: boolean;
  showPlatformBadge: boolean;
  timeFormat: 'relative' | 'absolute';
}

export interface IFeedRenderer {
  render(items: FeedItem[], options: FeedRenderOptions): unknown;
}
