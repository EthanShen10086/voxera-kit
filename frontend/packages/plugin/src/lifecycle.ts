import type { IPlugin, PluginContext, PluginState } from "./types.js";

/** Enum representation of plugin lifecycle phases. */
export enum PluginLifecycle {
  Registered = "registered",
  Initialized = "initialized",
  Mounted = "mounted",
  Destroyed = "destroyed",
}

/** Valid state transitions. */
const VALID_TRANSITIONS: Record<PluginState, PluginState[]> = {
  registered: ["initialized", "destroyed"],
  initialized: ["mounted", "destroyed"],
  mounted: ["destroyed"],
  destroyed: [],
};

/**
 * Check whether a lifecycle state transition is valid.
 * @returns `true` if transitioning from `from` to `to` is allowed.
 */
export function validateTransition(from: PluginState, to: PluginState): boolean {
  return VALID_TRANSITIONS[from]?.includes(to) ?? false;
}

type LifecycleHook = "onInit" | "onMount" | "onDestroy";

/**
 * Safely execute a lifecycle hook on a plugin.
 * Catches and wraps errors with the plugin name for easier debugging.
 */
export async function executeLifecycleHook(
  plugin: IPlugin,
  hook: LifecycleHook,
  ctx: PluginContext,
): Promise<void> {
  const fn = plugin[hook];
  if (typeof fn !== "function") return;

  try {
    await fn.call(plugin, ctx);
  } catch (error) {
    const message =
      error instanceof Error ? error.message : String(error);
    throw new Error(
      `[PluginManager] Lifecycle hook "${hook}" failed for plugin "${plugin.name}": ${message}`,
    );
  }
}
