import { afterAll, beforeAll, describe, expect, it } from "bun:test";
import { Hono } from "hono";
import { eq } from "drizzle-orm";
import { getDb } from "../src/db";
import { loadConfig } from "../src/config";
import { signAccessToken } from "../src/services/token";
import { schema } from "@sims/db";
import { notificationsRoutes } from "../src/routes/notifications";

const TEST_DB_URL = process.env.DATABASE_URL || "postgres://postgres@localhost:5433/simtas";
process.env.DATABASE_URL = TEST_DB_URL;
const cfg = loadConfig({ ...process.env, NODE_ENV: "test", DATABASE_URL: TEST_DB_URL } as any);
const db = getDb(cfg.databaseUrl);

const app = new Hono();
app.route("/api/v1/notifications", notificationsRoutes);

const OWNER_ID = "aaaaaaaa-0000-0000-0000-000000000001";
const OTHER_ID = "bbbbbbbb-0000-0000-0000-000000000002";

let roleId = 1;
let ownerToken = "";
let otherToken = "";

async function resolveRoleId(name: string): Promise<number> {
  const r: any = await db.select().from(schema.roles).where(eq(schema.roles.name, name));
  return r[0]?.id ?? 1;
}

beforeAll(async () => {
  roleId = await resolveRoleId("mahasiswa");
  await db.delete(schema.notifications).where(eq(schema.notifications.userId, OWNER_ID));
  await db.delete(schema.notifications).where(eq(schema.notifications.userId, OTHER_ID));
  await db.delete(schema.users).where(eq(schema.users.id, OWNER_ID));
  await db.delete(schema.users).where(eq(schema.users.id, OTHER_ID));

  await db.insert(schema.users).values([
    {
      id: OWNER_ID,
      email: "notif-owner@filkom.ac.id",
      fullName: "Owner",
      roleId,
      passwordHash: "x",
      loginAttemptCount: 0,
      tokenVersion: 0,
      mustChangePassword: false,
      isActive: true,
    } as any,
    {
      id: OTHER_ID,
      email: "notif-other@filkom.ac.id",
      fullName: "Other",
      roleId,
      passwordHash: "x",
      loginAttemptCount: 0,
      tokenVersion: 0,
      mustChangePassword: false,
      isActive: true,
    } as any,
  ]);

  await db.insert(schema.notifications).values([
    { userId: OWNER_ID, title: "Unread one", message: "m1", type: "INFO", isRead: false } as any,
    { userId: OWNER_ID, title: "Unread two", message: "m2", type: "INFO", isRead: false } as any,
    { userId: OTHER_ID, title: "Other's note", message: "m3", type: "INFO", isRead: false } as any,
  ]);

  ownerToken = await signAccessToken(OWNER_ID, "MAHASISWA", 0);
  otherToken = await signAccessToken(OTHER_ID, "MAHASISWA", 0);
});

afterAll(async () => {
  await db.delete(schema.notifications).where(eq(schema.notifications.userId, OWNER_ID));
  await db.delete(schema.notifications).where(eq(schema.notifications.userId, OTHER_ID));
  await db.delete(schema.users).where(eq(schema.users.id, OWNER_ID));
  await db.delete(schema.users).where(eq(schema.users.id, OTHER_ID));
});

let seq = 0;
const req = (path: string, init: Record<string, any> = {}) =>
  app.request(path, {
    ...init,
    headers: { "x-forwarded-for": `notif-${seq++}`, ...(init.headers ?? {}) },
  });

const auth = (token: string) => ({ authorization: `Bearer ${token}` });
const base = "/api/v1/notifications";

describe("notifications", () => {
  it("lists only the current user's notifications", async () => {
    const res = await req(base, { headers: auth(ownerToken) });
    expect(res.status).toBe(200);
    const body = (await res.json()) as any[];
    expect(body.length).toBe(2);
    expect(body.every((n) => n.id)).toBe(true);
  });

  it("filters unread only with ?unread=true", async () => {
    const res = await req(`${base}?unread=true`, { headers: auth(ownerToken) });
    expect(res.status).toBe(200);
    const body = (await res.json()) as any[];
    expect(body.length).toBe(2);
    expect(body.every((n) => n.is_read === false)).toBe(true);
  });

  it("returns 401 without a token", async () => {
    expect((await req(base)).status).toBe(401);
  });

  it("marks one notification read", async () => {
    const list: any[] = await (await req(base, { headers: auth(ownerToken) })).json();
    const id = list[0].id;
    const res = await req(`${base}/${id}/read`, { method: "PATCH", headers: auth(ownerToken) });
    expect(res.status).toBe(200);
    const after: any = await (await req(`${base}/${id}`, { headers: auth(ownerToken) })).json();
    expect(after.is_read).toBe(true);
    expect(after.read_at).not.toBeNull();
  });

  it("returns 404 for a non-existent notification id", async () => {
    const res = await req(`${base}/cccccccc-0000-0000-0000-000000000009/read`, {
      method: "PATCH",
      headers: auth(ownerToken),
    });
    expect(res.status).toBe(404);
  });

  it("returns 404 when reading another user's notification (ownership enforced)", async () => {
    const other: any[] = await (await req(base, { headers: auth(otherToken) })).json();
    const otherId = other[0].id;
    expect((await req(`${base}/${otherId}`, { headers: auth(ownerToken) })).status).toBe(404);
    expect(
      (await req(`${base}/${otherId}/read`, { method: "PATCH", headers: auth(ownerToken) })).status,
    ).toBe(404);
  });

  it("marks all read and reflects in unread-count", async () => {
    const before: any = await (await req(`${base}/unread-count`, { headers: auth(ownerToken) })).json();
    expect(before.unread_count).toBeGreaterThan(0);

    const markAll = await req(`${base}/read-all`, { method: "POST", headers: auth(ownerToken) });
    expect(markAll.status).toBe(200);

    const after: any = await (await req(`${base}/unread-count`, { headers: auth(ownerToken) })).json();
    expect(after.unread_count).toBe(0);
  });

  it("unread-count ignores other users", async () => {
    const other: any = await (await req(`${base}/unread-count`, { headers: auth(otherToken) })).json();
    expect(other.unread_count).toBe(1);
  });
});
