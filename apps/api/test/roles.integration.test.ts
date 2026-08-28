import { beforeAll, afterAll, describe, expect, it } from "bun:test";
import { createApp } from "../src/app";
import { loadConfig } from "../src/config";
import { signAccessToken } from "../src/services/token";
import { getDb } from "../src/db";
import { schema } from "@sims/db";
import { eq } from "drizzle-orm";
import { rolesRoutes } from "../src/routes/roles";

const TEST_DB_URL = process.env.DATABASE_URL ?? "postgres://postgres@localhost:5433/simtas";
if (!process.env.DATABASE_URL) process.env.DATABASE_URL = TEST_DB_URL;
const cfg = loadConfig({ ...process.env, NODE_ENV: "test", DATABASE_URL: TEST_DB_URL } as any);
const db = getDb(cfg.databaseUrl);

const app = createApp(cfg);
app.route("/api/v1/roles", rolesRoutes);

let seq = 0;
const req = (path: string, init: Record<string, any> = {}) =>
  app.request(path, {
    ...init,
    headers: { "x-forwarded-for": `roles-${seq++}`, ...(init.headers ?? {}) },
  });

const adminId = "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa";
const mhsId = "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb";
const adminEmail = `roles-admin-${Date.now()}@filkom.ac.id`;
const mhsEmail = `roles-mhs-${Date.now()}@filkom.ac.id`;

let adminTok: string;
let mhsTok: string;
let adminRoleId: number;

beforeAll(async () => {
  const r = await db.select().from(schema.roles).where(eq(schema.roles.name, "admin_fakultas"));
  adminRoleId = (r[0] as any).id;
  const m = await db.select().from(schema.roles).where(eq(schema.roles.name, "mahasiswa"));
  const mhsRoleId = (m[0] as any).id;

  await db.delete(schema.users).where(eq(schema.users.email, adminEmail));
  await db.delete(schema.users).where(eq(schema.users.email, mhsEmail));
  await db.insert(schema.users).values([
    {
      id: adminId,
      email: adminEmail,
      fullName: "Roles Admin",
      roleId: adminRoleId,
      passwordHash: "x",
      loginAttemptCount: 0,
      tokenVersion: 0,
      mustChangePassword: false,
      isActive: true,
    } as any,
    {
      id: mhsId,
      email: mhsEmail,
      fullName: "Roles Mahasiswa",
      roleId: mhsRoleId,
      passwordHash: "x",
      loginAttemptCount: 0,
      tokenVersion: 0,
      mustChangePassword: false,
      isActive: true,
    } as any,
  ]);

  adminTok = await signAccessToken(adminId, "ADMIN_FAKULTAS", 0);
  mhsTok = await signAccessToken(mhsId, "MAHASISWA", 0);
});

afterAll(async () => {
  await db.delete(schema.users).where(eq(schema.users.email, adminEmail));
  await db.delete(schema.users).where(eq(schema.users.email, mhsEmail));
});

const auth = (tok: string) => ({ authorization: `Bearer ${tok}` });
const json = (body: unknown, tok: string) => ({
  method: "POST",
  headers: { "content-type": "application/json", ...auth(tok) },
  body: JSON.stringify(body),
});

describe("roles", () => {
  it("lists roles (admin)", async () => {
    const res = await req("/api/v1/roles", { headers: auth(adminTok) });
    expect(res.status).toBe(200);
    const body = (await res.json()) as any;
    expect(Array.isArray(body.roles)).toBe(true);
    expect(body.roles.length).toBeGreaterThan(0);
  });

  it("gets a role by id (admin)", async () => {
    const res = await req(`/api/v1/roles/${adminRoleId}`, { headers: auth(adminTok) });
    expect(res.status).toBe(200);
    const body = (await res.json()) as any;
    expect(body.role.id).toBe(adminRoleId);
  });

  it("returns 404 for unknown role id (admin)", async () => {
    const res = await req("/api/v1/roles/999999999", { headers: auth(adminTok) });
    expect(res.status).toBe(404);
    expect((await res.json() as any).error.code).toBe("NOT_FOUND");
  });

  it("creates a role (admin)", async () => {
    const name = `role-${Date.now()}-${seq}`;
    const res = await req("/api/v1/roles", json({ name }, adminTok));
    expect(res.status).toBe(201);
    const body = (await res.json()) as any;
    expect(body.role.name).toBe(name);
    await db.delete(schema.roles).where(eq(schema.roles.id, body.role.id));
  });

  it("returns 409 on duplicate role name (admin)", async () => {
    const name = `dup-${Date.now()}-${seq}`;
    const first = await req("/api/v1/roles", json({ name }, adminTok));
    expect(first.status).toBe(201);
    const created = (await first.json()) as any;
    const second = await req("/api/v1/roles", json({ name }, adminTok));
    expect(second.status).toBe(409);
    expect((await second.json() as any).error.code).toBe("CONFLICT");
    await db.delete(schema.roles).where(eq(schema.roles.id, created.role.id));
  });

  it("updates a role name (admin)", async () => {
    const name = `upd-${Date.now()}-${seq}`;
    const created = await req("/api/v1/roles", json({ name }, adminTok));
    const createdBody = (await created.json()) as any;
    const newName = `upd2-${Date.now()}-${seq}`;
    const res = await req(`/api/v1/roles/${createdBody.role.id}`, {
      method: "PUT",
      headers: { "content-type": "application/json", ...auth(adminTok) },
      body: JSON.stringify({ name: newName }),
    });
    expect(res.status).toBe(200);
    expect((await res.json() as any).role.name).toBe(newName);
    await db.delete(schema.roles).where(eq(schema.roles.id, createdBody.role.id));
  });

  it("returns 401 without token", async () => {
    const res = await req("/api/v1/roles");
    expect(res.status).toBe(401);
  });

  it("forbids non-admin (403)", async () => {
    const res = await req("/api/v1/roles", { headers: auth(mhsTok) });
    expect(res.status).toBe(403);
  });
});
