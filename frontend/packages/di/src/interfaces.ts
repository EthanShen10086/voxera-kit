/**
 * A token used to identify a service in the container.
 * Can be a Symbol, string, or a class constructor.
 */
export type ServiceIdentifier<T = unknown> =
  | symbol
  | string
  | (new (...args: unknown[]) => T);

/**
 * A provider that knows how to resolve an instance of type T.
 */
export interface IProvider<T> {
  /** Resolve and return an instance of T. */
  resolve(): T;
}

/**
 * Controls how many instances the container creates for a given registration.
 */
export enum Lifecycle {
  /** A single shared instance is created and reused for every resolve call. */
  Singleton = "singleton",
  /** A new instance is created on every resolve call. */
  Transient = "transient",
}

/**
 * Inversion-of-Control container interface.
 * Manages service registration, resolution, and hierarchical scoping.
 */
export interface IContainer {
  /**
   * Register a factory for the given token.
   * @param token - Unique identifier for the service.
   * @param factory - Factory function that produces an instance.
   * @param lifecycle - Controls instance caching strategy. Defaults to `Singleton`.
   */
  register<T>(
    token: ServiceIdentifier<T>,
    factory: () => T,
    lifecycle?: Lifecycle,
  ): void;

  /**
   * Resolve an instance for the given token.
   * @param token - The service identifier to look up.
   * @throws {Error} If the token has not been registered.
   */
  resolve<T>(token: ServiceIdentifier<T>): T;

  /**
   * Check whether a token has been registered in this container or any parent.
   */
  has(token: ServiceIdentifier<unknown>): boolean;

  /** Clear all registrations and cached instances. */
  reset(): void;

  /**
   * Create a child container that delegates to this container
   * when a token is not found in its own registrations.
   */
  createChild(): IContainer;
}
