import { beforeAll, afterAll, describe, expect, it } from "bun:test";
import { createApp } from "../src/app";
import { loadConfig } from "../src/config";
import { seedTestUser, clearTestUser, TEST_USER } from "./helpers";

// Use temp DB URL if not set
const TEST_DB_URL = process.env.DATABASE_URL ?? "postgres://postgres@localhost:5433/simtas";
const app = createApp(loadConfig({ ...process.env, NODE_ENV: "test", DATABASE_URL: TEST_DB_URL } as any));

beforeAll(async () => {
  await seedTestUser();
});
afterAll(async () => {
  await clearTestUser();
});

describe("auth parity", () => {
  it("login → me → refresh → logout", async () => {
    const login = await app.request("/api/v1/auth/login", {
      method: "POST",
      headers: { "content-type": "application/json" },
      body: JSON.stringify({ email: TEST_USER.email, password: TEST_USER.password }),
    });
    expect(login.status).toBe(200);
    const { accessToken, refreshToken } = (await login.json()) as any;
    expect(accessToken).toBeDefined();
    expect(refreshToken).toBeDefined();

    const me = await app.request("/api/v1/auth/me", { headers: { authorization: `Bearer ${accessToken}` } });
    expect(me.status).toBe(200);

    const refresh = await app.request("/api/v1/auth/refresh", {
      method: "POST",
      headers: { "content-type": "application/json" },
      body: JSON.stringify({ refreshToken }),
    });
    expect(refresh.status).toBe(200);
    const refreshBody = (await refresh.json()) as any;
    expect(refreshBody.accessToken).toBeDefined();

    const logout = await app.request("/api/v1/auth/logout", {
      method: "POST",
      headers: { "content-type": "application/json" },
      body: JSON.stringify({ refreshToken: refreshBody.refreshToken ?? refreshToken }),
    });
    expect(logout.status).toBe(204);
  });

  it("returns 401 on wrong password", async () => {
    const res = await app.request("/api/v1/auth/login", {
      method: "POST",
      headers: { "content-type": "application/json" },
      body: JSON.stringify({ email: TEST_USER.email, password: "WrongPass1!" }),
    });
    expect(res.status).toBe(401);
  });
});
