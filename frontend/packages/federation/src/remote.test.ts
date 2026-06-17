import { describe, expect, it } from "vitest";

import { createRemoteBootstrap, defineRemote } from "./remote.js";

describe("federation remote", () => {
  it("defineRemote validates mount/unmount contract", () => {
    const mod = defineRemote(() => ({
      mount: () => undefined,
      unmount: () => undefined,
    }));
    expect(typeof mod.mount).toBe("function");
  });

  it("createRemoteBootstrap exposes lifecycle hooks", () => {
    const mod = defineRemote(() => ({
      mount: () => undefined,
      unmount: () => undefined,
    }));
    const bootstrap = createRemoteBootstrap(mod);
    expect(typeof bootstrap.mount).toBe("function");
    expect(typeof bootstrap.unmount).toBe("function");
  });
});
