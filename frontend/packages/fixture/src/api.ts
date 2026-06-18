import type { ApiErrorResponse, ApiListResponse } from "./types.js";

let apiSeq = 0;

/** Resets API fixture counters for deterministic tests. */
export function resetApiFixtures(): void {
  apiSeq = 0;
}

/** Creates a paginated list API response. */
export function createApiListResponse<T>(
  items: T[],
  overrides: Partial<ApiListResponse<T>> = {},
): ApiListResponse<T> {
  apiSeq += 1;
  return {
    data: items,
    total: items.length,
    page: 1,
    pageSize: Math.max(items.length, 10),
    ...overrides,
  };
}

/** Creates a generic success payload wrapper used by REST handlers in tests. */
export function createApiSuccess<T>(data: T): { data: T; requestId: string } {
  apiSeq += 1;
  return {
    data,
    requestId: `req-${apiSeq}`,
  };
}

/** Creates a standard API error response. */
export function createApiError(overrides: Partial<ApiErrorResponse> = {}): ApiErrorResponse {
  apiSeq += 1;
  return {
    code: "INTERNAL_ERROR",
    message: `Request failed (${apiSeq})`,
    ...overrides,
  };
}
