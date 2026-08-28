import { beforeAll, afterAll, describe, expect, it } from "bun:test";
import { createApp } from "../src/app";
import { loadConfig } from "../src/config";
import { signAccessToken } from "../src/services/token";
import { getDb } from "../src/db";
import { schema } from "@sims/db";
import { hashPassword } from "../src/services/password";
import { eq } from "drizzle-orm";
import { usersRoutes } from "../src/routes/users";

const TEST_DB_URL = process.env.DATABASE_URL ?? "postgres://postgres@localhost:5433/simtas";
const cfg = loadConfig({ ...process.env, NODE_ENV: "test", DATABASE_URL: TEST_DB_URL } as any);
const app = createApp(cfg);
app.route("/api/v1/users", usersRoutes);

let seq = 0;
const req = (path: string, init: Record<string, any> = {}) =>
  app.request(path, {
    ...init,
    headers: { "x-forwarded-for": `users-${seq++}`, ...(init.headers ?? {}) },
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
let adminId = "";
let adminToken = "";
let mhsId = "";

beforeAll(async () => {
  await seedRoleIds();
  adminId = await makeUser("admin-port@example.com", "ADMIN_FAKULTAS", "Port Admin");
  created.push(adminId);
  mhsId = await makeUser("mhs-port@example.com", "MAHASISWA", "Port Mhs");
  created.push(mhsId);
  adminToken = await signAccessToken(adminId, "ADMIN_FAKULTAS", 0);
});

afterAll(async () => {
  for (const id of created) {
    await db().delete(schema.users).where(eq(schema.users.id, id));
  }
});

describe("users parity", () => {
  it("admin lists users", async () => {
    const res = await req("/api/v1/users", { headers: { authorization: `Bearer ${adminToken}` } });
    expect(res.status).toBe(200);
    const body = (await res.json()) as any;
    expect(Array.isArray(body.data)).toBe(true);
  });

  it("filters by role", async () => {
    const res = await req("/api/v1/users?role=MAHASISWA", { headers: { authorization: `Bearer ${adminToken}` } });
    expect(res.status).toBe(200);
  });

  it("gets a user by id", async () => {
    const res = await req(`/api/v1/users/${mhsId}`, { headers: { authorization: `Bearer ${adminToken}` } });
    expect(res.status).toBe(200);
    const body = (await res.json()) as any;
    expect(body.id).toBe(mhsId);
  });

  it("returns 404 for unknown id", async () => {
    const res = await req(`/api/v1/users/${crypto.randomUUID()}`, { headers: { authorization: `Bearer ${adminToken}` } });
    expect(res.status).toBe(404);
  });

  it("admin creates a user", async () => {
    const res = await req("/api/v1/users", {
      method: "POST",
      headers: { "content-type": "application/json", authorization: `Bearer ${adminToken}` },
      body: JSON.stringify({ email: "created-port@example.com", fullName: "Created User", role: "MAHASISWA" }),
    });
    expect(res.status).toBe(201);
    const body = (await res.json()) as any;
    expect(body.id).toBeDefined();
    created.push(body.id);
  });

  it("rejects duplicate email with 409", async () => {
    const res = await req("/api/v1/users", {
      method: "POST",
      headers: { "content-type": "application/json", authorization: `Bearer ${adminToken}` },
      body: JSON.stringify({ email: "created-port@example.com", fullName: "Dup", role: "MAHASISWA" }),
    });
    expect(res.status).toBe(409);
  });

  it("rejects an invalid role with 400", async () => {
    const res = await req("/api/v1/users", {
      method: "POST",
      headers: { "content-type": "application/json", authorization: `Bearer ${adminToken}` },
      body: JSON.stringify({ email: "badrole-port@example.com", fullName: "Bad", role: "WIBBLE" }),
    });
    expect(res.status).toBe(400);
  });

  it("forbids a role outside the allowed set (403)", async () => {
    const guestToken = await signAccessToken(crypto.randomUUID(), "GUEST", 0);
    const res = await req("/api/v1/users", { headers: { authorization: `Bearer ${guestToken}` } });
    expect(res.status).toBe(403);
  });

  it("updates a profile; email/role are immutable", async () => {
    const res = await req(`/api/v1/users/${mhsId}`, {
      method: "PUT",
      headers: { "content-type": "application/json", authorization: `Bearer ${adminToken}` },
      body: JSON.stringify({ fullName: "Updated Name", email: "nope@example.com", role: "KAPRODI" }),
    });
    expect(res.status).toBe(200);
    const body = (await res.json()) as any;
    expect(body.fullName).toBe("Updated Name");
    expect(body.email).toBe("mhs-port@example.com");
  });

  it("deactivates a user", async () => {
    const res = await req(`/api/v1/users/${mhsId}/deactivate`, {
      method: "PATCH",
      headers: { authorization: `Bearer ${adminToken}` },
    });
    expect(res.status).toBe(200);
  });

  it("cannot deactivate self (403)", async () => {
    const res = await req(`/api/v1/users/${adminId}/deactivate`, {
      method: "PATCH",
      headers: { authorization: `Bearer ${adminToken}` },
    });
    expect(res.status).toBe(403);
  });
});
