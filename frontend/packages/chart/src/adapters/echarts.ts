import type { IChartAdapter } from '../adapter';
import type { ChartSpec } from '../types';

export interface EChartsAdapterOptions {
  renderer?: 'canvas' | 'svg';
}

export class EChartsAdapter implements IChartAdapter {
  constructor(_options: EChartsAdapterOptions = {}) {
    // Options stored for future echarts init
  }

  mount(_container: HTMLElement): void {
    // TODO: initialize echarts instance on container
    throw new Error('ECharts adapter: install echarts to use');
  }

  destroy(): void {
    // TODO: dispose echarts instance
    throw new Error('ECharts adapter: install echarts to use');
  }

  resize(): void {
    // TODO: call echarts instance resize
    throw new Error('ECharts adapter: install echarts to use');
  }

  render(spec: ChartSpec): void {
    // TODO: convert spec via mapSpecToOption and call setOption
    void spec;
    throw new Error('ECharts adapter: install echarts to use');
  }

  update(spec: Partial<ChartSpec>): void {
    // TODO: merge partial spec and re-render
    void spec;
    throw new Error('ECharts adapter: install echarts to use');
  }

  exportImage(_type: 'png' | 'svg'): Promise<Blob> {
    // TODO: export chart as image blob
    throw new Error('ECharts adapter: install echarts to use');
  }

  onChartClick(
    _handler: (params: {
      seriesName?: string;
      dataIndex?: number;
      value?: unknown;
    }) => void,
  ): () => void {
    // TODO: register click handler and return unsubscribe fn
    throw new Error('ECharts adapter: install echarts to use');
  }

}
