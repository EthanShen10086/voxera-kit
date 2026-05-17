import type { ISpreadsheetAdapter } from '../adapter';
import type {
  CellChangeHandler,
  CellStyle,
  CellValue,
  ConditionalFormatRule,
  MountOptions,
  SheetData,
  Unsubscribe,
} from '../types';

export class AGGridAdapter implements ISpreadsheetAdapter {
  mount(_container: HTMLElement, _options?: MountOptions): void {
    throw new Error('AG Grid adapter: install ag-grid-community to use');
  }

  destroy(): void {
    throw new Error('AG Grid adapter: install ag-grid-community to use');
  }

  setData(_sheets: SheetData[]): void {
    throw new Error('AG Grid adapter: install ag-grid-community to use');
  }

  getData(): SheetData[] {
    throw new Error('AG Grid adapter: install ag-grid-community to use');
  }

  getActiveSheet(): string {
    throw new Error('AG Grid adapter: install ag-grid-community to use');
  }

  setActiveSheet(_sheetId: string): void {
    throw new Error('AG Grid adapter: install ag-grid-community to use');
  }

  onCellChange(_handler: CellChangeHandler): Unsubscribe {
    throw new Error('AG Grid adapter: install ag-grid-community to use');
  }

  setCellValue(_sheetId: string, _row: number, _col: number, _value: CellValue): void {
    throw new Error('AG Grid adapter: install ag-grid-community to use');
  }

  setCellStyle(_sheetId: string, _row: number, _col: number, _style: CellStyle): void {
    throw new Error('AG Grid adapter: install ag-grid-community to use');
  }

  setColumnWidth(_sheetId: string, _col: number, _width: number): void {
    throw new Error('AG Grid adapter: install ag-grid-community to use');
  }

  freezePane(_sheetId: string, _rows: number, _cols: number): void {
    throw new Error('AG Grid adapter: install ag-grid-community to use');
  }

  addConditionalFormat(_sheetId: string, _rule: ConditionalFormatRule): void {
    throw new Error('AG Grid adapter: install ag-grid-community to use');
  }

  exportXlsx(): Promise<Blob> {
    throw new Error('AG Grid adapter: install ag-grid-community to use');
  }

  importXlsx(_data: ArrayBuffer): Promise<void> {
    throw new Error('AG Grid adapter: install ag-grid-community to use');
  }
}
