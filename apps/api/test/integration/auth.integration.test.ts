import { beforeAll, afterAll, describe, expect, it } from "bun:test";
import { createApp } from "../../src/app";
import { loadConfig } from "../../src/config";
import { signAccessToken } from "../../src/modules/auth";
import { getDb } from "../../src/db";
import { eq } from "drizzle-orm";
import { schema } from "@sims/db";
import { seedTestUser, clearTestUser, resetTestUserLock, setTestUserActive, TEST_USER } from "../support/helpers";

// Use temp DB URL if not set
const TEST_DB_URL = process.env.DATABASE_URL ?? "postgres://postgres@localhost:5433/simtas";
const app = createApp(loadConfig({ ...process.env, NODE_ENV: "test", DATABASE_URL: TEST_DB_URL } as any));

beforeAll(async () => {
  await seedTestUser();
});
afterAll(async () => {
  await clearTestUser();
});

// Each request gets a unique IP so the per-IP rate limiters don't accumulate across
// tests. The dedicated rate-limit tests below pass an explicit shared xff to prove the limit.
let seq = 0;
const req = (path: string, init: Record<string, any> = {}) =>
  app.request(path, {
    ...init,
    headers: { "x-forwarded-for": `parity-${seq++}`, ...(init.headers ?? {}) },
  });

const login = (body: unknown, headers: Record<string, string> = {}) =>
  req("/api/v1/auth/login", {
    method: "POST",
    headers: { "content-type": "application/json", ...headers },
    body: JSON.stringify(body),
  });

describe("auth parity", () => {
  it("login → me → refresh → logout", async () => {
    const res = await login({ email: TEST_USER.email, password: TEST_USER.password });
    expect(res.status).toBe(200);
    const { accessToken, refreshToken } = (await res.json()) as any;
    expect(accessToken).toBeDefined();
    expect(refreshToken).toBeDefined();

    const me = await req("/api/v1/auth/me", { headers: { authorization: `Bearer ${accessToken}` } });
    expect(me.status).toBe(200);

    const refresh = await req("/api/v1/auth/refresh", {
      method: "POST",
      headers: { "content-type": "application/json" },
      body: JSON.stringify({ refreshToken }),
    });
    expect(refresh.status).toBe(200);
    const refreshBody = (await refresh.json()) as any;
    expect(refreshBody.accessToken).toBeDefined();

    const logout = await req("/api/v1/auth/logout", {
      method: "POST",
      headers: { "content-type": "application/json" },
      body: JSON.stringify({ refreshToken: refreshBody.refreshToken ?? refreshToken }),
    });
    expect(logout.status).toBe(204);
  });

  it("returns 401 on wrong password", async () => {
    const res = await login({ email: TEST_USER.email, password: "WrongPass1!" });
    expect(res.status).toBe(401);
  });

  it("returns 401 for an unknown email (timing-equalized path runs)", async () => {
    const res = await login({ email: "definitely-not-here@filkom.ac.id", password: "WrongPass1!" });
    expect(res.status).toBe(401);
  });

  it("rejects an inactive account with 403", async () => {
    await setTestUserActive(false);
    const res = await login({ email: TEST_USER.email, password: TEST_USER.password });
    expect(res.status).toBe(403);
    await setTestUserActive(true);
  });

  it("rejects /me without or with a bad token (401)", async () => {
    expect((await req("/api/v1/auth/me")).status).toBe(401);
    expect((await req("/api/v1/auth/me", { headers: { authorization: "Bearer not-a-jwt" } })).status).toBe(401);
  });

  it("locks after 5 failures and stays locked (423)", async () => {
    await resetTestUserLock();
    for (let i = 1; i <= 5; i++) {
      const res = await login({ email: TEST_USER.email, password: "WrongPass1!" });
      expect(res.status).toBe(i < 5 ? 401 : 423);
    }
    const sixth = await login({ email: TEST_USER.email, password: "WrongPass1!" });
    expect(sixth.status).toBe(423);
    await resetTestUserLock();
  });

  it("refreshes once, then reuse of the old token revokes the family (TOKEN_REUSE)", async () => {
    const res = await login({ email: TEST_USER.email, password: TEST_USER.password });
    expect(res.status).toBe(200);
    const { refreshToken } = (await res.json()) as any;

    const refresh = await req("/api/v1/auth/refresh", {
      method: "POST",
      headers: { "content-type": "application/json" },
      body: JSON.stringify({ refreshToken }),
    });
    expect(refresh.status).toBe(200);
    const { refreshToken: nextRefresh } = (await refresh.json()) as any;

    // Replay the spent token -> family must be revoked.
    const replay = await req("/api/v1/auth/refresh", {
      method: "POST",
      headers: { "content-type": "application/json" },
      body: JSON.stringify({ refreshToken }),
    });
    expect(replay.status).toBe(401);
    expect((await replay.json() as any).error.code).toBe("TOKEN_REUSE");

    // The rotated token is now also dead (family revoked).
    const after = await req("/api/v1/auth/refresh", {
      method: "POST",
      headers: { "content-type": "application/json" },
      body: JSON.stringify({ refreshToken: nextRefresh }),
    });
    expect(after.status).toBe(401);
  });

  it("RBAC: admin reaches /admin/ping, other roles 403, no token 401", async () => {
    const adminTok = await signAccessToken("admin-1", "ADMIN_FAKULTAS", 0);
    const mhsTok = await signAccessToken("mhs-1", "MAHASISWA", 0);

    const ok = await req("/api/v1/auth/admin/ping", { headers: { authorization: `Bearer ${adminTok}` } });
    expect(ok.status).toBe(200);

    const forbidden = await req("/api/v1/auth/admin/ping", { headers: { authorization: `Bearer ${mhsTok}` } });
    expect(forbidden.status).toBe(403);

    const noTok = await req("/api/v1/auth/admin/ping");
    expect(noTok.status).toBe(401);
  });

  it(
    "login rate limit: 429 after 10 requests/min (unknown email avoids lockout)",
    async () => {
      for (let i = 1; i <= 11; i++) {
        const res = await login(
          { email: "nobody@filkom.ac.id", password: "WrongPass1!" },
          { "x-forwarded-for": "login-rl" },
        );
        if (i === 11) expect(res.status).toBe(429);
        else expect(res.status).toBe(401);
      }
    },
    { timeout: 30000 },
  );

  it("global rate limit: 429 after 100 requests/min", async () => {
    for (let i = 1; i <= 101; i++) {
      const res = await req("/api/v1/health", { headers: { "x-forwarded-for": "global-rl" } });
      if (i === 101) expect(res.status).toBe(429);
      else expect(res.status).toBe(200);
    }
  });
});

describe("health parity", () => {
  it("returns 503 when the database is unreachable", async () => {
    const badApp = createApp(
      loadConfig({ NODE_ENV: "test", DATABASE_URL: "postgres://postgres@localhost:1/simtas" } as any),
    );
    const res = await badApp.request("/api/v1/health");
    expect(res.status).toBe(503);
  });
});

describe("access-token revocation (ticket 4)", () => {
  it("blacklists the access token on logout so /me returns 401 afterwards", async () => {
    const res = await login({ email: TEST_USER.email, password: TEST_USER.password });
    expect(res.status).toBe(200);
    const { accessToken } = (await res.json()) as any;

    const before = await req("/api/v1/auth/me", { headers: { authorization: `Bearer ${accessToken}` } });
    expect(before.status).toBe(200);

    const logout = await req("/api/v1/auth/logout", {
      method: "POST",
      headers: { "content-type": "application/json", authorization: `Bearer ${accessToken}` },
      body: JSON.stringify({}),
    });
    expect(logout.status).toBe(204);

    const after = await req("/api/v1/auth/me", { headers: { authorization: `Bearer ${accessToken}` } });
    expect(after.status).toBe(401);
    expect((await after.json() as any).error.code).toBe("UNAUTHORIZED");
  });

  it("rejects a token whose token_version no longer matches the user (session bump)", async () => {
    const cfg = loadConfig({ ...process.env, NODE_ENV: "test", DATABASE_URL: TEST_DB_URL } as any);
    const db = getDb(cfg.databaseUrl);

    // seedTestUser writes token_version 0; sign a token claiming version 99.
    // The real row stays at 0, so Authenticate must reject the mismatch.
    const staleTok = await signAccessToken("00000000-0000-0000-0000-000000000001", "ADMIN_FAKULTAS", 99);
    const res = await req("/api/v1/auth/me", { headers: { authorization: `Bearer ${staleTok}` } });
    expect(res.status).toBe(401);
    expect((await res.json() as any).error.code).toBe("UNAUTHORIZED");

    // Restore the user's token_version for downstream tests.
    await db
      .update(schema.users)
      .set({ tokenVersion: 0 })
      .where(eq(schema.users.id, "00000000-0000-0000-0000-000000000001" as any));
  });
});
