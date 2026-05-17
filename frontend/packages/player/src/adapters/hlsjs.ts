import { PlayerEvent } from "../events.js";
import type { IPlayerAdapter, PlayerEventHandler, PlayerOptions } from "../types.js";

// TODO: import Hls from 'hls.js' once the dependency is added

/** Map internal PlayerEvent → native HTML5 video event names used with hls.js. */
const EVENT_MAP: Record<PlayerEvent, string> = {
  [PlayerEvent.Play]: "play",
  [PlayerEvent.Pause]: "pause",
  [PlayerEvent.TimeUpdate]: "timeupdate",
  [PlayerEvent.Ended]: "ended",
  [PlayerEvent.Error]: "error",
  [PlayerEvent.Seeking]: "seeking",
  [PlayerEvent.Seeked]: "seeked",
  [PlayerEvent.VolumeChange]: "volumechange",
  [PlayerEvent.RateChange]: "ratechange",
  [PlayerEvent.Waiting]: "waiting",
  [PlayerEvent.CanPlay]: "canplay",
  [PlayerEvent.LoadedMetadata]: "loadedmetadata",
  [PlayerEvent.DurationChange]: "durationchange",
  [PlayerEvent.BufferUpdate]: "progress",
};

/**
 * HLS.js adapter – stub implementation.
 * Uses a raw <video> element with hls.js for HLS stream support.
 */
export class HLSJSAdapter implements IPlayerAdapter {
  private videoElement: HTMLVideoElement | null = null;
  private listeners = new Map<string, Set<PlayerEventHandler>>();

  async init(_container: HTMLElement, _options?: PlayerOptions): Promise<void> {
    // TODO: create <video>, attach Hls instance, loadSource
    throw new Error("HLSJSAdapter.init() is not yet implemented.");
  }

  async play(): Promise<void> {
    // TODO: this.videoElement.play()
    throw new Error("HLSJSAdapter.play() is not yet implemented.");
  }

  async pause(): Promise<void> {
    // TODO: this.videoElement.pause()
    throw new Error("HLSJSAdapter.pause() is not yet implemented.");
  }

  async seek(_time: number): Promise<void> {
    // TODO: this.videoElement.currentTime = time
    throw new Error("HLSJSAdapter.seek() is not yet implemented.");
  }

  getCurrentTime(): number {
    return this.videoElement?.currentTime ?? 0;
  }

  getDuration(): number {
    return this.videoElement?.duration ?? 0;
  }

  getVolume(): number {
    return this.videoElement?.volume ?? 1;
  }

  setVolume(volume: number): void {
    if (this.videoElement) {
      this.videoElement.volume = volume;
    }
  }

  getPlaybackRate(): number {
    return this.videoElement?.playbackRate ?? 1;
  }

  setPlaybackRate(rate: number): void {
    if (this.videoElement) {
      this.videoElement.playbackRate = rate;
    }
  }

  on(event: PlayerEvent, handler: PlayerEventHandler): void {
    const nativeEvent = EVENT_MAP[event];
    if (!this.listeners.has(nativeEvent)) {
      this.listeners.set(nativeEvent, new Set());
    }
    this.listeners.get(nativeEvent)!.add(handler);
    this.videoElement?.addEventListener(nativeEvent, handler as EventListener);
  }

  off(event: PlayerEvent, handler: PlayerEventHandler): void {
    const nativeEvent = EVENT_MAP[event];
    this.listeners.get(nativeEvent)?.delete(handler);
    this.videoElement?.removeEventListener(nativeEvent, handler as EventListener);
  }

  destroy(): void {
    this.listeners.clear();
    // TODO: destroy hls.js instance
    if (this.videoElement) {
      this.videoElement.pause();
      this.videoElement.removeAttribute("src");
      this.videoElement.load();
      this.videoElement.remove();
      this.videoElement = null;
    }
  }
}
