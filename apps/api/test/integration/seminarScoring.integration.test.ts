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

const RUBRIC = [
  { component_name: "Presentasi", component_weight: 30, score: 80, notes: "Penyampaian jelas" },
  { component_name: "Penguasaan Materi", component_weight: 30, score: 70, notes: "Materi dikuasai" },
  { component_name: "Kualitas Naskah", component_weight: 25, score: 90, notes: "Naskah baik" },
  { component_name: "Kemampuan Menjawab", component_weight: 15, score: 80, notes: "Jawaban tepat" },
];

let studentId = "";
let examinerOneId = "";
let examinerTwoId = "";
let supervisorId = "";
let kaprodiId = "";
let academicYearId = "";
let thesisId = "";
let seminarId = "";
const userIds: string[] = [];

async function roleId(name: string): Promise<number> {
  const rows = await db.select().from(schema.roles).where(eq(schema.roles.name, name));
  if (rows[0]) return rows[0].id;
  const [created] = await db.insert(schema.roles).values({ name }).returning({ id: schema.roles.id });
  return created.id;
}

async function user(emailPrefix: string, name: string, roleId: number): Promise<string> {
  const [row] = await db.insert(schema.users).values({
    email: unique(emailPrefix),
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

const auth = async (id: string, role: string) => ({
  authorization: `Bearer ${await signAccessToken(id, role, 0)}`,
});

const request = (path: string, init: Record<string, any> = {}) => app.request(path, {
  ...init,
  headers: { "x-forwarded-for": "127.0.0.1", ...(init.headers ?? {}) },
});

beforeAll(async () => {
  const [mahasiswa, penguji, pembimbing, kaprodi] = await Promise.all([
    roleId("mahasiswa"),
    roleId("dosen_penguji"),
    roleId("dosen_pembimbing"),
    roleId("kaprodi"),
  ]);
  const [year] = await db.insert(schema.academicYears).values({
    name: `2026/${crypto.randomUUID().slice(0, 4)}`,
    semester: "ganjil",
    startDate: "2026-08-01",
    endDate: "2026-12-31",
    isActive: false,
  } as any).returning({ id: schema.academicYears.id });
  academicYearId = year.id;

  [studentId, examinerOneId, examinerTwoId, supervisorId, kaprodiId] = await Promise.all([
    user("scoring-student", "Scoring Student", mahasiswa),
    user("scoring-examiner-one", "Scoring Examiner One", penguji),
    user("scoring-examiner-two", "Scoring Examiner Two", penguji),
    user("scoring-supervisor", "Scoring Supervisor", pembimbing),
    user("scoring-kaprodi", "Scoring Kaprodi", kaprodi),
  ]);

  const [thesis] = await db.insert(schema.theses).values({
    studentId,
    academicYearId,
    title: "Analisis Sistem Informasi Akademik untuk Pengujian Nilai Seminar Tugas Akhir",
    abstract: "Abstrak penelitian untuk pengujian scoring Seminar dalam sistem tugas akhir.",
    thesisType: "skripsi",
    status: "seminar_ready",
  } as any).returning({ id: schema.theses.id });
  thesisId = thesis.id;
  await db.insert(schema.thesisSupervisors).values({ thesisId, supervisorId, assignedBy: kaprodiId } as any);

  const [seminar] = await db.insert(schema.seminars).values({
    thesisId,
    status: "scheduled",
    scheduledAt: new Date(Date.now() + 4 * 24 * 60 * 60 * 1000),
    room: "R-SCORING",
  } as any).returning({ id: schema.seminars.id });
  seminarId = seminar.id;
  await db.insert(schema.seminarExaminers).values([
    { seminarId, examinerId: examinerOneId, assignedBy: kaprodiId },
    { seminarId, examinerId: examinerTwoId, assignedBy: kaprodiId },
  ] as any);
});

afterAll(async () => {
  await db.delete(schema.seminarScores).where(eq(schema.seminarScores.seminarId, seminarId));
  await db.delete(schema.seminarExaminers).where(eq(schema.seminarExaminers.seminarId, seminarId));
  await db.delete(schema.auditLogs).where(and(eq(schema.auditLogs.entityType, "seminar"), eq(schema.auditLogs.entityId, seminarId)));
  await db.delete(schema.notifications).where(inArray(schema.notifications.userId, userIds as any));
  await db.delete(schema.seminars).where(eq(schema.seminars.id, seminarId));
  await db.delete(schema.thesisSupervisors).where(eq(schema.thesisSupervisors.thesisId, thesisId));
  await db.delete(schema.theses).where(eq(schema.theses.id, thesisId));
  await db.delete(schema.users).where(inArray(schema.users.id, userIds as any));
  await db.delete(schema.academicYears).where(eq(schema.academicYears.id, academicYearId));
});

describe("Seminar scoring and result finalization", () => {
  it("allows only an assigned Penguji to submit the complete fixed rubric", async () => {
    const forbidden = await request(`/api/v1/seminars/${seminarId}/scores`, {
      method: "PUT",
      headers: { "content-type": "application/json", ...(await auth(studentId, "MAHASISWA")) },
      body: JSON.stringify({ scores: RUBRIC }),
    });
    expect(forbidden.status).toBe(403);

    const response = await request(`/api/v1/seminars/${seminarId}/scores`, {
      method: "PUT",
      headers: { "content-type": "application/json", ...(await auth(examinerOneId, "DOSEN_PENGUJI")) },
      body: JSON.stringify({ scores: RUBRIC }),
    });
    expect(response.status).toBe(200);
    const body = await response.json() as any;
    expect(body.data.scores).toHaveLength(4);
    expect(body.data.scores.find((score: any) => score.componentName === "Presentasi").notes).toBe("Penyampaian jelas");
  });

  it("rejects duplicate, missing, unknown, changed-weight, and out-of-range rubric values", async () => {
    const cases = [
      RUBRIC.slice(0, 3),
      [RUBRIC[0], RUBRIC[0], RUBRIC[2], RUBRIC[3]],
      [...RUBRIC.slice(0, 3), { ...RUBRIC[3], component_name: "Komponen Tidak Valid" }],
      [...RUBRIC.slice(0, 3), { ...RUBRIC[3], component_weight: 20 }],
      [...RUBRIC.slice(0, 3), { ...RUBRIC[3], score: 101 }],
    ];

    for (const scores of cases) {
      const response = await request(`/api/v1/seminars/${seminarId}/scores`, {
        method: "PUT",
        headers: { "content-type": "application/json", ...(await auth(examinerOneId, "DOSEN_PENGUJI")) },
        body: JSON.stringify({ scores }),
      });
      expect(response.status).toBe(422);
    }
  });

  it("rejects finalization until every assigned Penguji has completed every component", async () => {
    const incomplete = await request(`/api/v1/seminars/${seminarId}/finalize`, {
      method: "POST",
      headers: { "content-type": "application/json", ...(await auth(kaprodiId, "KAPRODI")) },
    });
    expect(incomplete.status).toBe(422);

    const second = await request(`/api/v1/seminars/${seminarId}/scores`, {
      method: "PUT",
      headers: { "content-type": "application/json", ...(await auth(examinerTwoId, "DOSEN_PENGUJI")) },
      body: JSON.stringify({ scores: RUBRIC.map((score) => ({ ...score, score: 60 })) }),
    });
    expect(second.status).toBe(200);

    const finalized = await request(`/api/v1/seminars/${seminarId}/finalize`, {
      method: "POST",
      headers: { "content-type": "application/json", ...(await auth(kaprodiId, "KAPRODI")) },
    });
    expect(finalized.status).toBe(200);
    const result = await finalized.json() as any;
    expect(result.data.status).toBe("passed");
    expect(result.data.finalScore).toBe(69.75);

    const [thesis] = await db.select().from(schema.theses).where(eq(schema.theses.id, thesisId));
    expect(thesis.status).toBe("defense_ready");

    const audits = await db.select().from(schema.auditLogs).where(eq(schema.auditLogs.entityId, seminarId));
    expect(audits.some((row) => row.action === "seminar_scores_updated")).toBe(true);
    expect(audits.some((row) => row.action === "seminar_finalized")).toBe(true);

    const notifications = await db.select().from(schema.notifications).where(inArray(schema.notifications.userId, userIds as any));
    expect(notifications.some((row) => row.userId === studentId && row.title === "Hasil Seminar Lulus")).toBe(true);
    expect(notifications.some((row) => row.userId === supervisorId && row.title === "Hasil Seminar Lulus")).toBe(true);
    expect(notifications.some((row) => row.userId === examinerOneId && row.title === "Hasil Seminar Lulus")).toBe(true);
    expect(notifications.some((row) => row.userId === kaprodiId && row.title === "Hasil Seminar Lulus")).toBe(true);

    const rescored = await request(`/api/v1/seminars/${seminarId}/scores`, {
      method: "PUT",
      headers: { "content-type": "application/json", ...(await auth(examinerOneId, "DOSEN_PENGUJI")) },
      body: JSON.stringify({ scores: RUBRIC }),
    });
    expect(rescored.status).toBe(409);

    const duplicate = await request(`/api/v1/seminars/${seminarId}/finalize`, {
      method: "POST",
      headers: { "content-type": "application/json", ...(await auth(kaprodiId, "KAPRODI")) },
    });
    expect(duplicate.status).toBe(409);
  });
});
