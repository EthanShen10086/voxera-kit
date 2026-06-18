import { setupServer, type SetupServer } from "msw/node";
import type { RequestHandler } from "msw";

import { mswHandlers } from "./handlers.js";

/** Creates an MSW server with kit default handlers plus optional extras. */
export function createMockServer(extraHandlers: RequestHandler[] = []): SetupServer {
  return setupServer(...mswHandlers, ...extraHandlers);
}

export { mswHandlers, authHandlers, errorHandlers, paginationHandlers, API_BASE } from "./handlers.js";
