interface TrackEvent {
  type: 'impression' | 'click';
  adId: string;
  timestamp: number;
}

export class AdTracker {
  private buffer: TrackEvent[] = [];
  private endpoint: string;
  private flushInterval: ReturnType<typeof setInterval> | null = null;
  private readonly batchSize: number;

  constructor(endpoint: string, opts?: { flushMs?: number; batchSize?: number }) {
    this.endpoint = endpoint;
    this.batchSize = opts?.batchSize ?? 10;

    this.flushInterval = setInterval(() => this.flush(), opts?.flushMs ?? 5000);

    if (typeof document !== 'undefined') {
      document.addEventListener('visibilitychange', () => {
        if (document.visibilityState === 'hidden') {
          this.flush(true);
        }
      });
    }
  }

  trackImpression(adId: string): void {
    this.buffer.push({ type: 'impression', adId, timestamp: Date.now() });
    if (this.buffer.length >= this.batchSize) this.flush();
  }

  trackClick(adId: string): void {
    this.buffer.push({ type: 'click', adId, timestamp: Date.now() });
    this.flush();
  }

  dispose(): void {
    if (this.flushInterval) {
      clearInterval(this.flushInterval);
      this.flushInterval = null;
    }
    this.flush(true);
  }

  private flush(keepalive = false): void {
    if (this.buffer.length === 0) return;

    const events = this.buffer.splice(0);

    fetch(this.endpoint, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ events }),
      keepalive,
    }).catch(() => {
      this.buffer.unshift(...events);
    });
  }
}
