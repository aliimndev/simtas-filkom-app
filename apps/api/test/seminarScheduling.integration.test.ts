import { afterAll, beforeAll, describe, expect, it } from "bun:test";
import { and, eq, inArray } from "drizzle-orm";
import { createApp } from "../src/app";
import { loadConfig } from "../src/config";
import { getDb } from "../src/db";
import { signAccessToken } from "../src/services/token";
import { schema } from "@sims/db";
import { seminarsRoutes } from "../src/modules/seminars";

const TEST_DB_URL = process.env.DATABASE_URL ?? "postgres://postgres@localhost:5433/simtas";
const cfg = loadConfig({ ...process.env, NODE_ENV: "test", DATABASE_URL: TEST_DB_URL } as any);
const app = createApp(cfg);
app.route("/api/v1/seminars", seminarsRoutes);
const db = getDb(cfg.databaseUrl);
const DUMMY_HASH = "$2a$12$qMPO1EgF0zmpDh4W49ERVOfOxF28jsItEaiKEKZCWOL9NoKX3U7iC";
const unique = (prefix: string) => `${prefix}-${crypto.randomUUID()}@filkom.ac.id`;

let studentId = "";
let kaprodiId = "";
let examinerOneId = "";
let examinerTwoId = "";
let examinerThreeId = "";
let supervisorId = "";
let inactiveExaminerId = "";
let academicYearId = "";
let thesisId = "";
let seminarId = "";
const extraThesisIds: string[] = [];
const extraSeminarIds: string[] = [];

async function roleId(name: string): Promise<number> {
  const rows = await db.select().from(schema.roles).where(eq(schema.roles.name, name));
  if (rows[0]) return rows[0].id;
  const [created] = await db.insert(schema.roles).values({ name }).returning({ id: schema.roles.id });
  return created.id;
}

async function user(emailPrefix: string, name: string, roleId: number, isActive = true): Promise<string> {
  const [row] = await db
    .insert(schema.users)
    .values({
      email: unique(emailPrefix),
      fullName: name,
      roleId,
      passwordHash: DUMMY_HASH,
      isActive,
      mustChangePassword: false,
      tokenVersion: 0,
    } as any)
    .returning({ id: schema.users.id });
  return row.id;
}

async function createPendingSeminar(student: string, title: string): Promise<{ thesisId: string; seminarId: string }> {
  const [thesis] = await db
    .insert(schema.theses)
    .values({
      studentId: student,
      academicYearId,
      title,
      abstract: "Abstrak penelitian untuk pengujian penjadwalan Seminar.",
      thesisType: "skripsi",
      status: "seminar_ready",
    } as any)
    .returning({ id: schema.theses.id });
  const [seminar] = await db
    .insert(schema.seminars)
    .values({ thesisId: thesis.id, status: "pending" } as any)
    .returning({ id: schema.seminars.id });
  extraThesisIds.push(thesis.id);
  extraSeminarIds.push(seminar.id);
  return { thesisId: thesis.id, seminarId: seminar.id };
}

const token = (id: string, role: string) => signAccessToken(id, role, 0);
const auth = async (id: string, role: string) => ({ authorization: `Bearer ${await token(id, role)}` });
const request = (path: string, init: Record<string, any> = {}) =>
  app.request(path, { ...init, headers: { "x-forwarded-for": "127.0.0.1", ...(init.headers ?? {}) } });
const scheduleAt = (days: number) => new Date(Date.now() + days * 24 * 60 * 60 * 1000).toISOString();
const scheduleBody = (scheduledAt: string, room: string, examinerIds: string[]) => ({
  scheduled_at: scheduledAt,
  room,
  examiner_ids: examinerIds,
});

beforeAll(async () => {
  const [mahasiswa, kaprodi, penguji, pembimbing] = await Promise.all([
    roleId("mahasiswa"),
    roleId("kaprodi"),
    roleId("dosen_penguji"),
    roleId("dosen_pembimbing"),
  ]);
  const [year] = await db
    .insert(schema.academicYears)
    .values({ name: `2026/${crypto.randomUUID().slice(0, 4)}`, semester: "ganjil", startDate: "2026-08-01", endDate: "2026-12-31", isActive: false } as any)
    .returning({ id: schema.academicYears.id });
  academicYearId = year.id;

  [studentId, kaprodiId, examinerOneId, examinerTwoId, examinerThreeId, supervisorId, inactiveExaminerId] = await Promise.all([
    user("schedule-student", "Schedule Student", mahasiswa),
    user("schedule-kaprodi", "Schedule Kaprodi", kaprodi),
    user("schedule-examiner-one", "Examiner One", penguji),
    user("schedule-examiner-two", "Examiner Two", penguji),
    user("schedule-examiner-three", "Examiner Three", penguji),
    user("schedule-supervisor", "Schedule Supervisor", pembimbing),
    user("schedule-inactive", "Inactive Examiner", penguji, false),
  ]);

  const [thesis] = await db
    .insert(schema.theses)
    .values({
      studentId,
      academicYearId,
      title: "Analisis Penjadwalan Seminar Tugas Akhir Berbasis Sistem Informasi",
      abstract: "Abstrak penelitian untuk pengujian penjadwalan Seminar.",
      thesisType: "skripsi",
      status: "seminar_ready",
    } as any)
    .returning({ id: schema.theses.id });
  thesisId = thesis.id;
  const [seminar] = await db.insert(schema.seminars).values({ thesisId, status: "pending" } as any).returning({ id: schema.seminars.id });
  seminarId = seminar.id;
  await db.insert(schema.thesisSupervisors).values({ thesisId, supervisorId, assignedBy: kaprodiId } as any);
});

afterAll(async () => {
  const seminarIds = [seminarId, ...extraSeminarIds].filter(Boolean);
  const thesisIds = [thesisId, ...extraThesisIds].filter(Boolean);
  if (seminarIds.length) await db.delete(schema.seminarExaminers).where(inArray(schema.seminarExaminers.seminarId, seminarIds as any));
  if (thesisIds.length) await db.delete(schema.thesisSupervisors).where(inArray(schema.thesisSupervisors.thesisId, thesisIds as any));
  if (seminarIds.length) await db.delete(schema.auditLogs).where(and(eq(schema.auditLogs.entityType, "seminar"), inArray(schema.auditLogs.entityId, seminarIds as any)));
  await db.delete(schema.notifications).where(eq(schema.notifications.userId, kaprodiId));
  if (thesisIds.length) await db.delete(schema.seminars).where(inArray(schema.seminars.id, seminarIds as any));
  if (thesisIds.length) await db.delete(schema.theses).where(inArray(schema.theses.id, thesisIds as any));
  await db.delete(schema.users).where(inArray(schema.users.id, [studentId, kaprodiId, examinerOneId, examinerTwoId, examinerThreeId, supervisorId, inactiveExaminerId] as any));
  await db.delete(schema.academicYears).where(eq(schema.academicYears.id, academicYearId));
});

describe("Seminar scheduling and Penguji assignment", () => {
  it("requires Kaprodi/Admin authorization", async () => {
    const response = await request(`/api/v1/seminars/${seminarId}/schedule`, {
      method: "PUT",
      headers: { "content-type": "application/json", ...(await auth(studentId, "MAHASISWA")) },
      body: JSON.stringify(scheduleBody(scheduleAt(4), "R-101", [examinerOneId, examinerTwoId])),
    });
    expect(response.status).toBe(403);
  });

  it("rejects short lead time, duplicate ids, and fewer than two Penguji", async () => {
    const cases = [
      scheduleBody(scheduleAt(1), "R-102", [examinerOneId, examinerTwoId]),
      scheduleBody(scheduleAt(4), "R-103", [examinerOneId, examinerOneId]),
      scheduleBody(scheduleAt(4), "R-104", [examinerOneId]),
    ];
    for (const body of cases) {
      const response = await request(`/api/v1/seminars/${seminarId}/schedule`, {
        method: "PUT",
        headers: { "content-type": "application/json", ...(await auth(kaprodiId, "KAPRODI")) },
        body: JSON.stringify(body),
      });
      expect(response.status).toBe(422);
    }
  });

  it("rejects invalid, inactive, and non-Penguji users", async () => {
    for (const examinerIds of [[crypto.randomUUID(), examinerTwoId], [inactiveExaminerId, examinerTwoId], [supervisorId, examinerTwoId]]) {
      const response = await request(`/api/v1/seminars/${seminarId}/schedule`, {
        method: "PUT",
        headers: { "content-type": "application/json", ...(await auth(kaprodiId, "KAPRODI")) },
        body: JSON.stringify(scheduleBody(scheduleAt(4), `R-${crypto.randomUUID().slice(0, 4)}`, examinerIds)),
      });
      expect(response.status).toBe(422);
    }
  });

  it("schedules atomically, exposes Penguji, audits, and notifies Kaprodi", async () => {
    const body = scheduleBody(scheduleAt(4), "R-201", [examinerOneId, examinerTwoId]);
    const response = await request(`/api/v1/seminars/${seminarId}/schedule`, {
      method: "PUT",
      headers: { "content-type": "application/json", ...(await auth(kaprodiId, "KAPRODI")) },
      body: JSON.stringify(body),
    });
    expect(response.status).toBe(200);
    const result = (await response.json()) as any;
    expect(result.data.status).toBe("scheduled");
    expect(result.data.room).toBe("R-201");
    expect(result.data.examiners.map((examiner: any) => examiner.id).sort()).toEqual([examinerOneId, examinerTwoId].sort());

    const audit = await db.select().from(schema.auditLogs).where(eq(schema.auditLogs.entityId, seminarId));
    expect(audit.some((row) => row.action === "seminar_scheduled")).toBe(true);
    const notifications = await db.select().from(schema.notifications).where(eq(schema.notifications.userId, studentId));
    expect(notifications.some((row) => row.type === "seminar")).toBe(true);

    const examinerDetail = await request(`/api/v1/seminars/${seminarId}`, { headers: await auth(examinerOneId, "DOSEN_PENGUJI") });
    expect(examinerDetail.status).toBe(200);
  });

  it("rejects room and Penguji conflicts without changing the pending Seminar", async () => {
    const roomConflict = await createPendingSeminar(studentId, "Seminar Room Conflict Thesis");
    const scheduledAt = scheduleAt(4);
    const first = await request(`/api/v1/seminars/${roomConflict.seminarId}/schedule`, {
      method: "PUT",
      headers: { "content-type": "application/json", ...(await auth(kaprodiId, "KAPRODI")) },
      body: JSON.stringify(scheduleBody(scheduledAt, "R-CONFLICT", [examinerOneId, examinerThreeId])),
    });
    expect(first.status).toBe(200);

    const second = await createPendingSeminar(studentId, "Seminar Second Conflict Thesis");
    const room = await request(`/api/v1/seminars/${second.seminarId}/schedule`, {
      method: "PUT",
      headers: { "content-type": "application/json", ...(await auth(kaprodiId, "KAPRODI")) },
      body: JSON.stringify(scheduleBody(scheduledAt, "R-CONFLICT", [examinerTwoId, examinerThreeId])),
    });
    expect(room.status).toBe(409);

    const examiner = await request(`/api/v1/seminars/${second.seminarId}/schedule`, {
      method: "PUT",
      headers: { "content-type": "application/json", ...(await auth(kaprodiId, "KAPRODI")) },
      body: JSON.stringify(scheduleBody(scheduledAt, "R-OTHER", [examinerOneId, examinerTwoId])),
    });
    expect(examiner.status).toBe(409);
    const [unchanged] = await db.select().from(schema.seminars).where(eq(schema.seminars.id, second.seminarId));
    expect(unchanged.status).toBe("pending");
    expect(unchanged.scheduledAt).toBeNull();
  });

  it("reschedules an existing Seminar and replaces its Penguji atomically", async () => {
    const response = await request(`/api/v1/seminars/${seminarId}/schedule`, {
      method: "PUT",
      headers: { "content-type": "application/json", ...(await auth(kaprodiId, "KAPRODI")) },
      body: JSON.stringify(scheduleBody(scheduleAt(5), "R-202", [examinerTwoId, examinerThreeId])),
    });
    expect(response.status).toBe(200);
    const result = (await response.json()) as any;
    expect(result.data.status).toBe("scheduled");
    expect(result.data.examiners.map((examiner: any) => examiner.id).sort()).toEqual([examinerTwoId, examinerThreeId].sort());
  });
});
