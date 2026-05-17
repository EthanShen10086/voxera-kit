import type { ServiceIdentifier } from "./interfaces.js";

/**
 * Metadata key → set of classes marked as injectable.
 */
const injectableRegistry = new Set<Function>();

/**
 * Metadata key → map from class to its injection descriptors.
 * Key: target constructor. Value: map of parameter-index → token.
 */
const injectionMetadata = new Map<
  Function,
  Map<number | string, ServiceIdentifier<unknown>>
>();

/**
 * Mark a class as injectable so the container can construct it.
 *
 * @example
 * ```ts
 * @Injectable()
 * class MyService { }
 * ```
 */
export function Injectable(): ClassDecorator {
  return (target: Function) => {
    injectableRegistry.add(target);
  };
}

/**
 * Mark a constructor parameter or class property as an injection point.
 *
 * @param token - The service identifier to inject.
 *
 * @example
 * ```ts
 * @Injectable()
 * class Foo {
 *   constructor(@Inject(BAR_TOKEN) private bar: Bar) {}
 * }
 * ```
 */
export function Inject(
  token: ServiceIdentifier<unknown>,
): ParameterDecorator & PropertyDecorator {
  return (
    target: Object,
    propertyKey: string | symbol | undefined,
    parameterIndex?: number,
  ) => {
    const ctor =
      typeof target === "function" ? target : target.constructor;

    let meta = injectionMetadata.get(ctor);
    if (!meta) {
      meta = new Map();
      injectionMetadata.set(ctor, meta);
    }

    if (parameterIndex !== undefined) {
      meta.set(parameterIndex, token);
    } else if (propertyKey !== undefined) {
      meta.set(String(propertyKey), token);
    }
  };
}

/** Check whether a class has been decorated with `@Injectable()`. */
export function isInjectable(target: Function): boolean {
  return injectableRegistry.has(target);
}

/**
 * Retrieve injection metadata for a class.
 * Returns a map of parameter-index (or property name) → token.
 */
export function getInjectionMetadata(
  target: Function,
): ReadonlyMap<number | string, ServiceIdentifier<unknown>> {
  return injectionMetadata.get(target) ?? new Map();
}
