import type { IPlayerAdapter, PlayerOptions, PlayerType } from "./types.js";

/**
 * Factory for creating player adapter instances by type.
 * Adapters are lazily imported to avoid bundling unused engines.
 */
export class PlayerFactory {
  static async createPlayer(
    type: PlayerType,
    _options?: PlayerOptions,
  ): Promise<IPlayerAdapter> {
    switch (type) {
      case "xgplayer": {
        const { XGPlayerAdapter } = await import("./adapters/xgplayer.js");
        return new XGPlayerAdapter();
      }
      case "videojs": {
        const { VideoJSAdapter } = await import("./adapters/videojs.js");
        return new VideoJSAdapter();
      }
      case "hlsjs": {
        const { HLSJSAdapter } = await import("./adapters/hlsjs.js");
        return new HLSJSAdapter();
      }
      default:
        throw new Error(
          `[PlayerFactory] Unknown player type: "${type as string}". ` +
            `Supported types: xgplayer, videojs, hlsjs`,
        );
    }
  }
}
