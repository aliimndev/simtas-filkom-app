import { beforeAll, afterAll, describe, expect, it } from "bun:test";
import { createApp } from "../src/app";
import { loadConfig } from "../src/config";
import { signAccessToken } from "../src/services/token";
import { getDb } from "../src/db";
import { schema } from "@sims/db";
import { eq } from "drizzle-orm";
import { consultationLogsRoutes } from "../src/routes/consultationLogs";

const TEST_DB_URL = process.env.DATABASE_URL ?? "postgres://postgres@localhost:5433/simtas";
const cfg = loadConfig({ ...process.env, NODE_ENV: "test", DATABASE_URL: TEST_DB_URL } as any);
const app = createApp(cfg);
app.route("/api/v1/consultation-logs", consultationLogsRoutes);
const db = getDb(cfg.databaseUrl);

const DUMMY_HASH = "$2a$12$qMPO1EgF0zmpDh4W49ERVOfOxF28jsItEaiKEKZCWOL9NoKX3U7iC";

let roleMahasiswa = 1;
let roleDosen = 1;

const student = { id: "aaaaaaaa-0000-0000-0000-000000000001", email: "mhs-consl@filkom.ac.id", fullName: "Student Consl" };
const supervisor = { id: "bbbbbbbb-0000-0000-0000-000000000002", email: "dsn-consl@filkom.ac.id", fullName: "Sup Consl" };
const unrelated = { id: "cccccccc-0000-0000-0000-000000000003", email: "unr-consl@filkom.ac.id", fullName: "Unrelated Consl" };
let academicYearId = "";
let thesisId = "";

let seq = 0;
const req = (path: string, init: Record<string, any> = {}) =>
  app.request(path, {
    ...init,
    headers: { "x-forwarded-for": `consl-${seq++}`, ...(init.headers ?? {}) },
  });

async function roleIdByName(name: string): Promise<number> {
  const r = await db.select().from(schema.roles).where(eq(schema.roles.name, name));
  return (r[0] as any)?.id ?? 1;
}

beforeAll(async () => {
  roleMahasiswa = await roleIdByName("mahasiswa");
  roleDosen = await roleIdByName("dosen_pembimbing");

  for (const u of [student, supervisor, unrelated]) {
    await db.delete(schema.users).where(eq(schema.users.id, u.id as any));
  }
  await db.insert(schema.users).values([
    { id: student.id, email: student.email, fullName: student.fullName, roleId: roleMahasiswa, passwordHash: DUMMY_HASH, isActive: true, tokenVersion: 0, mustChangePassword: false } as any,
    { id: supervisor.id, email: supervisor.email, fullName: supervisor.fullName, roleId: roleDosen, passwordHash: DUMMY_HASH, isActive: true, tokenVersion: 0, mustChangePassword: false } as any,
    { id: unrelated.id, email: unrelated.email, fullName: unrelated.fullName, roleId: roleMahasiswa, passwordHash: DUMMY_HASH, isActive: true, tokenVersion: 0, mustChangePassword: false } as any,
  ] as any);

  const [ay] = await db
    .insert(schema.academicYears)
    .values({ name: "2026/2027", semester: "ganjil", startDate: "2026-08-01", endDate: "2026-12-31", isActive: false } as any)
    .returning();
  academicYearId = (ay as any).id;

  const [th] = await db
    .insert(schema.theses)
    .values({ studentId: student.id, academicYearId, title: "Thesis Consl", thesisType: "skripsi", status: "in_progress", submittedAt: new Date() } as any)
    .returning();
  thesisId = (th as any).id;

  await db
    .insert(schema.thesisSupervisors)
    .values({ thesisId, supervisorId: supervisor.id, assignedBy: student.id } as any);
});

afterAll(async () => {
  await db.delete(schema.consultationLogs).where(eq(schema.consultationLogs.thesisId, thesisId as any));
  await db.delete(schema.thesisSupervisors).where(eq(schema.thesisSupervisors.thesisId, thesisId as any));
  await db.delete(schema.theses).where(eq(schema.theses.id, thesisId as any));
  await db.delete(schema.academicYears).where(eq(schema.academicYears.id, academicYearId as any));
  for (const u of [student, supervisor, unrelated]) {
    await db.delete(schema.users).where(eq(schema.users.id, u.id as any));
  }
});

const auth = (id: string, role: string) => signAccessToken(id, role, 0).then((t) => ({ authorization: `Bearer ${t}` }));

describe("consultation logs", () => {
  it("student creates a consultation (status pending)", async () => {
    const res = await req("/api/v1/consultation-logs", {
      method: "POST",
      headers: { "content-type": "application/json", ...(await auth(student.id, "MAHASISWA")) },
      body: JSON.stringify({ thesisId, consultationDate: "2026-08-20", topicsDiscussed: "Bab 1 review" }),
    });
    expect(res.status).toBe(201);
    const body = (await res.json()) as any;
    expect(body.data.status).toBe("pending");
    expect(body.data.createdBy).toBe(student.id);
  });

  it("rejects missing topics_discussed with 400", async () => {
    const res = await req("/api/v1/consultation-logs", {
      method: "POST",
      headers: { "content-type": "application/json", ...(await auth(student.id, "MAHASISWA")) },
      body: JSON.stringify({ thesisId, consultationDate: "2026-08-20" }),
    });
    expect(res.status).toBe(400);
  });

  it("supervisor lists and approves the consultation", async () => {
    const list = await req(`/api/v1/consultation-logs?thesisId=${thesisId}`, {
      headers: await auth(supervisor.id, "DOSEN_PEMBIMBING"),
    });
    expect(list.status).toBe(200);
    const listBody = (await list.json()) as any;
    expect(listBody.meta.total).toBeGreaterThanOrEqual(1);
    const id = listBody.data[0].id;

    const appr = await req(`/api/v1/consultation-logs/${id}/approve?thesisId=${thesisId}`, {
      method: "PATCH",
      headers: await auth(supervisor.id, "DOSEN_PEMBIMBING"),
    });
    expect(appr.status).toBe(200);
    const apprBody = (await appr.json()) as any;
    expect(apprBody.data.status).toBe("approved");
    expect(apprBody.data.approvedBy).toBe(supervisor.id);
  });

  it("forbidden for unrelated user (403)", async () => {
    const res = await req(`/api/v1/consultation-logs?thesisId=${thesisId}`, {
      headers: await auth(unrelated.id, "MAHASISWA"),
    });
    expect(res.status).toBe(403);

    const get = await req(`/api/v1/consultation-logs/${"aaaaaaaa-0000-0000-0000-000000000009"}?thesisId=${thesisId}`, {
      headers: await auth(unrelated.id, "MAHASISWA"),
    });
    expect(get.status).toBe(403);
  });

  it("returns 404 for a non-existent consultation id (scoped)", async () => {
    const get = await req(`/api/v1/consultation-logs/${"aaaaaaaa-0000-0000-0000-000000000009"}?thesisId=${thesisId}`, {
      headers: await auth(student.id, "MAHASISWA"),
    });
    expect(get.status).toBe(404);
  });

  it("summary returns counts", async () => {
    const res = await req(`/api/v1/consultation-logs/summary?thesisId=${thesisId}`, {
      headers: await auth(student.id, "MAHASISWA"),
    });
    expect(res.status).toBe(200);
    const body = (await res.json()) as any;
    expect(body.data.totalConsultations).toBeGreaterThanOrEqual(1);
    expect(body.data.approvedCount).toBeGreaterThanOrEqual(1);
  });
});
