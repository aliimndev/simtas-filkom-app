import type { Context } from "hono";

export interface ApiErrorBody {
  error: { code: string; message: string };
}

export function throwError(c: Context, code: string, message: string, status: number) {
  return c.json({ error: { code, message } } satisfies ApiErrorBody, status as any);
}

export function errorHandler(err: any, c: Context) {
  const code = err?.code ?? "INTERNAL";
  const message = err?.message ?? "Internal server error";
  const status: Record<string, number> = {
    UNAUTHORIZED: 401,
    FORBIDDEN: 403,
    NOT_FOUND: 404,
    CONFLICT: 409,
    LOCKED: 423,
    RATE_LIMIT: 429,
    VALIDATION: 400,
  };
  return c.json({ error: { code, message } } satisfies ApiErrorBody, (status[code] ?? 500) as any);
}
