import type { IPlugin, PluginContext, PluginInfo, PluginState } from "./types.js";
import { executeLifecycleHook, validateTransition } from "./lifecycle.js";

type LifecycleEventHandler = (name: string, state: PluginState) => void;

interface PluginEntry {
  plugin: IPlugin;
  state: PluginState;
  registeredAt: number;
  lastTransitionAt: number;
}

/**
 * Manages plugin registration, lifecycle, and dependency resolution.
 */
export class PluginManager {
  private readonly plugins = new Map<string, PluginEntry>();
  private readonly listeners = new Set<LifecycleEventHandler>();

  /** Register a plugin. Throws if dependencies are unknown at registration time only when they are checked. */
  register(plugin: IPlugin): void {
    if (this.plugins.has(plugin.name)) {
      throw new Error(
        `[PluginManager] Plugin "${plugin.name}" is already registered.`,
      );
    }

    const now = Date.now();
    this.plugins.set(plugin.name, {
      plugin,
      state: "registered",
      registeredAt: now,
      lastTransitionAt: now,
    });

    this.emit(plugin.name, "registered");
  }

  /** Unregister a plugin, calling onDestroy first if it has been mounted. */
  async unregister(name: string, ctx: PluginContext = {}): Promise<void> {
    const entry = this.plugins.get(name);
    if (!entry) {
      throw new Error(
        `[PluginManager] Plugin "${name}" is not registered.`,
      );
    }

    if (entry.state === "mounted" || entry.state === "initialized") {
      await this.transitionTo(entry, "destroyed", ctx);
    }

    this.plugins.delete(name);
  }

  /** Initialize all registered plugins in dependency order. */
  async init(ctx: PluginContext = {}): Promise<void> {
    const sorted = this.topologicalSort();

    for (const name of sorted) {
      const entry = this.plugins.get(name)!;
      if (entry.state === "registered") {
        await this.transitionTo(entry, "initialized", ctx);
      }
    }
  }

  /** Mount all initialized plugins in dependency order. */
  async mount(ctx: PluginContext = {}): Promise<void> {
    const sorted = this.topologicalSort();

    for (const name of sorted) {
      const entry = this.plugins.get(name)!;
      if (entry.state === "initialized") {
        await this.transitionTo(entry, "mounted", ctx);
      }
    }
  }

  /** Destroy all plugins in reverse dependency order. */
  async destroy(ctx: PluginContext = {}): Promise<void> {
    const sorted = this.topologicalSort().reverse();

    for (const name of sorted) {
      const entry = this.plugins.get(name);
      if (entry && entry.state !== "destroyed") {
        await this.transitionTo(entry, "destroyed", ctx);
      }
    }
  }

  /** Retrieve a plugin by name. */
  get(name: string): IPlugin | undefined {
    return this.plugins.get(name)?.plugin;
  }

  /** Get info for all registered plugins. */
  getAll(): PluginInfo[] {
    return Array.from(this.plugins.values()).map(
      ({ plugin, state, registeredAt, lastTransitionAt }) => ({
        plugin,
        state,
        registeredAt,
        lastTransitionAt,
      }),
    );
  }

  /** Subscribe to lifecycle state changes. Returns an unsubscribe function. */
  onLifecycleChange(handler: LifecycleEventHandler): () => void {
    this.listeners.add(handler);
    return () => {
      this.listeners.delete(handler);
    };
  }

  // ── Private helpers ──────────────────────────────────────────────

  private async transitionTo(
    entry: PluginEntry,
    target: PluginState,
    ctx: PluginContext,
  ): Promise<void> {
    if (!validateTransition(entry.state, target)) {
      throw new Error(
        `[PluginManager] Invalid transition for "${entry.plugin.name}": ` +
          `${entry.state} → ${target}`,
      );
    }

    const hookMap: Record<string, "onInit" | "onMount" | "onDestroy"> = {
      initialized: "onInit",
      mounted: "onMount",
      destroyed: "onDestroy",
    };

    const hook = hookMap[target];
    if (hook) {
      await executeLifecycleHook(entry.plugin, hook, ctx);
    }

    entry.state = target;
    entry.lastTransitionAt = Date.now();
    this.emit(entry.plugin.name, target);
  }

  private emit(name: string, state: PluginState): void {
    for (const handler of this.listeners) {
      try {
        handler(name, state);
      } catch {
        // listener errors must not break the lifecycle
      }
    }
  }

  /**
   * Topological sort of plugins based on their declared dependencies.
   * Throws on missing dependencies or circular references.
   */
  private topologicalSort(): string[] {
    const visited = new Set<string>();
    const visiting = new Set<string>();
    const result: string[] = [];

    const visit = (name: string, chain: string[]): void => {
      if (visited.has(name)) return;
      if (visiting.has(name)) {
        throw new Error(
          `[PluginManager] Circular dependency detected: ${[...chain, name].join(" → ")}`,
        );
      }

      const entry = this.plugins.get(name);
      if (!entry) {
        throw new Error(
          `[PluginManager] Missing dependency: "${name}" ` +
            `(required by "${chain[chain.length - 1] ?? "root"}")`,
        );
      }

      visiting.add(name);

      for (const dep of entry.plugin.dependencies ?? []) {
        visit(dep, [...chain, name]);
      }

      visiting.delete(name);
      visited.add(name);
      result.push(name);
    };

    for (const name of this.plugins.keys()) {
      visit(name, []);
    }

    return result;
  }
}
