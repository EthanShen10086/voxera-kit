export type CellValue = string | number | boolean | null;

export interface CellStyle {
  bold?: boolean;
  italic?: boolean;
  color?: string;
  bgColor?: string;
  align?: string;
  format?: string;
  fontSize?: number;
}

export interface CellData {
  value: CellValue;
  style?: CellStyle;
  formula?: string;
  readOnly?: boolean;
}

export interface SheetData {
  id: string;
  name: string;
  rows: Record<number, Record<number, CellData>>;
  columnWidths?: Record<number, number>;
  frozenRows?: number;
  frozenCols?: number;
}

export interface MountOptions {
  readOnly?: boolean;
  showToolbar?: boolean;
  showFormulaBar?: boolean;
  showSheetTabs?: boolean;
  locale?: string;
}

export interface CellChangeEvent {
  sheetId: string;
  row: number;
  col: number;
  oldValue: CellValue;
  newValue: CellValue;
}

export type CellChangeHandler = (event: CellChangeEvent) => void;

export type Unsubscribe = () => void;

export interface ConditionalFormatRule {
  range: {
    startRow: number;
    endRow: number;
    startCol: number;
    endCol: number;
  };
  type: 'greaterThan' | 'lessThan' | 'between' | 'text';
  value: unknown;
  style: CellStyle;
}
