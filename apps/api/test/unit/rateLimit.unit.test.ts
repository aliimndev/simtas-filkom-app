import { describe, expect, test } from "bun:test";
import { rateLimit } from "../../src/middleware/rateLimit";

function context(ip: string, headers: Record<string, string> = {}) {
  const responseHeaders = new Headers();
  return {
    req: {
      header(name: string) {
        return headers[name.toLowerCase()] ?? (name === "x-forwarded-for" ? ip : undefined);
      },
    },
    header(name: string, value: string) {
      responseHeaders.set(name, value);
    },
    json(body: unknown, status: number) {
      return new Response(JSON.stringify(body), { status, headers: responseHeaders });
    },
  } as any;
}

describe("rateLimit middleware", () => {
  test("allows requests up to the configured maximum and rejects the next one", async () => {
    const middleware = rateLimit({ windowMs: 60_000, max: 2 });
    let nextCalls = 0;
    const next = async () => {
      nextCalls += 1;
    };

    expect(await middleware(context("10.0.0.1"), next)).toBeUndefined();
    expect(await middleware(context("10.0.0.1"), next)).toBeUndefined();

    const limited = await middleware(context("10.0.0.1"), next);
    expect(limited.status).toBe(429);
    expect(limited.headers.get("Retry-After")).toBeTruthy();
    expect(await limited.json()).toEqual({
      error: { code: "RATE_LIMIT", message: "Too many requests" },
    });
    expect(nextCalls).toBe(2);
  });

  test("uses Cloudflare's connecting IP before the forwarded IP", async () => {
    const middleware = rateLimit({ windowMs: 60_000, max: 1 });
    const next = async () => undefined;

    expect(await middleware(context("10.0.0.1", { "cf-connecting-ip": "10.0.0.2" }), next)).toBeUndefined();
    expect(await middleware(context("10.0.0.1", { "cf-connecting-ip": "10.0.0.2" }), next)).toBeDefined();
    expect(await middleware(context("10.0.0.1"), next)).toBeUndefined();
  });
});
