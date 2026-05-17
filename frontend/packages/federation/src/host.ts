import type { IRemoteModule, RemoteModuleConfig } from "./contract.js";

declare const __webpack_init_sharing__: (scope: string) => Promise<void>;
declare const __webpack_share_scopes__: Record<string, unknown>;

const LOAD_TIMEOUT_MS = 10_000;

const moduleCache = new Map<string, IRemoteModule>();
const containerCache = new Map<string, Record<string, unknown>>();

function loadScript(url: string, scope: string): Promise<void> {
  return new Promise((resolve, reject) => {
    const existing = document.querySelector(`script[data-federation="${scope}"]`);
    if (existing) {
      resolve();
      return;
    }

    const script = document.createElement("script");
    script.src = url;
    script.type = "text/javascript";
    script.async = true;
    script.dataset["federation"] = scope;

    const timer = setTimeout(() => {
      script.remove();
      reject(new Error(`Timed out loading remote "${scope}" from ${url}`));
    }, LOAD_TIMEOUT_MS);

    script.onload = () => {
      clearTimeout(timer);
      resolve();
    };

    script.onerror = () => {
      clearTimeout(timer);
      script.remove();
      reject(new Error(`Failed to load remote "${scope}" from ${url}`));
    };

    document.head.appendChild(script);
  });
}

async function initContainer(
  scope: string,
): Promise<Record<string, unknown>> {
  const cached = containerCache.get(scope);
  if (cached) return cached;

  await __webpack_init_sharing__("default");

  const container = (window as unknown as Record<string, unknown>)[scope] as
    | { init: (scopes: unknown) => Promise<void>; get: (module: string) => Promise<() => unknown> }
    | undefined;

  if (!container) {
    throw new Error(
      `Remote container "${scope}" not found on window after script load`,
    );
  }

  await container.init(__webpack_share_scopes__["default"]);

  const ref = container as unknown as Record<string, unknown>;
  containerCache.set(scope, ref);
  return ref;
}

/**
 * Load a remote module from a Module Federation host.
 *
 * 1. Injects the remote entry script if not already loaded.
 * 2. Initialises the webpack share scopes.
 * 3. Retrieves and caches the requested module factory.
 */
export async function loadRemoteModule(
  config: RemoteModuleConfig,
): Promise<IRemoteModule> {
  const cacheKey = `${config.scope}/${config.module}`;
  const cached = moduleCache.get(cacheKey);
  if (cached) return cached;

  await loadScript(config.url, config.scope);

  const container = await initContainer(config.scope);

  const getModule = container["get"] as
    | ((id: string) => Promise<() => Record<string, unknown>>)
    | undefined;

  if (typeof getModule !== "function") {
    throw new Error(`Container "${config.scope}" does not expose a "get" method`);
  }

  const factory = await getModule(config.module);
  const moduleExports = factory();

  const remote = (moduleExports["default"] ?? moduleExports) as IRemoteModule;

  if (typeof remote.mount !== "function" || typeof remote.unmount !== "function") {
    throw new Error(
      `Remote module "${cacheKey}" does not implement the IRemoteModule contract (missing mount/unmount)`,
    );
  }

  moduleCache.set(cacheKey, remote);
  return remote;
}

/**
 * Remove a previously loaded remote module from cache and clean up the injected script tag.
 */
export function unloadRemoteModule(scope: string): void {
  for (const [key] of moduleCache) {
    if (key.startsWith(`${scope}/`)) {
      moduleCache.delete(key);
    }
  }
  containerCache.delete(scope);

  const script = document.querySelector(`script[data-federation="${scope}"]`);
  script?.remove();

  delete (window as unknown as Record<string, unknown>)[scope];
}
