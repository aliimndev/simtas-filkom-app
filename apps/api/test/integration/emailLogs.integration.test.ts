import { beforeAll, afterAll, describe, expect, it } from "bun:test";
import { eq } from "drizzle-orm";
import { createApp } from "../../src/app";
import { loadConfig } from "../../src/config";
import { getDb } from "../../src/db";
import { schema } from "@sims/db";
import { signAccessToken } from "../../src/modules/auth";
import { emailLogsRoutes } from "../../src/modules/email-logs";

const TEST_DB_URL = process.env.DATABASE_URL ?? "postgres://postgres@localhost:5433/simtas";
const cfg = loadConfig({ ...process.env, NODE_ENV: "test", DATABASE_URL: TEST_DB_URL } as any);
const db = getDb(cfg.databaseUrl);
const app = createApp(cfg);
app.route("/api/v1/email-logs", emailLogsRoutes);

// Each request gets a unique IP so the per-IP rate limiters don't accumulate across tests.
let seq = 0;
const req = (path: string, init: Record<string, any> = {}) =>
  app.request(path, {
    ...init,
    headers: { "x-forwarded-for": `emaillogs-${seq++}`, ...(init.headers ?? {}) },
  });

const adminTok = await signAccessToken("admin-email-1", "ADMIN_FAKULTAS", 0);
const mhsTok = await signAccessToken("mhs-email-1", "MAHASISWA", 0);
const auth = (t: string) => ({ authorization: `Bearer ${t}` });

let eventA: string;
let rowId: string;

beforeAll(async () => {
  eventA = `email_log_test_${Date.now()}_${Math.floor(Math.random() * 1e9)}`;
  const inserted: any = await db
    .insert(schema.emailLogs)
    .values([
      {
        recipientEmail: "a@filkom.ac.id",
        eventType: eventA,
        subject: "Subjek A",
        status: "sent",
        provider: "resend",
        attempts: 1,
      },
      {
        recipientEmail: "b@filkom.ac.id",
        eventType: eventA,
        subject: "Subjek B",
        status: "failed",
        provider: "resend",
        errorMessage: "boom",
        attempts: 3,
      },
    ])
    .returning();
  rowId = inserted[0].id;
});

afterAll(async () => {
  await db.delete(schema.emailLogs).where(eq(schema.emailLogs.eventType, eventA));
});

describe("emailLogs", () => {
  it("admin lists email logs", async () => {
    const res = await req("/api/v1/email-logs", { headers: auth(adminTok) });
    expect(res.status).toBe(200);
    const body = (await res.json()) as any;
    expect(Array.isArray(body.data)).toBe(true);
    expect(body.total).toBeGreaterThanOrEqual(2);
    expect(body.limit).toBeGreaterThan(0);
  });

  it("filters by event_type", async () => {
    const res = await req(`/api/v1/email-logs?event_type=${eventA}`, { headers: auth(adminTok) });
    expect(res.status).toBe(200);
    const body = (await res.json()) as any;
    expect(body.total).toBe(2);
    expect(body.data.every((r: any) => r.eventType === eventA)).toBe(true);
  });

  it("filters by status", async () => {
    const res = await req(`/api/v1/email-logs?status=failed`, { headers: auth(adminTok) });
    expect(res.status).toBe(200);
    const body = (await res.json()) as any;
    expect(body.data.every((r: any) => r.status === "failed")).toBe(true);
  });

  it("gets a single email log by id", async () => {
    const res = await req(`/api/v1/email-logs/${rowId}`, { headers: auth(adminTok) });
    expect(res.status).toBe(200);
    const body = (await res.json()) as any;
    expect(body.data.id).toBe(rowId);
  });

  it("returns 404 for an unknown id", async () => {
    const res = await req(`/api/v1/email-logs/00000000-0000-0000-0000-000000000000`, {
      headers: auth(adminTok),
    });
    expect(res.status).toBe(404);
    expect((await res.json() as any).error.code).toBe("NOT_FOUND");
  });

  it("forbids a mahasiswa (403)", async () => {
    const res = await req("/api/v1/email-logs", { headers: auth(mhsTok) });
    expect(res.status).toBe(403);
    expect((await res.json() as any).error.code).toBe("FORBIDDEN");
  });

  it("rejects requests without a token (401)", async () => {
    const res = await req("/api/v1/email-logs");
    expect(res.status).toBe(401);
  });
});
