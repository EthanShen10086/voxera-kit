import { describe, expect, it } from "vitest";

import {
  createApiError,
  createApiListResponse,
  createApiSuccess,
  resetApiFixtures,
} from "./api.js";
import { createUser } from "./user.js";

describe("api fixtures", () => {
  it("builds list and success payloads", () => {
    resetApiFixtures();
    const users = [createUser()];
    expect(createApiListResponse(users)).toMatchObject({
      data: users,
      total: 1,
      page: 1,
    });
    expect(createApiSuccess({ ok: true })).toEqual({
      data: { ok: true },
      requestId: "req-2",
    });
  });

  it("builds error payloads", () => {
    resetApiFixtures();
    expect(createApiError({ code: "NOT_FOUND" })).toMatchObject({
      code: "NOT_FOUND",
    });
  });
});
