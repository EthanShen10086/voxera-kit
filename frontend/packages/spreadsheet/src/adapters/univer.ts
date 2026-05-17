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

export class UniverAdapter implements ISpreadsheetAdapter {
  mount(_container: HTMLElement, _options?: MountOptions): void {
    throw new Error('Univer adapter: install @univerjs/core to use');
  }

  destroy(): void {
    throw new Error('Univer adapter: install @univerjs/core to use');
  }

  setData(_sheets: SheetData[]): void {
    throw new Error('Univer adapter: install @univerjs/core to use');
  }

  getData(): SheetData[] {
    throw new Error('Univer adapter: install @univerjs/core to use');
  }

  getActiveSheet(): string {
    throw new Error('Univer adapter: install @univerjs/core to use');
  }

  setActiveSheet(_sheetId: string): void {
    throw new Error('Univer adapter: install @univerjs/core to use');
  }

  onCellChange(_handler: CellChangeHandler): Unsubscribe {
    throw new Error('Univer adapter: install @univerjs/core to use');
  }

  setCellValue(_sheetId: string, _row: number, _col: number, _value: CellValue): void {
    throw new Error('Univer adapter: install @univerjs/core to use');
  }

  setCellStyle(_sheetId: string, _row: number, _col: number, _style: CellStyle): void {
    throw new Error('Univer adapter: install @univerjs/core to use');
  }

  setColumnWidth(_sheetId: string, _col: number, _width: number): void {
    throw new Error('Univer adapter: install @univerjs/core to use');
  }

  freezePane(_sheetId: string, _rows: number, _cols: number): void {
    throw new Error('Univer adapter: install @univerjs/core to use');
  }

  addConditionalFormat(_sheetId: string, _rule: ConditionalFormatRule): void {
    throw new Error('Univer adapter: install @univerjs/core to use');
  }

  exportXlsx(): Promise<Blob> {
    throw new Error('Univer adapter: install @univerjs/core to use');
  }

  importXlsx(_data: ArrayBuffer): Promise<void> {
    throw new Error('Univer adapter: install @univerjs/core to use');
  }
}
