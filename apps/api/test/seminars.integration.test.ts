import { afterAll, beforeAll, describe, expect, it } from "bun:test";
import { and, eq } from "drizzle-orm";
import { createApp } from "../src/app";
import { loadConfig } from "../src/config";
import { getDb } from "../src/db";
import { signAccessToken } from "../src/services/token";
import { schema } from "@sims/db";
import { seminarSubmissionRoutes, seminarsRoutes } from "../src/routes/seminars";

const TEST_DB_URL = process.env.DATABASE_URL ?? "postgres://postgres@localhost:5433/simtas";
const cfg = loadConfig({ ...process.env, NODE_ENV: "test", DATABASE_URL: TEST_DB_URL } as any);
const app = createApp(cfg);
app.route("/api/v1/seminars", seminarsRoutes);
app.route("/api/v1/theses", seminarSubmissionRoutes);
const db = getDb(cfg.databaseUrl);

const DUMMY_HASH = "$2a$12$qMPO1EgF0zmpDh4W49ERVOfOxF28jsItEaiKEKZCWOL9NoKX3U7iC";
const unique = (prefix: string) => `${prefix}-${crypto.randomUUID()}@filkom.ac.id`;

let studentId = "";
let otherStudentId = "";
let supervisorId = "";
let examinerId = "";
let kaprodiId = "";
let academicYearId = "";
let thesisId = "";
let seminarId = "";
let missingDocumentThesisId = "";

async function roleId(name: string): Promise<number> {
  const rows = await db.select().from(schema.roles).where(eq(schema.roles.name, name));
  if (rows[0]) return rows[0].id;
  const [created] = await db.insert(schema.roles).values({ name }).returning({ id: schema.roles.id });
  return created.id;
}

async function user(email: string, name: string, roleId: number): Promise<string> {
  const [row] = await db
    .insert(schema.users)
    .values({
      email,
      fullName: name,
      roleId,
      passwordHash: DUMMY_HASH,
      isActive: true,
      mustChangePassword: false,
      tokenVersion: 0,
    } as any)
    .returning({ id: schema.users.id });
  return row.id;
}

const token = (id: string, role: string) => signAccessToken(id, role, 0);
const auth = async (id: string, role: string) => ({ authorization: `Bearer ${await token(id, role)}` });
const request = (path: string, init: Record<string, any> = {}) =>
  app.request(path, {
    ...init,
    headers: { "x-forwarded-for": "127.0.0.1", ...(init.headers ?? {}) },
  });

beforeAll(async () => {
  const [mahasiswa, dosen, penguji, kaprodi] = await Promise.all([
    roleId("mahasiswa"),
    roleId("dosen_pembimbing"),
    roleId("dosen_penguji"),
    roleId("kaprodi"),
  ]);
  const [year] = await db
    .insert(schema.academicYears)
    .values({ name: `2026/${crypto.randomUUID().slice(0, 4)}`, semester: "ganjil", startDate: "2026-08-01", endDate: "2026-12-31", isActive: false } as any)
    .returning({ id: schema.academicYears.id });
  academicYearId = year.id;

  [studentId, otherStudentId, supervisorId, examinerId, kaprodiId] = await Promise.all([
    user(unique("seminar-student"), "Seminar Student", mahasiswa),
    user(unique("seminar-other"), "Other Student", mahasiswa),
    user(unique("seminar-supervisor"), "Seminar Supervisor", dosen),
    user(unique("seminar-examiner"), "Seminar Examiner", penguji),
    user(unique("seminar-kaprodi"), "Seminar Kaprodi", kaprodi),
  ]);

  const [thesis] = await db
    .insert(schema.theses)
    .values({
      studentId,
      academicYearId,
      title: "Analisis Sistem Informasi Akademik untuk Pengelolaan Tugas Akhir Mahasiswa",
      abstract: "Abstrak penelitian sistem informasi akademik untuk pengelolaan tugas akhir mahasiswa.",
      thesisType: "skripsi",
      status: "in_progress",
    } as any)
    .returning({ id: schema.theses.id });
  thesisId = thesis.id;

  await db.insert(schema.thesisSupervisors).values({ thesisId, supervisorId, assignedBy: kaprodiId } as any);
  await db.insert(schema.documents).values({
    thesisId,
    uploadedBy: studentId,
    documentType: "seminar_doc",
    version: 1,
    fileName: "seminar.pdf",
    fileUrl: "https://storage.example/seminar.pdf",
    status: "approved",
    reviewerId: supervisorId,
    reviewedAt: new Date(),
  } as any);

  const [missingThesis] = await db
    .insert(schema.theses)
    .values({
      studentId,
      academicYearId,
      title: "Analisis Sistem Informasi Akademik Tanpa Dokumen Seminar Approved",
      abstract: "Abstrak penelitian sistem informasi akademik tanpa dokumen seminar yang disetujui.",
      thesisType: "skripsi",
      status: "in_progress",
    } as any)
    .returning({ id: schema.theses.id });
  missingDocumentThesisId = missingThesis.id;
});

afterAll(async () => {
  await db.delete(schema.auditLogs).where(and(eq(schema.auditLogs.entityType, "seminar"), eq(schema.auditLogs.userId, studentId)));
  await db.delete(schema.notifications).where(eq(schema.notifications.userId, kaprodiId));
  await db.delete(schema.seminars).where(eq(schema.seminars.thesisId, thesisId));
  await db.delete(schema.documents).where(eq(schema.documents.thesisId, thesisId));
  await db.delete(schema.documents).where(eq(schema.documents.thesisId, missingDocumentThesisId));
  await db.delete(schema.theses).where(eq(schema.theses.studentId, studentId));
  await db.delete(schema.thesisSupervisors).where(eq(schema.thesisSupervisors.supervisorId, supervisorId));
  await db.delete(schema.users).where(eq(schema.users.id, studentId));
  await db.delete(schema.users).where(eq(schema.users.id, otherStudentId));
  await db.delete(schema.users).where(eq(schema.users.id, supervisorId));
  await db.delete(schema.users).where(eq(schema.users.id, examinerId));
  await db.delete(schema.users).where(eq(schema.users.id, kaprodiId));
  await db.delete(schema.academicYears).where(eq(schema.academicYears.id, academicYearId));
});

describe("Seminar submission and role-scoped access", () => {
  it("submits a pending Seminar, transitions the thesis, audits, and notifies Kaprodi", async () => {
    const response = await request(`/api/v1/theses/${thesisId}/seminars`, {
      method: "POST",
      headers: { "content-type": "application/json", ...(await auth(studentId, "MAHASISWA")) },
    });
    expect(response.status).toBe(201);
    const body = (await response.json()) as any;
    seminarId = body.data.id;
    expect(body.data.status).toBe("pending");
    expect(body.data.thesis.id).toBe(thesisId);
    expect(body.data.thesis.student.id).toBe(studentId);

    const [thesis] = await db.select().from(schema.theses).where(eq(schema.theses.id, thesisId));
    expect(thesis.status).toBe("seminar_ready");
    const audit = await db.select().from(schema.auditLogs).where(and(eq(schema.auditLogs.entityType, "seminar"), eq(schema.auditLogs.entityId, seminarId)));
    expect(audit.some((row) => row.action === "seminar_submitted")).toBe(true);
    const notifications = await db.select().from(schema.notifications).where(eq(schema.notifications.userId, kaprodiId));
    expect(notifications.some((row) => row.type === "seminar")).toBe(true);
  });

  it("rejects a duplicate active Seminar", async () => {
    const response = await request(`/api/v1/theses/${thesisId}/seminars`, {
      method: "POST",
      headers: { "content-type": "application/json", ...(await auth(studentId, "MAHASISWA")) },
    });
    expect(response.status).toBe(409);
    expect((await response.json() as any).error.code).toBe("CONFLICT");
  });

  it("rejects a Seminar without an approved seminar document", async () => {
    const response = await request(`/api/v1/theses/${missingDocumentThesisId}/seminars`, {
      method: "POST",
      headers: { "content-type": "application/json", ...(await auth(studentId, "MAHASISWA")) },
    });
    expect(response.status).toBe(422);
  });

  it("rejects a different Mahasiswa from submitting", async () => {
    const response = await request(`/api/v1/theses/${thesisId}/seminars`, {
      method: "POST",
      headers: { "content-type": "application/json", ...(await auth(otherStudentId, "MAHASISWA")) },
    });
    expect(response.status).toBe(403);
  });

  it("lists and returns detail for owner, supervisor, and Kaprodi but rejects an unrelated Mahasiswa", async () => {
    const list = await request("/api/v1/seminars?page=1&per_page=10&status=pending", { headers: await auth(studentId, "MAHASISWA") });
    expect(list.status).toBe(200);
    const listBody = (await list.json()) as any;
    expect(listBody.data.some((row: any) => row.id === seminarId)).toBe(true);
    expect(listBody.meta.total).toBeGreaterThanOrEqual(1);

    const detail = await request(`/api/v1/seminars/${seminarId}`, { headers: await auth(supervisorId, "DOSEN_PEMBIMBING") });
    expect(detail.status).toBe(200);
    expect((await detail.json() as any).data.id).toBe(seminarId);

    const kaprodi = await request("/api/v1/seminars", { headers: await auth(kaprodiId, "KAPRODI") });
    expect(kaprodi.status).toBe(200);

    const forbidden = await request(`/api/v1/seminars/${seminarId}`, { headers: await auth(otherStudentId, "MAHASISWA") });
    expect(forbidden.status).toBe(403);
  });
});
