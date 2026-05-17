import type { ChartSpec } from './types';

export interface IChartAdapter {
  mount(container: HTMLElement): void;
  destroy(): void;
  resize(): void;
  render(spec: ChartSpec): void;
  update(spec: Partial<ChartSpec>): void;
  exportImage(type: 'png' | 'svg'): Promise<Blob>;
  onChartClick(
    handler: (params: {
      seriesName?: string;
      dataIndex?: number;
      value?: unknown;
    }) => void,
  ): () => void;
}
