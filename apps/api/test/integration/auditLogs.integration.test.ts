import { afterAll, beforeAll, describe, expect, it } from "bun:test";
import { eq } from "drizzle-orm";
import { createApp } from "../../src/app";
import { loadConfig } from "../../src/config";
import { signAccessToken } from "../../src/modules/auth";
import { getDb } from "../../src/db";
import { schema } from "@sims/db";
import { auditLogsRoutes } from "../../src/modules/audit-logs";

const TEST_DB_URL = process.env.DATABASE_URL ?? "postgres://postgres@localhost:5433/simtas";
const cfg = loadConfig({ ...process.env, NODE_ENV: "test", DATABASE_URL: TEST_DB_URL } as any);
const db = getDb(cfg.databaseUrl);

const app = createApp(cfg);
app.route("/api/v1/audit-logs", auditLogsRoutes);

// Each request gets a unique IP so the per-IP rate limiter doesn't accumulate across tests.
let seq = 0;
const req = (path: string, init: Record<string, any> = {}) =>
  app.request(path, {
    ...init,
    headers: { "x-forwarded-for": `audit-${seq++}`, ...(init.headers ?? {}) },
  });

const UNIQUE_ACTION = `test.action.${Date.now()}`;
const OTHER_ACTION = `test.other.${Date.now()}`;
let insertedIds: string[] = [];
let adminTok = "";
let mhsTok = "";

beforeAll(async () => {
  adminTok = await signAccessToken("admin-audit-1", "ADMIN_FAKULTAS", 0);
  mhsTok = await signAccessToken("mhs-audit-1", "MAHASISWA", 0);

  const rows = await db
    .insert(schema.auditLogs)
    .values([
      {
        action: UNIQUE_ACTION,
        entityType: "thesis",
        entityId: crypto.randomUUID(),
        oldValue: { title: "old" },
        newValue: { title: "new" },
        ipAddress: "127.0.0.1",
      },
      { action: OTHER_ACTION, entityType: "user", entityId: crypto.randomUUID() },
    ])
    .returning({ id: schema.auditLogs.id });
  insertedIds = rows.map((r: any) => r.id);
});

afterAll(async () => {
  await db.delete(schema.auditLogs).where(eq(schema.auditLogs.action, UNIQUE_ACTION));
  await db.delete(schema.auditLogs).where(eq(schema.auditLogs.action, OTHER_ACTION));
});

describe("audit logs", () => {
  it("lists audit logs for admin (200, envelope + meta, jsonb as JSON)", async () => {
    const res = await req("/api/v1/audit-logs", { headers: { authorization: `Bearer ${adminTok}` } });
    expect(res.status).toBe(200);
    const body = (await res.json()) as any;
    expect(body.success).toBe(true);
    expect(Array.isArray(body.data)).toBe(true);
    expect(body.meta).toMatchObject({ page: 1, perPage: 50 });
    expect(body.meta.total).toBeGreaterThanOrEqual(2);

    const mine = body.data.find((d: any) => d.action === UNIQUE_ACTION);
    expect(mine).toBeDefined();
    expect(mine.oldValue).toEqual({ title: "old" });
    expect(mine.newValue).toEqual({ title: "new" });
  });

  it("filters by action query param", async () => {
    const res = await req(`/api/v1/audit-logs?action=${UNIQUE_ACTION}`, {
      headers: { authorization: `Bearer ${adminTok}` },
    });
    expect(res.status).toBe(200);
    const body = (await res.json()) as any;
    expect(body.data.length).toBe(1);
    expect(body.data[0].action).toBe(UNIQUE_ACTION);
    expect(body.meta.total).toBe(1);
  });

  it("gets a single audit log by id", async () => {
    const res = await req(`/api/v1/audit-logs/${insertedIds[0]}`, {
      headers: { authorization: `Bearer ${adminTok}` },
    });
    expect(res.status).toBe(200);
    const body = (await res.json()) as any;
    expect(body.data.id).toBe(insertedIds[0]);
    expect(body.data.action).toBe(UNIQUE_ACTION);
  });

  it("returns 404 for an unknown id", async () => {
    const res = await req(`/api/v1/audit-logs/${crypto.randomUUID()}`, {
      headers: { authorization: `Bearer ${adminTok}` },
    });
    expect(res.status).toBe(404);
    expect((await res.json() as any).error.code).toBe("NOT_FOUND");
  });

  it("forbids a mahasiswa (403)", async () => {
    const res = await req("/api/v1/audit-logs", { headers: { authorization: `Bearer ${mhsTok}` } });
    expect(res.status).toBe(403);
    expect((await res.json() as any).error.code).toBe("FORBIDDEN");
  });

  it("rejects a missing token (401)", async () => {
    const res = await req("/api/v1/audit-logs");
    expect(res.status).toBe(401);
  });
});
