import { describe, expect, it } from "vitest";

import { Container } from "./container.js";
import { Lifecycle } from "./interfaces.js";

describe("Container", () => {
  it("resolves singleton services once", () => {
    const container = new Container();
    let calls = 0;
    const token = Symbol("svc");
    container.register(
      token,
      () => {
        calls += 1;
        return { id: calls };
      },
      Lifecycle.Singleton,
    );

    const a = container.resolve(token);
    const b = container.resolve(token);
    expect(a).toBe(b);
    expect(calls).toBe(1);
  });
});
