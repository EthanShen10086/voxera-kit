import type { ISEOManager, MetaTag, PageSEO } from "./types.js";

export class SEOManager implements ISEOManager {
  private page: PageSEO = { title: "" };

  setPage(seo: PageSEO): void {
    this.page = seo;
  }

  getTitle(): string {
    return this.page.title;
  }

  getMetaTags(): MetaTag[] {
    const tags: MetaTag[] = [];

    if (this.page.description) {
      tags.push({ name: "description", content: this.page.description });
    }

    if (this.page.robots) {
      tags.push({ name: "robots", content: this.page.robots });
    }

    if (this.page.canonical) {
      tags.push({ name: "canonical", content: this.page.canonical });
    }

    if (this.page.openGraph) {
      const og = this.page.openGraph;
      tags.push({ property: "og:title", content: og.title });
      tags.push({ property: "og:description", content: og.description });
      if (og.image) tags.push({ property: "og:image", content: og.image });
      if (og.url) tags.push({ property: "og:url", content: og.url });
      if (og.type) tags.push({ property: "og:type", content: og.type });
      if (og.siteName) tags.push({ property: "og:site_name", content: og.siteName });
    }

    if (this.page.twitter) {
      const tw = this.page.twitter;
      tags.push({ name: "twitter:card", content: tw.card });
      tags.push({ name: "twitter:title", content: tw.title });
      tags.push({ name: "twitter:description", content: tw.description });
      if (tw.image) tags.push({ name: "twitter:image", content: tw.image });
    }

    if (this.page.meta) {
      tags.push(...this.page.meta);
    }

    return tags;
  }

  toJSON(): Record<string, unknown> {
    return {
      title: this.page.title,
      meta: this.getMetaTags(),
      structuredData: this.page.structuredData ?? [],
    };
  }
}
