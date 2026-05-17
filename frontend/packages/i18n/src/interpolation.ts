import type { InterpolationParams } from "./types.js";

/**
 * Resolve a dot-notation path against a flat params object.
 * Supports `{user.name}` by splitting on "." and traversing nested objects,
 * but since `InterpolationParams` is flat, we also accept `"user.name"` as a
 * literal key for convenience.
 */
function resolveParam(
  path: string,
  params: InterpolationParams,
): string | number | undefined {
  if (path in params) {
    return params[path];
  }

  const segments = path.split(".");
  let current: unknown = params;

  for (const segment of segments) {
    if (current === null || current === undefined || typeof current !== "object") {
      return undefined;
    }
    current = (current as Record<string, unknown>)[segment];
  }

  if (typeof current === "string" || typeof current === "number") {
    return current;
  }

  return undefined;
}

/**
 * Replace `{param}` placeholders in `template` with values from `params`.
 * Missing params are left as-is (e.g. `"{missing}"` stays `"{missing}"`).
 */
export function interpolate(
  template: string,
  params: InterpolationParams,
): string {
  return template.replace(/\{([^}]+)\}/g, (_match, key: string) => {
    const value = resolveParam(key.trim(), params);
    return value !== undefined ? String(value) : `{${key}}`;
  });
}
