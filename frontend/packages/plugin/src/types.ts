/**
 * Context passed to plugin lifecycle hooks.
 */
export interface PluginContext {
  /** IoC container reference (typed as unknown to avoid coupling to @voxera-kit/di). */
  container?: unknown;
  /** Arbitrary configuration key-value pairs. */
  config?: Record<string, unknown>;
}

/**
 * A plugin that can be registered with the PluginManager.
 * Plugins follow a strict lifecycle: registered → initialized → mounted → destroyed.
 */
export interface IPlugin {
  /** Unique human-readable plugin name. */
  readonly name: string;
  /** Semver version string. */
  readonly version: string;
  /** Names of plugins that must be loaded before this one. */
  readonly dependencies?: string[];

  /** Called once when the plugin is first initialized. */
  onInit?(ctx: PluginContext): Promise<void> | void;
  /** Called after initialization to activate the plugin. */
  onMount?(ctx: PluginContext): Promise<void> | void;
  /** Called when the plugin is being torn down. */
  onDestroy?(ctx: PluginContext): Promise<void> | void;
}

/** Possible states a plugin can be in. */
export type PluginState = "registered" | "initialized" | "mounted" | "destroyed";

/** Runtime metadata about a registered plugin. */
export interface PluginInfo {
  /** The plugin instance. */
  plugin: IPlugin;
  /** Current lifecycle state. */
  state: PluginState;
  /** Timestamp when the plugin was registered. */
  registeredAt: number;
  /** Timestamp of the last state transition. */
  lastTransitionAt: number;
}
