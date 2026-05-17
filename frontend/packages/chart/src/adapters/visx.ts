import type { IChartAdapter } from '../adapter';
import type { ChartSpec } from '../types';

export class VisxAdapter implements IChartAdapter {
  mount(_container: HTMLElement): void {
    // TODO: initialize visx rendering context
    throw new Error('Visx adapter: install @visx packages to use');
  }

  destroy(): void {
    // TODO: clean up visx rendering context
    throw new Error('Visx adapter: install @visx packages to use');
  }

  resize(): void {
    // TODO: handle container resize
    throw new Error('Visx adapter: install @visx packages to use');
  }

  render(_spec: ChartSpec): void {
    // TODO: render chart from spec using visx primitives
    throw new Error('Visx adapter: install @visx packages to use');
  }

  update(_spec: Partial<ChartSpec>): void {
    // TODO: update chart with partial spec
    throw new Error('Visx adapter: install @visx packages to use');
  }

  exportImage(_type: 'png' | 'svg'): Promise<Blob> {
    // TODO: export visx chart as image blob
    throw new Error('Visx adapter: install @visx packages to use');
  }

  onChartClick(
    _handler: (params: {
      seriesName?: string;
      dataIndex?: number;
      value?: unknown;
    }) => void,
  ): () => void {
    // TODO: register click handler and return unsubscribe fn
    throw new Error('Visx adapter: install @visx packages to use');
  }
}
