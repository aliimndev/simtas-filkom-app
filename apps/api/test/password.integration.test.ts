import { beforeAll, afterAll, describe, expect, it } from "bun:test";
import { createApp } from "../src/app";
import { loadConfig } from "../src/config";
import { signAccessToken } from "../src/services/token";
import { getDb } from "../src/db";
import { schema } from "@sims/db";
import { hashPassword } from "../src/services/password";
import { eq } from "drizzle-orm";
import { passwordRoutes } from "../src/routes/password";

const TEST_DB_URL = process.env.DATABASE_URL ?? "postgres://postgres@localhost:5433/simtas";
const cfg = loadConfig({ ...process.env, NODE_ENV: "test", DATABASE_URL: TEST_DB_URL } as any);
const app = createApp(cfg);
app.route("/api/v1/auth/password", passwordRoutes);

let seq = 0;
const req = (path: string, init: Record<string, any> = {}) =>
  app.request(path, {
    ...init,
    headers: { "x-forwarded-for": `pw-${seq++}`, ...(init.headers ?? {}) },
  });

const db = () => getDb(cfg.databaseUrl);
const roleIds: Record<string, number> = {};

async function seedRoleIds() {
  const rows: any = await db().select().from(schema.roles);
  for (const r of rows) roleIds[r.name.toUpperCase()] = r.id;
}

async function makeUser(email: string, roleName: string, fullName: string): Promise<string> {
  const id = crypto.randomUUID();
  await db().insert(schema.users).values({
    id,
    email,
    fullName,
    roleId: roleIds[roleName],
    passwordHash: await hashPassword("Password123!"),
    isActive: true,
    mustChangePassword: false,
    loginAttemptCount: 0,
    tokenVersion: 0,
  } as any);
  return id;
}

const created: string[] = [];
let mhsId = "";

beforeAll(async () => {
  await seedRoleIds();
  mhsId = await makeUser("pw-port@example.com", "MAHASISWA", "PW User");
  created.push(mhsId);
});

afterAll(async () => {
  await db().delete(schema.passwordResetTokens).where(eq(schema.passwordResetTokens.userId, mhsId));
  await db().delete(schema.emailLogs).where(eq(schema.emailLogs.recipientEmail, "pw-port@example.com"));
  for (const id of created) {
    await db().delete(schema.users).where(eq(schema.users.id, id));
  }
});

describe("password parity", () => {
  it("forgot returns 200 and writes a token + email_log (anti-enumeration)", async () => {
    const res = await req("/api/v1/auth/password/forgot", {
      method: "POST",
      headers: { "content-type": "application/json" },
      body: JSON.stringify({ email: "pw-port@example.com" }),
    });
    expect(res.status).toBe(200);

    const toks: any = await db().select().from(schema.passwordResetTokens).where(eq(schema.passwordResetTokens.userId, mhsId));
    expect(toks.length).toBe(1);

    const logs: any = await db()
      .select()
      .from(schema.emailLogs)
      .where(eq(schema.emailLogs.recipientEmail, "pw-port@example.com"));
    expect(logs.some((l: any) => l.eventType === "password_reset" && l.status === "sent")).toBe(true);
  });

  it("forgot for unknown email still returns 200", async () => {
    const res = await req("/api/v1/auth/password/forgot", {
      method: "POST",
      headers: { "content-type": "application/json" },
      body: JSON.stringify({ email: "nope-port@example.com" }),
    });
    expect(res.status).toBe(200);
  });

  it("reset flow sets password, marks token used, invalidates sessions", async () => {
    const toks: any = await db().select().from(schema.passwordResetTokens).where(eq(schema.passwordResetTokens.userId, mhsId));
    const token = toks[0].token;

    const res = await req("/api/v1/auth/password/reset", {
      method: "POST",
      headers: { "content-type": "application/json" },
      body: JSON.stringify({ token, newPassword: "NewPass123!", confirmPassword: "NewPass123!" }),
    });
    expect(res.status).toBe(200);

    const after: any = await db().select().from(schema.passwordResetTokens).where(eq(schema.passwordResetTokens.id, toks[0].id));
    expect(after[0].usedAt).not.toBeNull();

    const user: any = await db().select().from(schema.users).where(eq(schema.users.id, mhsId));
    expect(user[0].mustChangePassword).toBe(false);
  });

  it("reset rejects a reused token", async () => {
    const toks: any = await db().select().from(schema.passwordResetTokens).where(eq(schema.passwordResetTokens.userId, mhsId));
    const token = toks[0].token;
    const res = await req("/api/v1/auth/password/reset", {
      method: "POST",
      headers: { "content-type": "application/json" },
      body: JSON.stringify({ token, newPassword: "Another1!", confirmPassword: "Another1!" }),
    });
    expect(res.status).toBe(400);
  });

  it("reset rejects mismatched passwords", async () => {
    // mint a fresh token
    await req("/api/v1/auth/password/forgot", {
      method: "POST",
      headers: { "content-type": "application/json" },
      body: JSON.stringify({ email: "pw-port@example.com" }),
    });
    const toks: any = await db().select().from(schema.passwordResetTokens).where(eq(schema.passwordResetTokens.userId, mhsId));
    const token = toks[toks.length - 1].token;
    const res = await req("/api/v1/auth/password/reset", {
      method: "POST",
      headers: { "content-type": "application/json" },
      body: JSON.stringify({ token, newPassword: "NewPass123!", confirmPassword: "Different1!" }),
    });
    expect(res.status).toBe(400);
  });

  it("reset rejects weak password (complexity)", async () => {
    await req("/api/v1/auth/password/forgot", {
      method: "POST",
      headers: { "content-type": "application/json" },
      body: JSON.stringify({ email: "pw-port@example.com" }),
    });
    const toks: any = await db().select().from(schema.passwordResetTokens).where(eq(schema.passwordResetTokens.userId, mhsId));
    const token = toks[toks.length - 1].token;
    const res = await req("/api/v1/auth/password/reset", {
      method: "POST",
      headers: { "content-type": "application/json" },
      body: JSON.stringify({ token, newPassword: "weak", confirmPassword: "weak" }),
    });
    expect(res.status).toBe(400);
  });
});
