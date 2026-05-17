export type CellType =
  | 'string'
  | 'number'
  | 'percentage'
  | 'date'
  | 'boolean'
  | 'empty';

export interface ParsedCell {
  raw: string;
  value: string | number | boolean | null;
  type: CellType;
}

export type DataSource =
  | 'excel'
  | 'google-sheets'
  | 'feishu'
  | 'tencent-docs'
  | 'dingtalk'
  | 'html-table'
  | 'plain-text'
  | 'unknown';

export interface ClipboardTableData {
  rows: ParsedCell[][];
  headers?: string[];
  source: DataSource;
  rowCount: number;
  colCount: number;
}

export interface IClipboardParser {
  parse(event: ClipboardEvent): ClipboardTableData;
  parseText(text: string): ClipboardTableData;
  parseHtml(html: string): ClipboardTableData;
  detectSource(html: string): DataSource;
}
