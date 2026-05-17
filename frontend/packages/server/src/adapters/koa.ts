import type { IHttpServer, Middleware, RouteHandler } from '../types.js';

const NOT_IMPLEMENTED = 'KoaServer not yet implemented';

export class KoaServer implements IHttpServer {
  use(..._middlewares: Middleware[]): this {
    throw new Error(NOT_IMPLEMENTED);
  }

  get(_path: string, ..._handlers: RouteHandler[]): this {
    throw new Error(NOT_IMPLEMENTED);
  }

  post(_path: string, ..._handlers: RouteHandler[]): this {
    throw new Error(NOT_IMPLEMENTED);
  }

  put(_path: string, ..._handlers: RouteHandler[]): this {
    throw new Error(NOT_IMPLEMENTED);
  }

  delete(_path: string, ..._handlers: RouteHandler[]): this {
    throw new Error(NOT_IMPLEMENTED);
  }

  patch(_path: string, ..._handlers: RouteHandler[]): this {
    throw new Error(NOT_IMPLEMENTED);
  }

  group(_prefix: string): this {
    throw new Error(NOT_IMPLEMENTED);
  }

  listen(_port: number, _hostname?: string): Promise<void> {
    throw new Error(NOT_IMPLEMENTED);
  }

  close(): Promise<void> {
    throw new Error(NOT_IMPLEMENTED);
  }

  address(): { host: string; port: number } | null {
    throw new Error(NOT_IMPLEMENTED);
  }
}
