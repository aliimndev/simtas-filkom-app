import { beforeAll, afterAll, describe, expect, it } from "bun:test";
import { createApp } from "../src/app";
import { loadConfig } from "../src/config";
import { signAccessToken } from "../src/services/token";
import { getDb } from "../src/db";
import { schema } from "@sims/db";
import { eq } from "drizzle-orm";
import { academicYearsRoutes } from "../src/routes/academicYears";

const TEST_DB_URL = process.env.DATABASE_URL ?? "postgres://postgres@localhost:5433/simtas";
const cfg = loadConfig({ ...process.env, NODE_ENV: "test", DATABASE_URL: TEST_DB_URL } as any);
const app = createApp(cfg);
app.route("/api/v1/academic-years", academicYearsRoutes);

const db = getDb(cfg.databaseUrl);

let seq = 0;
const req = (path: string, init: Record<string, any> = {}) =>
  app.request(path, {
    ...init,
    headers: { "x-forwarded-for": `ay-${seq++}`, ...(init.headers ?? {}) },
  });

async function getRoleId(roleName: string): Promise<number> {
  const rows = await db.select().from(schema.roles);
  const r = rows.find((x) => String(x.name).toUpperCase() === roleName);
  if (!r) throw new Error(`role ${roleName} not found`);
  return r.id;
}

const adminId = crypto.randomUUID();
const mhsId = crypto.randomUUID();
const adminTok = await signAccessToken(adminId, "ADMIN_FAKULTAS", 0);
const mhsTok = await signAccessToken(mhsId, "MAHASISWA", 0);

const createdYearIds: string[] = [];

beforeAll(async () => {
  await db.insert(schema.users).values([
    {
      id: adminId,
      email: `ay-admin-${adminId}@example.com`,
      passwordHash: "x",
      fullName: "AY Admin",
      roleId: await getRoleId("ADMIN_FAKULTAS"),
    },
    {
      id: mhsId,
      email: `ay-mhs-${mhsId}@example.com`,
      passwordHash: "x",
      fullName: "AY Mahasiswa",
      roleId: await getRoleId("MAHASISWA"),
    },
  ]);
});

afterAll(async () => {
  if (createdYearIds.length) {
    for (const id of createdYearIds) {
      await db.delete(schema.academicYears).where(eq(schema.academicYears.id, id));
    }
  }
  await db.delete(schema.users).where(eq(schema.users.id, adminId));
  await db.delete(schema.users).where(eq(schema.users.id, mhsId));
});

const auth = (tok: string) => ({ authorization: `Bearer ${tok}`, "content-type": "application/json" });
const sample = (suffix: string) => ({
  name: `2026/2027-${suffix}`,
  semester: "ganjil",
  startDate: "2026-09-01",
  endDate: "2027-01-31",
});

describe("academic-years parity", () => {
  it("creates an academic year (201) and returns it on GET", async () => {
    const create = await req("/api/v1/academic-years", {
      method: "POST",
      headers: auth(adminTok),
      body: JSON.stringify(sample("a")),
    });
    expect(create.status).toBe(201);
    const body = (await create.json()) as any;
    expect(body.id).toBeDefined();
    expect(body.name).toBe(`2026/2027-a`);
    createdYearIds.push(body.id);

    const get = await req(`/api/v1/academic-years/${body.id}`, { headers: auth(adminTok) });
    expect(get.status).toBe(200);
    const got = (await get.json()) as any;
    expect(got.id).toBe(body.id);
  });

  it("rejects invalid semester (400)", async () => {
    const res = await req("/api/v1/academic-years", {
      method: "POST",
      headers: auth(adminTok),
      body: JSON.stringify({ ...sample("bad"), semester: "panas" }),
    });
    expect(res.status).toBe(400);
  });

  it("activate sets exactly one active year (deactivates others)", async () => {
    const a = await req("/api/v1/academic-years", {
      method: "POST",
      headers: auth(adminTok),
      body: JSON.stringify(sample("act-a")),
    });
    const b = await req("/api/v1/academic-years", {
      method: "POST",
      headers: auth(adminTok),
      body: JSON.stringify(sample("act-b")),
    });
    const idA = ((await a.json()) as any).id;
    const idB = ((await b.json()) as any).id;
    createdYearIds.push(idA, idB);

    await req(`/api/v1/academic-years/${idA}/activate`, { method: "PATCH", headers: auth(adminTok) });
    await req(`/api/v1/academic-years/${idB}/activate`, { method: "PATCH", headers: auth(adminTok) });

    const list = await req("/api/v1/academic-years", { headers: auth(adminTok) });
    const years = (await list.json()) as any[];
    const active = years.filter((y) => y.isActive);
    expect(active.length).toBe(1);
    expect(active[0].id).toBe(idB);
  });

  it("returns 404 for unknown id (GET and activate)", async () => {
    const missing = crypto.randomUUID();
    const get = await req(`/api/v1/academic-years/${missing}`, { headers: auth(adminTok) });
    expect(get.status).toBe(404);
    const act = await req(`/api/v1/academic-years/${missing}/activate`, { method: "PATCH", headers: auth(adminTok) });
    expect(act.status).toBe(404);
  });

  it("forbids MAHASISWA (403)", async () => {
    const res = await req("/api/v1/academic-years", {
      method: "POST",
      headers: auth(mhsTok),
      body: JSON.stringify(sample("no")),
    });
    expect(res.status).toBe(403);
    const list = await req("/api/v1/academic-years", { headers: auth(mhsTok) });
    expect(list.status).toBe(403);
  });
});
