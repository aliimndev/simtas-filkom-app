import type { MiddlewareHandler } from "hono";

// Propagate or mint a request id and echo it back; the error envelope reads it too.
export const requestId: MiddlewareHandler = async (c, next) => {
  const id = c.req.header("x-request-id") ?? crypto.randomUUID();
  c.set("requestId", id);
  c.header("X-Request-Id", id);
  await next();
};

// ponytail: one-line access log; swap for pino/otelbats when we need structured sinks.
export const logger: MiddlewareHandler = async (c, next) => {
  const start = Date.now();
  await next();
  console.log(`${c.req.method} ${c.req.path} ${c.res.status} ${c.get("requestId")} ${Date.now() - start}ms`);
};
