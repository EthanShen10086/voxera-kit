/** Standardized player events across all adapter implementations. */
export enum PlayerEvent {
  Play = "play",
  Pause = "pause",
  TimeUpdate = "timeupdate",
  Ended = "ended",
  Error = "error",
  Seeking = "seeking",
  Seeked = "seeked",
  VolumeChange = "volumechange",
  RateChange = "ratechange",
  Waiting = "waiting",
  CanPlay = "canplay",
  LoadedMetadata = "loadedmetadata",
  DurationChange = "durationchange",
  BufferUpdate = "bufferupdate",
}
