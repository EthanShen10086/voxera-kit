export interface MetaTag {
  name?: string;
  property?: string;
  content: string;
}

export interface OpenGraphData {
  title: string;
  description: string;
  image?: string;
  url?: string;
  type?: string;
  siteName?: string;
}

export interface TwitterCardData {
  card: "summary" | "summary_large_image";
  title: string;
  description: string;
  image?: string;
}

export interface StructuredData {
  type: string;
  data: Record<string, unknown>;
}

export interface PageSEO {
  title: string;
  description?: string;
  canonical?: string;
  robots?: string;
  openGraph?: OpenGraphData;
  twitter?: TwitterCardData;
  structuredData?: StructuredData[];
  meta?: MetaTag[];
}

export interface ISEOManager {
  setPage(seo: PageSEO): void;
  getMetaTags(): MetaTag[];
  getTitle(): string;
  toJSON(): Record<string, unknown>;
}

export interface SitemapEntry {
  url: string;
  lastmod?: string;
  changefreq?: string;
  priority?: number;
}

export interface ISitemapGenerator {
  addEntry(entry: SitemapEntry): void;
  generate(): string;
}
