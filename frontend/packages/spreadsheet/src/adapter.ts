import type {
  CellChangeHandler,
  CellStyle,
  CellValue,
  ConditionalFormatRule,
  MountOptions,
  SheetData,
  Unsubscribe,
} from './types';

export interface ISpreadsheetAdapter {
  mount(container: HTMLElement, options?: MountOptions): void;
  destroy(): void;

  setData(sheets: SheetData[]): void;
  getData(): SheetData[];

  getActiveSheet(): string;
  setActiveSheet(sheetId: string): void;

  onCellChange(handler: CellChangeHandler): Unsubscribe;

  setCellValue(sheetId: string, row: number, col: number, value: CellValue): void;
  setCellStyle(sheetId: string, row: number, col: number, style: CellStyle): void;
  setColumnWidth(sheetId: string, col: number, width: number): void;

  freezePane(sheetId: string, rows: number, cols: number): void;
  addConditionalFormat(sheetId: string, rule: ConditionalFormatRule): void;

  exportXlsx(): Promise<Blob>;
  importXlsx(data: ArrayBuffer): Promise<void>;
}
