import { PlayerEvent } from "../events.js";
import type { IPlayerAdapter, PlayerEventHandler, PlayerOptions } from "../types.js";

// TODO: import Player from 'xgplayer' once the dependency is added

/** Map internal PlayerEvent → xgplayer native event names. */
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
  [PlayerEvent.LoadedMetadata]: "loadeddata",
  [PlayerEvent.DurationChange]: "durationchange",
  [PlayerEvent.BufferUpdate]: "progress",
};

/**
 * XGPlayer adapter – stub implementation.
 * Wire up xgplayer SDK methods once the dependency is installed.
 */
export class XGPlayerAdapter implements IPlayerAdapter {
  private player: unknown = null;
  private listeners = new Map<string, Set<PlayerEventHandler>>();

  async init(_container: HTMLElement, _options?: PlayerOptions): Promise<void> {
    // TODO: instantiate xgplayer Player
    throw new Error("XGPlayerAdapter.init() is not yet implemented.");
  }

  async play(): Promise<void> {
    // TODO: this.player.play()
    throw new Error("XGPlayerAdapter.play() is not yet implemented.");
  }

  async pause(): Promise<void> {
    // TODO: this.player.pause()
    throw new Error("XGPlayerAdapter.pause() is not yet implemented.");
  }

  async seek(_time: number): Promise<void> {
    // TODO: this.player.currentTime = time
    throw new Error("XGPlayerAdapter.seek() is not yet implemented.");
  }

  getCurrentTime(): number {
    // TODO: return this.player.currentTime
    return 0;
  }

  getDuration(): number {
    // TODO: return this.player.duration
    return 0;
  }

  getVolume(): number {
    // TODO: return this.player.volume
    return 1;
  }

  setVolume(_volume: number): void {
    // TODO: this.player.volume = volume
  }

  getPlaybackRate(): number {
    // TODO: return this.player.playbackRate
    return 1;
  }

  setPlaybackRate(_rate: number): void {
    // TODO: this.player.playbackRate = rate
  }

  on(event: PlayerEvent, handler: PlayerEventHandler): void {
    const nativeEvent = EVENT_MAP[event];
    if (!this.listeners.has(nativeEvent)) {
      this.listeners.set(nativeEvent, new Set());
    }
    this.listeners.get(nativeEvent)!.add(handler);
    // TODO: this.player.on(nativeEvent, handler)
  }

  off(event: PlayerEvent, handler: PlayerEventHandler): void {
    const nativeEvent = EVENT_MAP[event];
    this.listeners.get(nativeEvent)?.delete(handler);
    // TODO: this.player.off(nativeEvent, handler)
  }

  destroy(): void {
    this.listeners.clear();
    // TODO: this.player.destroy()
    this.player = null;
  }
}
