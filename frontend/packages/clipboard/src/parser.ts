import { detectSource } from './detector';
import type {
  ClipboardTableData,
  DataSource,
  IClipboardParser,
  ParsedCell,
} from './types';

const DATE_ISO = /^\d{4}-\d{2}-\d{2}$/;
const DATE_US = /^\d{1,2}\/\d{1,2}\/\d{2,4}$/;
const DATE_VERBOSE =
  /^(?:Jan|Feb|Mar|Apr|May|Jun|Jul|Aug|Sep|Oct|Nov|Dec)\s+\d{1,2},?\s+\d{4}$/i;

const NUMBER_PLAIN = /^-?\d[\d,]*(\.\d+)?$/;
const NUMBER_PARENS = /^\([\d,]+(\.\d+)?\)$/;

const PERCENTAGE = /^-?\d+(\.\d+)?%$/;

export class ClipboardParser implements IClipboardParser {
  parse(event: ClipboardEvent): ClipboardTableData {
    const clipboardData = event.clipboardData;
    if (!clipboardData) {
      return this.emptyResult('unknown');
    }

    const html = clipboardData.getData('text/html');
    if (html) {
      return this.parseHtml(html);
    }

    const text = clipboardData.getData('text/plain');
    return this.parseText(text);
  }

  parseText(text: string): ClipboardTableData {
    if (!text.trim()) {
      return this.emptyResult('plain-text');
    }

    const lines = text.split(/\r?\n/).filter((line) => line.length > 0);
    const rows: ParsedCell[][] = lines.map((line) =>
      line.split('\t').map((cell) => this.inferType(cell)),
    );

    const colCount = Math.max(0, ...rows.map((r) => r.length));

    return {
      rows,
      source: 'plain-text',
      rowCount: rows.length,
      colCount,
    };
  }

  parseHtml(html: string): ClipboardTableData {
    const source = detectSource(html);

    const doc = new DOMParser().parseFromString(html, 'text/html');
    const table = doc.querySelector('table');

    if (!table) {
      const text = doc.body.textContent ?? '';
      return this.parseText(text);
    }

    const rows: ParsedCell[][] = [];
    const headers: string[] = [];

    const theadRows = table.querySelectorAll('thead tr');
    if (theadRows.length > 0) {
      const headerRow = theadRows[theadRows.length - 1];
      headerRow?.querySelectorAll('th, td').forEach((cell) => {
        headers.push((cell.textContent ?? '').trim());
      });
    }

    const bodyRows = table.querySelectorAll('tbody tr');
    const targetRows = bodyRows.length > 0 ? bodyRows : table.querySelectorAll('tr');

    targetRows.forEach((tr) => {
      const cells: ParsedCell[] = [];
      tr.querySelectorAll('td, th').forEach((td) => {
        const raw = (td.textContent ?? '').trim();
        cells.push(this.inferType(raw));
      });
      if (cells.length > 0) {
        rows.push(cells);
      }
    });

    if (headers.length === 0 && rows.length > 0) {
      const firstRowAllStrings = rows[0]!.every(
        (c) => c.type === 'string' || c.type === 'empty',
      );
      if (firstRowAllStrings && rows.length > 1) {
        const shifted = rows.shift()!;
        for (const cell of shifted) {
          headers.push(String(cell.raw));
        }
      }
    }

    const colCount = Math.max(
      0,
      headers.length,
      ...rows.map((r) => r.length),
    );

    return {
      rows,
      ...(headers.length > 0 ? { headers } : {}),
      source,
      rowCount: rows.length,
      colCount,
    };
  }

  detectSource(html: string): DataSource {
    return detectSource(html);
  }

  private inferType(raw: string): ParsedCell {
    const trimmed = raw.trim();

    if (trimmed === '') {
      return { raw, value: null, type: 'empty' };
    }

    const lower = trimmed.toLowerCase();
    if (lower === 'true' || lower === 'false') {
      return { raw, value: lower === 'true', type: 'boolean' };
    }

    if (PERCENTAGE.test(trimmed)) {
      const pct = this.parsePercentage(trimmed);
      if (pct !== null) {
        return { raw, value: pct, type: 'percentage' };
      }
    }

    if (NUMBER_PLAIN.test(trimmed) || NUMBER_PARENS.test(trimmed)) {
      const num = this.parseNumber(trimmed);
      if (num !== null) {
        return { raw, value: num, type: 'number' };
      }
    }

    if (
      DATE_ISO.test(trimmed) ||
      DATE_US.test(trimmed) ||
      DATE_VERBOSE.test(trimmed)
    ) {
      const ts = Date.parse(trimmed);
      if (!Number.isNaN(ts)) {
        return { raw, value: trimmed, type: 'date' };
      }
    }

    return { raw, value: trimmed, type: 'string' };
  }

  private parseNumber(raw: string): number | null {
    let str = raw.trim();

    const isParens = NUMBER_PARENS.test(str);
    if (isParens) {
      str = str.slice(1, -1);
    }

    str = str.replace(/,/g, '');
    const num = Number(str);

    if (Number.isNaN(num)) return null;
    return isParens ? -num : num;
  }

  private parsePercentage(raw: string): number | null {
    const str = raw.trim().slice(0, -1);
    const num = Number(str);
    if (Number.isNaN(num)) return null;
    return num / 100;
  }

  private emptyResult(source: DataSource): ClipboardTableData {
    return { rows: [], source, rowCount: 0, colCount: 0 };
  }
}
