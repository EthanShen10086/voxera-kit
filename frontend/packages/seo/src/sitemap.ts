import type { ISitemapGenerator, SitemapEntry } from "./types.js";

export class SitemapGenerator implements ISitemapGenerator {
  private readonly entries: SitemapEntry[] = [];

  addEntry(entry: SitemapEntry): void {
    this.entries.push(entry);
  }

  generate(): string {
    const urls = this.entries.map((entry) => {
      let xml = `  <url>\n    <loc>${escapeXml(entry.url)}</loc>`;
      if (entry.lastmod) xml += `\n    <lastmod>${escapeXml(entry.lastmod)}</lastmod>`;
      if (entry.changefreq) xml += `\n    <changefreq>${escapeXml(entry.changefreq)}</changefreq>`;
      if (entry.priority != null) xml += `\n    <priority>${entry.priority}</priority>`;
      xml += "\n  </url>";
      return xml;
    });

    return [
      '<?xml version="1.0" encoding="UTF-8"?>',
      '<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">',
      ...urls,
      "</urlset>",
    ].join("\n");
  }
}

function escapeXml(str: string): string {
  return str
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
    .replace(/"/g, "&quot;")
    .replace(/'/g, "&apos;");
}
