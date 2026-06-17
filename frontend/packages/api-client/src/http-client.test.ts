import { describe, expect, it } from "vitest";

import { HttpClientError } from "./http-client.js";

describe("HttpClientError", () => {
  it("captures status and message", () => {
    const err = new HttpClientError("boom", 503, { url: "/x", method: "GET" });
    expect(err.message).toBe("boom");
    expect(err.status).toBe(503);
  });
});
