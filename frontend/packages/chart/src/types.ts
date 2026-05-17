export interface DataPoint {
  label: string;
  value: number;
  color?: string;
}

export interface SeriesData {
  name: string;
  data: DataPoint[];
  type?: string;
}

export interface ChartTheme {
  colors: string[];
  backgroundColor?: string;
  fontFamily?: string;
  fontSize?: number;
}

export interface BaseChartSpec {
  title?: string;
  theme?: ChartTheme;
  width?: number;
  height?: number;
  animation?: boolean;
}

export interface LineChartSpec extends BaseChartSpec {
  type: 'line';
  series: SeriesData[];
  xAxis?: { label?: string; categories?: string[] };
  yAxis?: { label?: string; min?: number; max?: number };
  smooth?: boolean;
}

export interface BarChartSpec extends BaseChartSpec {
  type: 'bar';
  series: SeriesData[];
  xAxis?: { label?: string; categories?: string[] };
  yAxis?: { label?: string };
  stacked?: boolean;
  horizontal?: boolean;
}

export interface PieChartSpec extends BaseChartSpec {
  type: 'pie';
  data: DataPoint[];
  innerRadius?: number;
  showLabels?: boolean;
  showLegend?: boolean;
}

export interface SankeyNode {
  id: string;
  name: string;
  color?: string;
}

export interface SankeyLink {
  source: string;
  target: string;
  value: number;
  color?: string;
}

export interface SankeyChartSpec extends BaseChartSpec {
  type: 'sankey';
  nodes: SankeyNode[];
  links: SankeyLink[];
  nodeWidth?: number;
  nodePadding?: number;
}

export interface AreaChartSpec extends BaseChartSpec {
  type: 'area';
  series: SeriesData[];
  xAxis?: { label?: string; categories?: string[] };
  yAxis?: { label?: string };
  stacked?: boolean;
}

export interface WaterfallItem {
  label: string;
  value: number;
  isTotal?: boolean;
}

export interface WaterfallChartSpec extends BaseChartSpec {
  type: 'waterfall';
  data: WaterfallItem[];
  positiveColor?: string;
  negativeColor?: string;
  totalColor?: string;
}

export interface RadarIndicator {
  name: string;
  max: number;
}

export interface RadarSeriesData {
  name: string;
  values: number[];
}

export interface RadarChartSpec extends BaseChartSpec {
  type: 'radar';
  indicators: RadarIndicator[];
  series: RadarSeriesData[];
}

export type ChartSpec =
  | LineChartSpec
  | BarChartSpec
  | PieChartSpec
  | SankeyChartSpec
  | AreaChartSpec
  | WaterfallChartSpec
  | RadarChartSpec;
