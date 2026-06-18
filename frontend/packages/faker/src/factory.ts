import type { CreateFakerOptions, FakerPort } from "./port.js";
import { createFakerJSAdapter } from "./fakerjs.js";
import { createLightweightFaker } from "./lightweight.js";

/** Creates a pluggable faker; defaults to fakerjs, falls back to lightweight. */
export function createFaker(options: CreateFakerOptions = {}): FakerPort {
  const provider = options.provider ?? "fakerjs";
  if (provider === "lightweight") {
    return createLightweightFaker(options.seed ?? 42);
  }
  try {
    return createFakerJSAdapter(options.seed);
  } catch {
    return createLightweightFaker(options.seed ?? 42);
  }
}
