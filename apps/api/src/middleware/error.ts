import type { MiddlewareHandler } from "hono";

export interface ApiErrorBody {
  error: { code: string; message: string };
}

// Full onError renderer + throwError() helper are implemented in Task 7.
export const errorStub: MiddlewareHandler = async (_c, next) => {
  await next();
};
