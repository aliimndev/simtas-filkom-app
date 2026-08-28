import { describe, expect, it } from "bun:test";
import { createApp } from "../src/app";
import { loadConfig } from "../src/config";

const app = createApp(loadConfig({ NODE_ENV: "test" }));

describe("health", () => {
  it("returns 200 status ok", async () => {
    const res = await app.request("/api/v1/health");
    expect(res.status).toBe(200);
    expect(await res.json()).toEqual({ status: "ok", db: "healthy" });
  });
  it("404 on unknown route", async () => {
    expect((await app.request("/api/v1/nope")).status).toBe(404);
  });
});
