import type { DataSource } from './types';

export function detectSource(html: string): DataSource {
  if (!html) return 'plain-text';

  if (html.includes('google-sheets-html-origin')) {
    return 'google-sheets';
  }

  if (html.includes('data-lark') || html.includes('lark-record')) {
    return 'feishu';
  }

  if (html.includes('qqDoc') || html.includes('tencentdocs')) {
    return 'tencent-docs';
  }

  if (html.includes('dingtalk')) {
    return 'dingtalk';
  }

  if (
    html.includes('urn:schemas-microsoft-com') ||
    html.includes('xmlns:x')
  ) {
    return 'excel';
  }

  if (/<table[\s>]/i.test(html)) {
    return 'html-table';
  }

  return 'plain-text';
}
