import type { ISpreadsheetAdapter } from './adapter';
import { AGGridAdapter } from './adapters/ag-grid';
import { UniverAdapter } from './adapters/univer';

export class SpreadsheetFactory {
  static create(type: 'univer' | 'ag-grid'): ISpreadsheetAdapter {
    switch (type) {
      case 'univer':
        return new UniverAdapter();
      case 'ag-grid':
        return new AGGridAdapter();
      default:
        throw new Error(`Unknown spreadsheet adapter type: ${type}`);
    }
  }
}
