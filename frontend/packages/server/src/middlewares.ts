import type { CorsConfig, Middleware } from './types.js';
import { randomUUID } from 'node:crypto';

export function cors(config?: CorsConfig): Middleware {
  const origin = config?.origin ?? '*';
  const methods = config?.methods ?? ['GET', 'POST', 'PUT', 'DELETE', 'PATCH', 'OPTIONS', 'HEAD'];
  const allowedHeaders = config?.allowedHeaders ?? ['Content-Type', 'Authorization'];
  const credentials = config?.credentials ?? false;
  const maxAge = config?.maxAge ?? 86400;

  return async (req, res, next) => {
    const requestOrigin = req.headers['origin'] as string | undefined;

    let allowOrigin: string;
    if (origin === true) {
      allowOrigin = requestOrigin ?? '*';
    } else if (origin === false) {
      allowOrigin = '';
    } else if (Array.isArray(origin)) {
      allowOrigin = requestOrigin && origin.includes(requestOrigin) ? requestOrigin : '';
    } else {
      allowOrigin = origin;
    }

    if (allowOrigin) {
      res.header('Access-Control-Allow-Origin', allowOrigin);
    }
    res.header('Access-Control-Allow-Methods', methods.join(', '));
    res.header('Access-Control-Allow-Headers', allowedHeaders.join(', '));
    if (credentials) {
      res.header('Access-Control-Allow-Credentials', 'true');
    }
    res.header('Access-Control-Max-Age', String(maxAge));

    if (req.method === 'OPTIONS') {
      res.status(204).send('');
      return;
    }

    await next();
  };
}

export function bodyParser(): Middleware {
  return async (_req, _res, next) => {
    await next();
  };
}

export function requestId(): Middleware {
  return async (req, res, next) => {
    const id = (req.headers['x-request-id'] as string) ?? randomUUID();
    res.header('X-Request-ID', id);
    await next();
  };
}

export function responseTime(): Middleware {
  return async (_req, res, next) => {
    const start = performance.now();
    await next();
    const duration = (performance.now() - start).toFixed(2);
    res.header('X-Response-Time', `${duration}ms`);
  };
}

export function helmet(): Middleware {
  return async (_req, res, next) => {
    res.header('X-Frame-Options', 'SAMEORIGIN');
    res.header('X-Content-Type-Options', 'nosniff');
    res.header('X-XSS-Protection', '0');
    res.header('Referrer-Policy', 'strict-origin-when-cross-origin');
    res.header('X-Permitted-Cross-Domain-Policies', 'none');
    res.header('X-Download-Options', 'noopen');
    await next();
  };
}

export function compress(): Middleware {
  // TODO: implement actual gzip compression
  return async (_req, _res, next) => {
    await next();
  };
}

export function rateLimitMiddleware(config: { windowMs: number; max: number }): Middleware {
  const hits = new Map<string, { count: number; resetAt: number }>();

  return async (req, res, next) => {
    const key = req.ip;
    const now = Date.now();
    let entry = hits.get(key);

    if (!entry || now >= entry.resetAt) {
      entry = { count: 0, resetAt: now + config.windowMs };
      hits.set(key, entry);
    }

    entry.count++;

    res.header('X-RateLimit-Limit', String(config.max));
    res.header('X-RateLimit-Remaining', String(Math.max(0, config.max - entry.count)));
    res.header('X-RateLimit-Reset', String(Math.ceil(entry.resetAt / 1000)));

    if (entry.count > config.max) {
      res.status(429).json({ error: 'Too Many Requests' });
      return;
    }

    await next();
  };
}
