import { createServer, type IncomingMessage, type ServerResponse as HttpServerResponse, type Server } from 'node:http';
import { URL } from 'node:url';
import type {
  HttpMethod,
  IHttpServer,
  Middleware,
  NextFunction,
  RouteHandler,
  ServerRequest,
  ServerResponse,
} from '../types.js';

interface Route {
  method: HttpMethod;
  pattern: string;
  segments: string[];
  handlers: RouteHandler[];
}

function parseQuery(searchParams: URLSearchParams): Record<string, string | string[]> {
  const query: Record<string, string | string[]> = {};
  for (const [key, value] of searchParams) {
    const existing = query[key];
    if (existing === undefined) {
      query[key] = value;
    } else if (Array.isArray(existing)) {
      existing.push(value);
    } else {
      query[key] = [existing, value];
    }
  }
  return query;
}

function matchRoute(routeSegments: string[], pathSegments: string[]): Record<string, string> | null {
  if (routeSegments.length !== pathSegments.length) return null;
  const params: Record<string, string> = {};
  for (let i = 0; i < routeSegments.length; i++) {
    const seg = routeSegments[i];
    if (seg.startsWith(':')) {
      params[seg.slice(1)] = decodeURIComponent(pathSegments[i]);
    } else if (seg !== pathSegments[i]) {
      return null;
    }
  }
  return params;
}

function splitPath(path: string): string[] {
  return path.split('/').filter(Boolean);
}

function readBody(req: IncomingMessage): Promise<string> {
  return new Promise((resolve, reject) => {
    const chunks: Buffer[] = [];
    req.on('data', (chunk: Buffer) => chunks.push(chunk));
    req.on('end', () => resolve(Buffer.concat(chunks).toString('utf-8')));
    req.on('error', reject);
  });
}

function wrapRequest(req: IncomingMessage, params: Record<string, string>, body: unknown): ServerRequest {
  const url = req.url ?? '/';
  const parsed = new URL(url, `http://${req.headers.host ?? 'localhost'}`);
  return {
    method: (req.method ?? 'GET').toUpperCase() as HttpMethod,
    url,
    path: parsed.pathname,
    query: parseQuery(parsed.searchParams),
    params,
    headers: req.headers as Record<string, string | string[] | undefined>,
    body,
    ip: req.socket.remoteAddress ?? '127.0.0.1',
    raw: req,
  };
}

function wrapResponse(res: HttpServerResponse): ServerResponse {
  let statusCode = 200;
  let headersSent = false;

  const wrapped: ServerResponse = {
    raw: res,
    status(code: number) {
      statusCode = code;
      return wrapped;
    },
    json(data: unknown) {
      if (headersSent) return;
      headersSent = true;
      res.writeHead(statusCode, { 'Content-Type': 'application/json' });
      res.end(JSON.stringify(data));
    },
    send(data: string | Buffer) {
      if (headersSent) return;
      headersSent = true;
      res.writeHead(statusCode);
      res.end(data);
    },
    header(name: string, value: string) {
      res.setHeader(name, value);
      return wrapped;
    },
    redirect(url: string, code = 302) {
      if (headersSent) return;
      headersSent = true;
      res.writeHead(code, { Location: url });
      res.end();
    },
    stream(readable: NodeJS.ReadableStream) {
      if (headersSent) return;
      headersSent = true;
      res.writeHead(statusCode);
      readable.pipe(res);
    },
  };
  return wrapped;
}

export class RawHttpServer implements IHttpServer {
  private server: Server | null = null;
  private middlewares: Middleware[] = [];
  private routes: Route[] = [];
  private prefix = '';

  private addRoute(method: HttpMethod, path: string, handlers: RouteHandler[]): this {
    const fullPath = this.prefix + path;
    this.routes.push({
      method,
      pattern: fullPath,
      segments: splitPath(fullPath),
      handlers,
    });
    return this;
  }

  use(...middlewares: Middleware[]): this {
    this.middlewares.push(...middlewares);
    return this;
  }

  get(path: string, ...handlers: RouteHandler[]): this {
    return this.addRoute('GET', path, handlers);
  }

  post(path: string, ...handlers: RouteHandler[]): this {
    return this.addRoute('POST', path, handlers);
  }

  put(path: string, ...handlers: RouteHandler[]): this {
    return this.addRoute('PUT', path, handlers);
  }

  delete(path: string, ...handlers: RouteHandler[]): this {
    return this.addRoute('DELETE', path, handlers);
  }

  patch(path: string, ...handlers: RouteHandler[]): this {
    return this.addRoute('PATCH', path, handlers);
  }

  group(prefix: string): RawHttpServer {
    const child = new RawHttpServer();
    child.prefix = this.prefix + prefix;
    child.routes = this.routes;
    child.middlewares = this.middlewares;
    return child;
  }

  listen(port: number, hostname?: string): Promise<void> {
    return new Promise((resolve, reject) => {
      this.server = createServer(async (raw, rawRes) => {
        const bodyText = await readBody(raw);
        let body: unknown = bodyText;
        const contentType = raw.headers['content-type'] ?? '';
        if (contentType.includes('application/json') && bodyText) {
          try {
            body = JSON.parse(bodyText);
          } catch {
            // leave as string
          }
        }

        const url = raw.url ?? '/';
        const parsed = new URL(url, `http://${raw.headers.host ?? 'localhost'}`);
        const method = (raw.method ?? 'GET').toUpperCase() as HttpMethod;
        const pathSegments = splitPath(parsed.pathname);

        let matchedRoute: Route | undefined;
        let params: Record<string, string> = {};
        for (const route of this.routes) {
          if (route.method !== method) continue;
          const result = matchRoute(route.segments, pathSegments);
          if (result !== null) {
            matchedRoute = route;
            params = result;
            break;
          }
        }

        const req = wrapRequest(raw, params, body);
        const res = wrapResponse(rawRes);

        const allMiddlewares = [...this.middlewares];
        let idx = 0;

        const runNext: NextFunction = async () => {
          if (idx < allMiddlewares.length) {
            const mw = allMiddlewares[idx++];
            await mw(req, res, runNext);
          } else if (matchedRoute) {
            for (const handler of matchedRoute.handlers) {
              await handler(req, res);
            }
          } else {
            res.status(404).json({ error: 'Not Found' });
          }
        };

        try {
          await runNext();
        } catch (err) {
          if (!rawRes.headersSent) {
            res.status(500).json({ error: 'Internal Server Error' });
          }
        }
      });

      this.server.on('error', reject);
      this.server.listen(port, hostname ?? '0.0.0.0', () => resolve());
    });
  }

  close(): Promise<void> {
    return new Promise((resolve, reject) => {
      if (!this.server) {
        resolve();
        return;
      }
      this.server.close((err) => (err ? reject(err) : resolve()));
    });
  }

  address(): { host: string; port: number } | null {
    const addr = this.server?.address();
    if (!addr || typeof addr === 'string') return null;
    return { host: addr.address, port: addr.port };
  }
}
