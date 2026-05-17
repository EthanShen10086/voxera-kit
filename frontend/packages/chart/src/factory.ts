import type { IChartAdapter } from './adapter';
import { EChartsAdapter } from './adapters/echarts';
import { VisxAdapter } from './adapters/visx';

export class ChartFactory {
  static create(type: 'echarts' | 'visx'): IChartAdapter {
    switch (type) {
      case 'echarts':
        return new EChartsAdapter();
      case 'visx':
        return new VisxAdapter();
      default:
        throw new Error(`Unknown chart adapter type: ${type as string}`);
    }
  }
}
