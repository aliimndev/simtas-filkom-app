import { beforeAll, afterAll, describe, expect, it } from "bun:test";
import { createApp } from "../src/app";
import { loadConfig } from "../src/config";
import { getDb } from "../src/db";
import { signAccessToken } from "../src/services/token";
import { schema } from "@sims/db";
import { eq } from "drizzle-orm";
import { thesesRoutes } from "../src/routes/theses";

const TEST_DB_URL = process.env.DATABASE_URL ?? "postgres://postgres@localhost:5433/simtas";
const app = createApp(loadConfig({ ...process.env, NODE_ENV: "test", DATABASE_URL: TEST_DB_URL } as any));
app.route("/api/v1/theses", thesesRoutes);

const db = getDb(TEST_DB_URL);

let seq = 0;
const req = (path: string, init: Record<string, any> = {}) =>
  app.request(path, {
    ...init,
    headers: { "x-forwarded-for": `theses-${seq++}`, ...(init.headers ?? {}) },
  });

const auth = (token: string, init: Record<string, any> = {}) =>
  req(init.path!, { ...init, headers: { authorization: `Bearer ${token}`, ...(init.headers ?? {}) } });

// ── fixtures ──
let studentId: string;
let otherStudentId: string;
let kaprodiId: string;
let supervisorId: string;
let ayId: string;
let roleIdMhs: number;
let roleIdKaprodi: number;
let roleIdDosen: number;

const uniq = (p: string) => `${p}-${Date.now()}-${Math.floor(Math.random() * 1e9)}`;

async function roleIdByName(name: string): Promise<number> {
  const r = await db.select().from(schema.roles).where(eq(schema.roles.name, name));
  if (r.length === 0) throw new Error(`role ${name} not found`);
  return r[0].id;
}

async function makeUser(email: string, fullName: string, roleId: number) {
  const [u] = await db
    .insert(schema.users)
    .values({
      email,
      passwordHash: "$2a$12$qMPO1EgF0zmpDh4W49ERVOfOxF28jsItEaiKEKZCWOL9NoKX3U7iC",
      fullName,
      roleId,
      isActive: true,
      mustChangePassword: false,
    })
    .returning();
  return u.id;
}

beforeAll(async () => {
  roleIdMhs = await roleIdByName("mahasiswa");
  roleIdKaprodi = await roleIdByName("kaprodi");
  roleIdDosen = await roleIdByName("dosen_pembimbing");

  studentId = await makeUser(uniq("mhs@filkom.ac.id"), "Mahasiswa Test", roleIdMhs);
  otherStudentId = await makeUser(uniq("mhs2@filkom.ac.id"), "Mahasiswa Lain", roleIdMhs);
  kaprodiId = await makeUser(uniq("kaprodi@filkom.ac.id"), "Kaprodi Test", roleIdKaprodi);
  supervisorId = await makeUser(uniq("dosen@filkom.ac.id"), "Dosen Pembimbing Test", roleIdDosen);

  const [ay] = await db
    .insert(schema.academicYears)
    .values({ name: "2025/2026", semester: "genap", startDate: "2026-01-01", endDate: "2026-06-30", isActive: true })
    .returning();
  ayId = ay.id;
});

afterAll(async () => {
  await db.delete(schema.thesisSupervisors).where(eq(schema.thesisSupervisors.assignedBy, kaprodiId));
  await db.delete(schema.theses).where(eq(schema.theses.studentId, studentId));
  await db.delete(schema.theses).where(eq(schema.theses.studentId, otherStudentId));
  await db.delete(schema.notifications).where(eq(schema.notifications.userId, studentId));
  await db.delete(schema.notifications).where(eq(schema.notifications.userId, supervisorId));
  await db.delete(schema.notifications).where(eq(schema.notifications.userId, kaprodiId));
  await db.delete(schema.auditLogs).where(eq(schema.auditLogs.userId, kaprodiId));
  await db.delete(schema.auditLogs).where(eq(schema.auditLogs.userId, studentId));
  await db.delete(schema.academicYears).where(eq(schema.academicYears.id, ayId));
  await db.delete(schema.users).where(eq(schema.users.id, studentId));
  await db.delete(schema.users).where(eq(schema.users.id, otherStudentId));
  await db.delete(schema.users).where(eq(schema.users.id, kaprodiId));
  await db.delete(schema.users).where(eq(schema.users.id, supervisorId));
});

const mhsTok = () => signAccessToken(studentId, "MAHASISWA", 0);
const kaprodiTok = () => signAccessToken(kaprodiId, "KAPRODI", 0);
const otherTok = () => signAccessToken(otherStudentId, "MAHASISWA", 0);

const longTitle =
  "Analisis dan implementasi sistem pendukung keputusan berbasis metode profile matching untuk seleksi beasiswa mahasiswa";
const longAbstract = Array(110).fill("Abstrak penelitian ini menjelaskan perancangan implementasi evaluasi sistem informasi akademik secara terukur").join(" ");

describe("theses integration", () => {
  let thesisId: string;

  it("student creates a thesis (201, status submitted)", async () => {
    const res = await auth(await mhsTok(), {
      path: "/api/v1/theses",
      method: "POST",
      headers: { "content-type": "application/json" },
      body: JSON.stringify({ title: longTitle, abstract: longAbstract, thesisType: "skripsi" }),
    });
    expect(res.status).toBe(201);
    const body = (await res.json()) as any;
    thesisId = body.id;
    expect(body.status).toBe("submitted");
    expect(body.student.id).toBe(studentId);
  });

  it("kaprodi approves the thesis (status transition)", async () => {
    const res = await auth(await kaprodiTok(), {
      path: `/api/v1/theses/${thesisId}`,
      method: "PUT",
      headers: { "content-type": "application/json" },
      body: JSON.stringify({ decision: "approved", notes: "Layak" }),
    });
    expect(res.status).toBe(200);
    const body = (await res.json()) as any;
    expect(body.status).toBe("approved");
    expect(body.approvedAt).not.toBeNull();
  });

  it("another student is forbidden from viewing (403)", async () => {
    const res = await auth(await otherTok(), { path: `/api/v1/theses/${thesisId}` });
    expect(res.status).toBe(403);
  });

  it("returns 404 for a non-existent thesis", async () => {
    const res = await auth(await kaprodiTok(), { path: `/api/v1/theses/${crypto.randomUUID()}` });
    expect(res.status).toBe(404);
  });

  it("kaprodi assigns a supervisor (200, in_progress, supervisor present)", async () => {
    const res = await auth(await kaprodiTok(), {
      path: `/api/v1/theses/${thesisId}/supervisors`,
      method: "POST",
      headers: { "content-type": "application/json" },
      body: JSON.stringify({ supervisorIds: [supervisorId] }),
    });
    expect(res.status).toBe(200);
    const body = (await res.json()) as any;
    expect(body.status).toBe("in_progress");
    expect(body.supervisors.map((s: any) => s.id)).toContain(supervisorId);
  });

  it("mahasiswa sees only their own thesis in list", async () => {
    const res = await auth(await mhsTok(), { path: "/api/v1/theses" });
    expect(res.status).toBe(200);
    const body = (await res.json()) as any;
    expect(body.data.length).toBeGreaterThanOrEqual(1);
    expect(body.data.every((t: any) => t.student.id === studentId)).toBe(true);
  });
});
