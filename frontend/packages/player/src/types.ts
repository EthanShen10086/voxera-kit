import type { PlayerEvent } from "./events.js";

/** Supported player engine types. */
export type PlayerType = "xgplayer" | "videojs" | "hlsjs";

/** Handler signature for player events. */
export type PlayerEventHandler = (data?: unknown) => void;

/** Options for initializing a player adapter. */
export interface PlayerOptions {
  /** Media source URL. */
  src?: string;
  /** Whether the video should start playing automatically. */
  autoplay?: boolean;
  /** Whether the video is muted initially. */
  muted?: boolean;
  /** Initial volume (0–1). */
  volume?: number;
  /** Initial playback rate. */
  playbackRate?: number;
  /** Whether to show native controls. */
  controls?: boolean;
  /** Poster image URL. */
  poster?: string;
  /** Whether the video should loop. */
  loop?: boolean;
  /** Additional adapter-specific options. */
  [key: string]: unknown;
}

/**
 * Unified player adapter interface.
 * All player engine integrations must implement this contract.
 */
export interface IPlayerAdapter {
  /** Initialize the player inside the given container element. */
  init(container: HTMLElement, options?: PlayerOptions): Promise<void>;

  /** Start or resume playback. */
  play(): Promise<void>;

  /** Pause playback. */
  pause(): Promise<void>;

  /** Seek to a specific time in seconds. */
  seek(time: number): Promise<void>;

  /** Get the current playback position in seconds. */
  getCurrentTime(): number;

  /** Get the total media duration in seconds. */
  getDuration(): number;

  /** Get the current volume (0–1). */
  getVolume(): number;

  /** Set the volume (0–1). */
  setVolume(volume: number): void;

  /** Get the current playback rate. */
  getPlaybackRate(): number;

  /** Set the playback rate (e.g. 1.0 = normal, 2.0 = double speed). */
  setPlaybackRate(rate: number): void;

  /** Subscribe to a player event. */
  on(event: PlayerEvent, handler: PlayerEventHandler): void;

  /** Unsubscribe from a player event. */
  off(event: PlayerEvent, handler: PlayerEventHandler): void;

  /** Tear down the player and release resources. */
  destroy(): void;
}
