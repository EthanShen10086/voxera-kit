import type { IRemoteModule } from "./contract.js";

/**
 * Validate that a value satisfies the minimal `IRemoteModule` contract.
 */
function assertRemoteModule(value: unknown): asserts value is IRemoteModule {
  if (typeof value !== "object" || value === null) {
    throw new Error("Remote factory must return an object");
  }
  const obj = value as Record<string, unknown>;

  if (typeof obj["mount"] !== "function") {
    throw new Error('Remote module is missing required "mount" method');
  }
  if (typeof obj["unmount"] !== "function") {
    throw new Error('Remote module is missing required "unmount" method');
  }
}

/**
 * Helper for remote applications to create a properly typed `IRemoteModule`.
 * The factory is invoked immediately and its output validated.
 */
export function defineRemote(factory: () => IRemoteModule): IRemoteModule {
  const module = factory();
  assertRemoteModule(module);
  return module;
}

/**
 * Set up the remote entry point so that the host shell can mount the module.
 *
 * In a Module Federation remote, this would typically be called from the
 * `exposes` entry to wire up lifecycle hooks.
 */
export function createRemoteBootstrap(module: IRemoteModule): {
  mount: IRemoteModule["mount"];
  unmount: IRemoteModule["unmount"];
} {
  assertRemoteModule(module);

  return {
    mount: (container, props) => module.mount(container, props),
    unmount: () => module.unmount(),
  };
}
