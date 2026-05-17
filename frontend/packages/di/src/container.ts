import type { IContainer, ServiceIdentifier } from "./interfaces.js";
import { Lifecycle } from "./interfaces.js";

interface Registration {
  factory: () => unknown;
  lifecycle: Lifecycle;
  instance?: unknown;
}

/**
 * A lightweight IoC container with singleton/transient lifecycle support
 * and hierarchical child containers.
 */
export class Container implements IContainer {
  private readonly registrations = new Map<
    ServiceIdentifier<unknown>,
    Registration
  >();

  constructor(private readonly parent?: Container) {}

  register<T>(
    token: ServiceIdentifier<T>,
    factory: () => T,
    lifecycle: Lifecycle = Lifecycle.Singleton,
  ): void {
    this.registrations.set(token, { factory, lifecycle });
  }

  resolve<T>(token: ServiceIdentifier<T>): T {
    const registration = this.registrations.get(token);

    if (registration) {
      if (
        registration.lifecycle === Lifecycle.Singleton &&
        registration.instance !== undefined
      ) {
        return registration.instance as T;
      }

      const instance = registration.factory() as T;

      if (registration.lifecycle === Lifecycle.Singleton) {
        registration.instance = instance;
      }

      return instance;
    }

    if (this.parent) {
      return this.parent.resolve<T>(token);
    }

    const tokenName =
      typeof token === "symbol"
        ? token.toString()
        : typeof token === "string"
          ? `"${token}"`
          : token.name;

    throw new Error(
      `[Container] No registration found for token ${tokenName}. ` +
        `Did you forget to call container.register()?`,
    );
  }

  has(token: ServiceIdentifier<unknown>): boolean {
    if (this.registrations.has(token)) {
      return true;
    }
    return this.parent?.has(token) ?? false;
  }

  reset(): void {
    this.registrations.clear();
  }

  createChild(): IContainer {
    return new Container(this);
  }
}
