export interface FeedItem {
  id: string;
  platformId: string;
  platform: 'twitter' | 'bilibili' | string;
  authorId: string;
  authorName: string;
  authorAvatar?: string;
  content: string;
  mediaUrls: string[];
  publishedAt: Date;
  fetchedAt: Date;
  url: string;
  metadata?: Record<string, unknown>;
}

export type FeedViewMode = 'timeline' | 'card' | 'compact';

export interface FeedFilter {
  platforms?: string[];
  authors?: string[];
  keywords?: string[];
  dateFrom?: Date;
  dateTo?: Date;
}

export interface FeedPagination {
  cursor?: string;
  limit: number;
  hasMore: boolean;
}

export interface FeedPage {
  items: FeedItem[];
  pagination: FeedPagination;
}
