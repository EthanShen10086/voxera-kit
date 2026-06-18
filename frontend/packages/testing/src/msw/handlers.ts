import {
  createApiError,
  createApiListResponse,
  createApiSuccess,
  createSession,
  createUser,
  resetFixtures,
} from "@voxera-kit/fixture";
import { http, HttpResponse } from "msw";

const API_BASE = "http://localhost/api";

/** Auth-related REST handlers for tests. */
export const authHandlers = [
  http.post(`${API_BASE}/auth/login`, () => {
    resetFixtures();
    const user = createUser();
    const session = createSession({ userId: user.id });
    return HttpResponse.json(createApiSuccess({ user, session }));
  }),
  http.get(`${API_BASE}/auth/me`, () => {
    return HttpResponse.json(createApiSuccess(createUser()));
  }),
  http.post(`${API_BASE}/auth/logout`, () => {
    return HttpResponse.json(createApiSuccess({ ok: true }));
  }),
];

/** Paginated list REST handlers for tests. */
export const paginationHandlers = [
  http.get(`${API_BASE}/items`, ({ request }) => {
    const url = new URL(request.url);
    const page = Number(url.searchParams.get("page") ?? "1");
    const pageSize = Number(url.searchParams.get("pageSize") ?? "10");
    const items = [createUser(), createUser()];
    return HttpResponse.json(
      createApiListResponse(items, {
        page,
        pageSize,
        total: items.length,
      }),
    );
  }),
];

/** Error response handlers keyed by HTTP status. */
export const errorHandlers = [
  http.get(`${API_BASE}/errors/:code`, ({ params }) => {
    const code = String(params.code);
    const status = code === "unauthorized" ? 401 : code === "forbidden" ? 403 : 400;
    return HttpResponse.json(createApiError({ code: code.toUpperCase() }), { status });
  }),
];

/** Default MSW handlers bundled for kit and product tests. */
export const mswHandlers = [...authHandlers, ...paginationHandlers, ...errorHandlers];

export { API_BASE };
