import { afterAll, beforeAll, describe, expect, it } from "bun:test";
import { and, eq, inArray } from "drizzle-orm";
import { createApp } from "../../src/app";
import { loadConfig } from "../../src/config";
import { getDb } from "../../src/db";
import { signAccessToken } from "../../src/modules/auth";
import { schema } from "@sims/db";

const TEST_DB_URL = process.env.DATABASE_URL ?? "postgres://postgres@localhost:5433/simtas";
const cfg = loadConfig({ ...process.env, NODE_ENV: "test", DATABASE_URL: TEST_DB_URL } as any);
const app = createApp(cfg);
const db = getDb(cfg.databaseUrl);
const DUMMY_HASH = "$2a$12$qMPO1EgF0zmpDh4W49ERVOfOxF28jsItEaiKEKZCWOL9NoKX3U7iC";
const unique = (prefix: string) => `${prefix}-${crypto.randomUUID()}@filkom.ac.id`;

let studentId = "";
let otherStudentId = "";
let supervisorId = "";
let examinerOneId = "";
let examinerTwoId = "";
let kaprodiId = "";
let adminId = "";
let academicYearId = "";
let thesisId = "";
let scheduledThesisId = "";
let pendingSeminarId = "";
let scheduledSeminarId = "";
let failedThesisId = "";
let noDocumentThesisId = "";
let passedThesisId = "";
let raceThesisId = "";
const userIds: string[] = [];
const seminarIds: string[] = [];
const thesisIds: string[] = [];

async function roleId(name: string): Promise<number> {
  const rows = await db.select().from(schema.roles).where(eq(schema.roles.name, name));
  if (rows[0]) return rows[0].id;
  const [created] = await db.insert(schema.roles).values({ name }).returning({ id: schema.roles.id });
  return created.id;
}

async function user(prefix: string, name: string, roleId: number): Promise<string> {
  const [row] = await db.insert(schema.users).values({
    email: unique(prefix),
    fullName: name,
    roleId,
    passwordHash: DUMMY_HASH,
    isActive: true,
    mustChangePassword: false,
    tokenVersion: 0,
  } as any).returning({ id: schema.users.id });
  userIds.push(row.id);
  return row.id;
}

async function createThesis(student: string, status = "in_progress") {
  const [thesis] = await db.insert(schema.theses).values({
    studentId: student,
    academicYearId,
    title: `Pengujian pembatalan Seminar ${crypto.randomUUID()}`,
    abstract: "Abstrak pengujian lifecycle Seminar.",
    thesisType: "skripsi",
    status,
  } as any).returning({ id: schema.theses.id });
  thesisIds.push(thesis.id);
  await db.insert(schema.documents).values({
    thesisId: thesis.id,
    uploadedBy: student,
    documentType: "seminar_doc",
    version: 1,
    fileName: "seminar.pdf",
    fileUrl: "https://storage.example/seminar.pdf",
    status: "approved",
    reviewerId: supervisorId,
    reviewedAt: new Date(),
  } as any);
  await db.insert(schema.thesisSupervisors).values({ thesisId: thesis.id, supervisorId, assignedBy: kaprodiId } as any);
  return thesis.id;
}

const auth = async (id: string, role: string) => ({
  authorization: `Bearer ${await signAccessToken(id, role, 0)}`,
});

const request = (path: string, init: Record<string, any> = {}) => app.request(path, {
  ...init,
  headers: { "x-forwarded-for": "127.0.0.1", ...(init.headers ?? {}) },
});

beforeAll(async () => {
  const [mahasiswa, pembimbing, penguji, kaprodi, admin] = await Promise.all([
    roleId("mahasiswa"),
    roleId("dosen_pembimbing"),
    roleId("dosen_penguji"),
    roleId("kaprodi"),
    roleId("admin_fakultas"),
  ]);
  const [year] = await db.insert(schema.academicYears).values({
    name: `2026/${crypto.randomUUID().slice(0, 4)}`,
    semester: "ganjil",
    startDate: "2026-08-01",
    endDate: "2026-12-31",
    isActive: false,
  } as any).returning({ id: schema.academicYears.id });
  academicYearId = year.id;

  [studentId, otherStudentId, supervisorId, examinerOneId, examinerTwoId, kaprodiId, adminId] = await Promise.all([
    user("cancel-student", "Cancellation Student", mahasiswa),
    user("cancel-other", "Other Student", mahasiswa),
    user("cancel-supervisor", "Cancellation Supervisor", pembimbing),
    user("cancel-examiner-one", "Cancellation Examiner One", penguji),
    user("cancel-examiner-two", "Cancellation Examiner Two", penguji),
    user("cancel-kaprodi", "Cancellation Kaprodi", kaprodi),
    user("cancel-admin", "Cancellation Admin", admin),
  ]);

  thesisId = await createThesis(studentId);
  const [pending] = await db.insert(schema.seminars).values({ thesisId, status: "pending" } as any).returning({ id: schema.seminars.id });
  pendingSeminarId = pending.id;
  seminarIds.push(pendingSeminarId);

  scheduledThesisId = await createThesis(studentId);
  const [scheduled] = await db.insert(schema.seminars).values({
    thesisId: scheduledThesisId,
    status: "scheduled",
    scheduledAt: new Date(Date.now() + 7 * 24 * 60 * 60 * 1000),
    room: "R-CANCEL",
  } as any).returning({ id: schema.seminars.id });
  scheduledSeminarId = scheduled.id;
  seminarIds.push(scheduledSeminarId);
  await db.insert(schema.seminarExaminers).values([
    { seminarId: scheduledSeminarId, examinerId: examinerOneId, assignedBy: kaprodiId },
    { seminarId: scheduledSeminarId, examinerId: examinerTwoId, assignedBy: kaprodiId },
  ] as any);

  failedThesisId = await createThesis(studentId, "seminar_ready");
  const [failed] = await db.insert(schema.seminars).values({ thesisId: failedThesisId, status: "failed", finalScore: "40" } as any).returning({ id: schema.seminars.id });
  seminarIds.push(failed.id);

  noDocumentThesisId = await createThesis(studentId);
  await db.delete(schema.documents).where(eq(schema.documents.thesisId, noDocumentThesisId));

  passedThesisId = await createThesis(studentId, "seminar_ready");
  const [passed] = await db.insert(schema.seminars).values({ thesisId: passedThesisId, status: "passed", finalScore: "80" } as any).returning({ id: schema.seminars.id });
  seminarIds.push(passed.id);

  raceThesisId = await createThesis(studentId);
});

afterAll(async () => {
  if (seminarIds.length > 0) {
    await db.delete(schema.seminarScores).where(inArray(schema.seminarScores.seminarId, seminarIds as any));
    await db.delete(schema.seminarExaminers).where(inArray(schema.seminarExaminers.seminarId, seminarIds as any));
    await db.delete(schema.auditLogs).where(and(eq(schema.auditLogs.entityType, "seminar"), inArray(schema.auditLogs.entityId, seminarIds as any)));
    await db.delete(schema.notifications).where(inArray(schema.notifications.userId, userIds as any));
    await db.delete(schema.seminars).where(inArray(schema.seminars.id, seminarIds as any));
  }
  if (thesisIds.length > 0) {
    await db.delete(schema.documents).where(inArray(schema.documents.thesisId, thesisIds as any));
    await db.delete(schema.thesisSupervisors).where(inArray(schema.thesisSupervisors.thesisId, thesisIds as any));
    await db.delete(schema.theses).where(inArray(schema.theses.id, thesisIds as any));
  }
  await db.delete(schema.users).where(inArray(schema.users.id, userIds as any));
  await db.delete(schema.academicYears).where(eq(schema.academicYears.id, academicYearId));
});

describe("Seminar cancellation and retry lifecycle", () => {
  it("requires administrative authorization and a non-blank reason", async () => {
    const student = await request(`/api/v1/seminars/${pendingSeminarId}/cancel`, {
      method: "POST",
      headers: { "content-type": "application/json", ...(await auth(studentId, "MAHASISWA")) },
      body: JSON.stringify({ reason: "Tidak boleh" }),
    });
    expect(student.status).toBe(403);

    const blank = await request(`/api/v1/seminars/${pendingSeminarId}/cancel`, {
      method: "POST",
      headers: { "content-type": "application/json", ...(await auth(kaprodiId, "KAPRODI")) },
      body: JSON.stringify({ reason: "  " }),
    });
    expect(blank.status).toBe(422);

    const missing = await request(`/api/v1/seminars/${pendingSeminarId}/cancel`, {
      method: "POST",
      headers: { "content-type": "application/json", ...(await auth(adminId, "ADMIN_FAKULTAS")) },
      body: JSON.stringify({}),
    });
    expect(missing.status).toBe(422);
  });

  it("cancels a pending Seminar transactionally, audits/notifies, and keeps history", async () => {
    const response = await request(`/api/v1/seminars/${pendingSeminarId}/cancel`, {
      method: "POST",
      headers: { "content-type": "application/json", ...(await auth(kaprodiId, "KAPRODI")) },
      body: JSON.stringify({ reason: "Mahasiswa meminta penjadwalan ulang" }),
    });
    expect(response.status).toBe(200);
    const body = await response.json() as any;
    expect(body.data.status).toBe("cancelled");
    expect(body.data.cancellationReason).toBe("Mahasiswa meminta penjadwalan ulang");

    const [thesis] = await db.select().from(schema.theses).where(eq(schema.theses.id, thesisId));
    expect(thesis.status).toBe("seminar_ready");
    const audits = await db.select().from(schema.auditLogs).where(eq(schema.auditLogs.entityId, pendingSeminarId));
    expect(audits.some((row) => row.action === "seminar_cancelled")).toBe(true);
    const notifications = await db.select().from(schema.notifications).where(inArray(schema.notifications.userId, [studentId, supervisorId, kaprodiId] as any));
    expect(notifications.some((row) => row.title === "Seminar Dibatalkan" && row.userId === studentId)).toBe(true);
    expect(notifications.some((row) => row.title === "Seminar Dibatalkan" && row.userId === supervisorId)).toBe(true);

    const history = await request("/api/v1/seminars?status=cancelled", { headers: await auth(studentId, "MAHASISWA") });
    expect(history.status).toBe(200);
    expect((await history.json() as any).data.some((row: any) => row.id === pendingSeminarId)).toBe(true);
  });

  it("keeps cancellation terminal for scheduling, scoring, finalization, and repeat cancellation", async () => {
    const schedule = await request(`/api/v1/seminars/${pendingSeminarId}/schedule`, {
      method: "PUT",
      headers: { "content-type": "application/json", ...(await auth(kaprodiId, "KAPRODI")) },
      body: JSON.stringify({ scheduled_at: new Date(Date.now() + 8 * 86400000).toISOString(), room: "R-NEW", examiner_ids: [examinerOneId, examinerTwoId] }),
    });
    expect(schedule.status).toBe(409);

    const score = await request(`/api/v1/seminars/${pendingSeminarId}/scores`, {
      method: "PUT",
      headers: { "content-type": "application/json", ...(await auth(examinerOneId, "DOSEN_PENGUJI")) },
      body: JSON.stringify({ scores: [
        { component_name: "Presentasi", component_weight: 30, score: 80 },
        { component_name: "Penguasaan Materi", component_weight: 30, score: 80 },
        { component_name: "Kualitas Naskah", component_weight: 25, score: 80 },
        { component_name: "Kemampuan Menjawab", component_weight: 15, score: 80 },
      ] }),
    });
    expect(score.status).toBe(409);

    const finalize = await request(`/api/v1/seminars/${pendingSeminarId}/finalize`, {
      method: "POST",
      headers: await auth(kaprodiId, "KAPRODI"),
    });
    expect(finalize.status).toBe(409);

    const repeat = await request(`/api/v1/seminars/${pendingSeminarId}/cancel`, {
      method: "POST",
      headers: { "content-type": "application/json", ...(await auth(adminId, "ADMIN_FAKULTAS")) },
      body: JSON.stringify({ reason: "Alasan kedua" }),
    });
    expect(repeat.status).toBe(409);
  });

  it("cancels a scheduled Seminar and notifies assigned Penguji", async () => {
    const response = await request(`/api/v1/seminars/${scheduledSeminarId}/cancel`, {
      method: "POST",
      headers: { "content-type": "application/json", ...(await auth(adminId, "ADMIN_FAKULTAS")) },
      body: JSON.stringify({ reason: "Perubahan agenda fakultas" }),
    });
    expect(response.status).toBe(200);
    const body = await response.json() as any;
    expect(body.data.status).toBe("cancelled");
    const notifications = await db.select().from(schema.notifications).where(inArray(schema.notifications.userId, [examinerOneId, examinerTwoId] as any));
    expect(notifications.filter((row) => row.title === "Seminar Dibatalkan")).toHaveLength(2);
  });

  it("allows retry after failed or cancelled attempts only with an approved document", async () => {
    const retryCancelled = await request(`/api/v1/theses/${thesisId}/seminars`, {
      method: "POST",
      headers: { "content-type": "application/json", ...(await auth(studentId, "MAHASISWA")) },
    });
    expect(retryCancelled.status).toBe(201);
    const retry = await retryCancelled.json() as any;
    expect(retry.data.status).toBe("pending");
    seminarIds.push(retry.data.id);

    const history = await request(`/api/v1/seminars?status=cancelled`, { headers: await auth(studentId, "MAHASISWA") });
    expect((await history.json() as any).data.some((row: any) => row.id === pendingSeminarId)).toBe(true);

    const retryFailed = await request(`/api/v1/theses/${failedThesisId}/seminars`, {
      method: "POST",
      headers: { "content-type": "application/json", ...(await auth(studentId, "MAHASISWA")) },
    });
    expect(retryFailed.status).toBe(201);
    seminarIds.push((await retryFailed.json() as any).data.id);

    const noDocument = await request(`/api/v1/theses/${noDocumentThesisId}/seminars`, {
      method: "POST",
      headers: { "content-type": "application/json", ...(await auth(studentId, "MAHASISWA")) },
    });
    expect(noDocument.status).toBe(422);

    const passed = await request(`/api/v1/theses/${passedThesisId}/seminars`, {
      method: "POST",
      headers: { "content-type": "application/json", ...(await auth(studentId, "MAHASISWA")) },
    });
    expect(passed.status).toBe(409);
  });

  it("enforces one pending retry under concurrent submissions", async () => {
    const responses = await Promise.all([
      request(`/api/v1/theses/${raceThesisId}/seminars`, { method: "POST", headers: { "content-type": "application/json", ...(await auth(studentId, "MAHASISWA")) } }),
      request(`/api/v1/theses/${raceThesisId}/seminars`, { method: "POST", headers: { "content-type": "application/json", ...(await auth(studentId, "MAHASISWA")) } }),
    ]);
    expect(responses.filter((response) => response.status === 201)).toHaveLength(1);
    expect(responses.filter((response) => response.status === 409)).toHaveLength(1);
    for (const response of responses) {
      if (response.status === 201) seminarIds.push((await response.json() as any).data.id);
    }
  });
});
