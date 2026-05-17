export type HttpMethod = 'GET' | 'POST' | 'PUT' | 'DELETE' | 'PATCH' | 'OPTIONS' | 'HEAD';

export interface ServerRequest {
  method: HttpMethod;
  url: string;
  path: string;
  query: Record<string, string | string[]>;
  params: Record<string, string>;
  headers: Record<string, string | string[] | undefined>;
  body: unknown;
  ip: string;
  raw: unknown;
}

export interface ServerResponse {
  status(code: number): ServerResponse;
  json(data: unknown): void;
  send(data: string | Buffer): void;
  header(name: string, value: string): ServerResponse;
  redirect(url: string, code?: number): void;
  stream(readable: NodeJS.ReadableStream): void;
  raw: unknown;
}

export type NextFunction = () => Promise<void> | void;

export type Middleware = (req: ServerRequest, res: ServerResponse, next: NextFunction) => Promise<void> | void;

export type RouteHandler = (req: ServerRequest, res: ServerResponse) => Promise<void> | void;

export interface Router {
  use(...middlewares: Middleware[]): Router;
  get(path: string, ...handlers: RouteHandler[]): Router;
  post(path: string, ...handlers: RouteHandler[]): Router;
  put(path: string, ...handlers: RouteHandler[]): Router;
  delete(path: string, ...handlers: RouteHandler[]): Router;
  patch(path: string, ...handlers: RouteHandler[]): Router;
  group(prefix: string): Router;
}

export interface IHttpServer extends Router {
  listen(port: number, hostname?: string): Promise<void>;
  close(): Promise<void>;
  address(): { host: string; port: number } | null;
}

export interface ServerConfig {
  port?: number;
  hostname?: string;
  trustProxy?: boolean;
  maxBodySize?: string;
  cors?: CorsConfig;
  compression?: boolean;
}

export interface CorsConfig {
  origin: string | string[] | boolean;
  methods?: HttpMethod[];
  allowedHeaders?: string[];
  credentials?: boolean;
  maxAge?: number;
}

export interface IServerFactory {
  create(config?: ServerConfig): IHttpServer;
}
