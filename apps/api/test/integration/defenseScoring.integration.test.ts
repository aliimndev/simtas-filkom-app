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
  { component_name: "Penguasaan Materi", component_weight: 30, score: 80, notes: "Materi dikuasai" },
  { component_name: "Kualitas Naskah", component_weight: 25, score: 80, notes: "Naskah baik" },
  { component_name: "Kemampuan Menjawab", component_weight: 15, score: 80, notes: "Jawaban tepat" },
];

let studentId = "";
let otherExaminerId = "";
let examinerOneId = "";
let examinerTwoId = "";
let supervisorId = "";
let kaprodiId = "";
let adminId = "";
let academicYearId = "";
let passDefenseId = "";
let revisionDefenseId = "";
let failedDefenseId = "";
let raceDefenseId = "";
const thesisIds: string[] = [];
const defenseIds: string[] = [];
const userIds: string[] = [];

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

async function scheduledDefense(title: string): Promise<string> {
  const [thesis] = await db.insert(schema.theses).values({
    studentId,
    academicYearId,
    title,
    abstract: "Abstrak pengujian finalisasi Sidang.",
    thesisType: "skripsi",
    status: "defense_ready",
  } as any).returning({ id: schema.theses.id });
  thesisIds.push(thesis.id);
  await db.insert(schema.thesisSupervisors).values({ thesisId: thesis.id, supervisorId, assignedBy: kaprodiId } as any);
  const [defense] = await db.insert(schema.thesisDefenses).values({
    thesisId: thesis.id,
    status: "scheduled",
    scheduledAt: new Date(Date.now() + 8 * 86400000),
    room: `R-${crypto.randomUUID().slice(0, 6)}`,
  } as any).returning({ id: schema.thesisDefenses.id });
  defenseIds.push(defense.id);
  await db.insert(schema.defenseExaminers).values([
    { defenseId: defense.id, examinerId: examinerOneId, assignedBy: kaprodiId },
    { defenseId: defense.id, examinerId: examinerTwoId, assignedBy: kaprodiId },
  ] as any);
  return defense.id;
}

const auth = async (id: string, role: string) => ({
  authorization: `Bearer ${await signAccessToken(id, role, 0)}`,
});
const request = (path: string, init: Record<string, any> = {}) => app.request(path, {
  ...init,
  headers: { "x-forwarded-for": "127.0.0.1", ...(init.headers ?? {}) },
});
const scores = (value: number, notes = "Catatan Penguji") => RUBRIC.map((score) => ({ ...score, score: value, notes }));

beforeAll(async () => {
  const [mahasiswa, penguji, pembimbing, kaprodi, admin] = await Promise.all([
    roleId("mahasiswa"),
    roleId("dosen_penguji"),
    roleId("dosen_pembimbing"),
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

  [studentId, otherExaminerId, examinerOneId, examinerTwoId, supervisorId, kaprodiId, adminId] = await Promise.all([
    user("defense-score-student", "Defense Score Student", mahasiswa),
    user("defense-score-other", "Unrelated Examiner", penguji),
    user("defense-score-one", "Defense Score Examiner One", penguji),
    user("defense-score-two", "Defense Score Examiner Two", penguji),
    user("defense-score-supervisor", "Defense Score Supervisor", pembimbing),
    user("defense-score-kaprodi", "Defense Score Kaprodi", kaprodi),
    user("defense-score-admin", "Defense Score Admin", admin),
  ]);

  passDefenseId = await scheduledDefense("Sidang Threshold Passed");
  revisionDefenseId = await scheduledDefense("Sidang Threshold Revision");
  failedDefenseId = await scheduledDefense("Sidang Threshold Failed");
  raceDefenseId = await scheduledDefense("Sidang Concurrent Finalization");
});

afterAll(async () => {
  if (defenseIds.length) {
    await db.delete(schema.defenseScores).where(inArray(schema.defenseScores.defenseId, defenseIds as any));
    await db.delete(schema.defenseExaminers).where(inArray(schema.defenseExaminers.defenseId, defenseIds as any));
    await db.delete(schema.auditLogs).where(and(eq(schema.auditLogs.entityType, "defense"), inArray(schema.auditLogs.entityId, defenseIds as any)));
    await db.delete(schema.thesisDefenses).where(inArray(schema.thesisDefenses.id, defenseIds as any));
  }
  if (thesisIds.length) {
    await db.delete(schema.thesisSupervisors).where(inArray(schema.thesisSupervisors.thesisId, thesisIds as any));
    await db.delete(schema.theses).where(inArray(schema.theses.id, thesisIds as any));
  }
  await db.delete(schema.notifications).where(inArray(schema.notifications.userId, userIds as any));
  await db.delete(schema.users).where(inArray(schema.users.id, userIds as any));
  await db.delete(schema.academicYears).where(eq(schema.academicYears.id, academicYearId));
});

describe("Sidang scoring and finalization", () => {
  it("allows only an assigned Penguji to submit the fixed rubric and notes", async () => {
    const student = await request(`/api/v1/defenses/${passDefenseId}/scores`, {
      method: "PUT",
      headers: { "content-type": "application/json", ...(await auth(studentId, "MAHASISWA")) },
      body: JSON.stringify({ scores: RUBRIC }),
    });
    expect(student.status).toBe(403);

    const unrelated = await request(`/api/v1/defenses/${passDefenseId}/scores`, {
      method: "PUT",
      headers: { "content-type": "application/json", ...(await auth(otherExaminerId, "DOSEN_PENGUJI")) },
      body: JSON.stringify({ scores: RUBRIC }),
    });
    expect(unrelated.status).toBe(403);

    const response = await request(`/api/v1/defenses/${passDefenseId}/scores`, {
      method: "PUT",
      headers: { "content-type": "application/json", ...(await auth(examinerOneId, "DOSEN_PENGUJI")) },
      body: JSON.stringify({ scores: RUBRIC }),
    });
    expect(response.status).toBe(200);
    const body = await response.json() as any;
    expect(body.data.scores).toHaveLength(4);
    expect(body.data.scores.find((score: any) => score.componentName === "Presentasi").notes).toBe("Penyampaian jelas");
  });

  it("rejects incomplete, duplicate, unknown, changed-weight, and out-of-range rubric values", async () => {
    const cases = [
      RUBRIC.slice(0, 3),
      [RUBRIC[0], RUBRIC[0], RUBRIC[2], RUBRIC[3]],
      [...RUBRIC.slice(0, 3), { ...RUBRIC[3], component_name: "Komponen Tidak Valid" }],
      [...RUBRIC.slice(0, 3), { ...RUBRIC[3], component_weight: 20 }],
      [...RUBRIC.slice(0, 3), { ...RUBRIC[3], score: 101 }],
    ];
    for (const payload of cases) {
      const response = await request(`/api/v1/defenses/${passDefenseId}/scores`, {
        method: "PUT",
        headers: { "content-type": "application/json", ...(await auth(examinerOneId, "DOSEN_PENGUJI")) },
        body: JSON.stringify({ scores: payload }),
      });
      expect(response.status).toBe(422);
    }
  });

  it("requires every assigned Penguji before finalization and applies the passed threshold", async () => {
    const incomplete = await request(`/api/v1/defenses/${passDefenseId}/finalize`, {
      method: "POST",
      headers: { "content-type": "application/json", ...(await auth(kaprodiId, "KAPRODI")) },
      body: JSON.stringify({}),
    });
    expect(incomplete.status).toBe(422);

    const second = await request(`/api/v1/defenses/${passDefenseId}/scores`, {
      method: "PUT",
      headers: { "content-type": "application/json", ...(await auth(examinerTwoId, "DOSEN_PENGUJI")) },
      body: JSON.stringify({ scores: scores(70, "Nilai Penguji kedua") }),
    });
    expect(second.status).toBe(200);

    const finalized = await request(`/api/v1/defenses/${passDefenseId}/finalize`, {
      method: "POST",
      headers: { "content-type": "application/json", ...(await auth(kaprodiId, "KAPRODI")) },
      body: JSON.stringify({}),
    });
    expect(finalized.status).toBe(200);
    const result = await finalized.json() as any;
    expect(result.data.status).toBe("passed");
    expect(result.data.finalScore).toBe(75);

    const [thesis] = await db.select().from(schema.theses).where(eq(schema.theses.id, thesisIds[0]));
    expect(thesis.status).toBe("defense_done");
    const audits = await db.select().from(schema.auditLogs).where(eq(schema.auditLogs.entityId, passDefenseId));
    expect(audits.some((row) => row.action === "defense_scores_updated")).toBe(true);
    expect(audits.some((row) => row.action === "defense_finalized")).toBe(true);
    const notifications = await db.select().from(schema.notifications).where(inArray(schema.notifications.userId, userIds as any));
    expect(notifications.some((row) => row.userId === studentId && row.title === "Hasil Sidang Lulus")).toBe(true);
    expect(notifications.some((row) => row.userId === supervisorId && row.title === "Hasil Sidang Lulus")).toBe(true);
    expect(notifications.some((row) => row.userId === examinerOneId && row.title === "Hasil Sidang Lulus")).toBe(true);
    expect(notifications.some((row) => row.userId === kaprodiId && row.title === "Hasil Sidang Lulus")).toBe(true);
  });

  it("maps 60-74 to revision_required, stores revision notes, and moves Thesis to defense_done", async () => {
    for (const examinerId of [examinerOneId, examinerTwoId]) {
      const response = await request(`/api/v1/defenses/${revisionDefenseId}/scores`, {
        method: "PUT",
        headers: { "content-type": "application/json", ...(await auth(examinerId, "DOSEN_PENGUJI")) },
        body: JSON.stringify({ scores: scores(70) }),
      });
      expect(response.status).toBe(200);
    }
    const finalized = await request(`/api/v1/defenses/${revisionDefenseId}/finalize`, {
      method: "POST",
      headers: { "content-type": "application/json", ...(await auth(adminId, "ADMIN_FAKULTAS")) },
      body: JSON.stringify({ revision_notes: "Perbaiki pembahasan dan kesimpulan." }),
    });
    expect(finalized.status).toBe(200);
    const result = await finalized.json() as any;
    expect(result.data.status).toBe("revision_required");
    expect(result.data.finalScore).toBe(70);
    expect(result.data.revisionNotes).toBe("Perbaiki pembahasan dan kesimpulan.");
    const [thesis] = await db.select().from(schema.theses).where(eq(schema.theses.id, thesisIds[1]));
    expect(thesis.status).toBe("defense_done");
  });

  it("maps below 60 to failed and keeps Thesis defense_ready", async () => {
    for (const examinerId of [examinerOneId, examinerTwoId]) {
      const response = await request(`/api/v1/defenses/${failedDefenseId}/scores`, {
        method: "PUT",
        headers: { "content-type": "application/json", ...(await auth(examinerId, "DOSEN_PENGUJI")) },
        body: JSON.stringify({ scores: scores(50) }),
      });
      expect(response.status).toBe(200);
    }
    const finalized = await request(`/api/v1/defenses/${failedDefenseId}/finalize`, {
      method: "POST",
      headers: { "content-type": "application/json", ...(await auth(kaprodiId, "KAPRODI")) },
      body: JSON.stringify({}),
    });
    expect(finalized.status).toBe(200);
    expect((await finalized.json() as any).data.status).toBe("failed");
    const [thesis] = await db.select().from(schema.theses).where(eq(schema.theses.id, thesisIds[2]));
    expect(thesis.status).toBe("defense_ready");
  });

  it("allows only one concurrent finalization", async () => {
    for (const examinerId of [examinerOneId, examinerTwoId]) {
      const response = await request(`/api/v1/defenses/${raceDefenseId}/scores`, {
        method: "PUT",
        headers: { "content-type": "application/json", ...(await auth(examinerId, "DOSEN_PENGUJI")) },
        body: JSON.stringify({ scores: scores(80) }),
      });
      expect(response.status).toBe(200);
    }

    const responses = await Promise.all([
      request(`/api/v1/defenses/${raceDefenseId}/finalize`, {
        method: "POST",
        headers: { "content-type": "application/json", ...(await auth(kaprodiId, "KAPRODI")) },
        body: JSON.stringify({}),
      }),
      request(`/api/v1/defenses/${raceDefenseId}/finalize`, {
        method: "POST",
        headers: { "content-type": "application/json", ...(await auth(adminId, "ADMIN_FAKULTAS")) },
        body: JSON.stringify({}),
      }),
    ]);
    expect(responses.filter((response) => response.status === 200)).toHaveLength(1);
    expect(responses.filter((response) => response.status === 409)).toHaveLength(1);
  });

  it("makes finalized scores, outcome, assignment, and finalization immutable", async () => {
    const rescore = await request(`/api/v1/defenses/${passDefenseId}/scores`, {
      method: "PUT",
      headers: { "content-type": "application/json", ...(await auth(examinerOneId, "DOSEN_PENGUJI")) },
      body: JSON.stringify({ scores: RUBRIC }),
    });
    expect(rescore.status).toBe(409);

    const duplicate = await request(`/api/v1/defenses/${passDefenseId}/finalize`, {
      method: "POST",
      headers: { "content-type": "application/json", ...(await auth(kaprodiId, "KAPRODI")) },
      body: JSON.stringify({}),
    });
    expect(duplicate.status).toBe(409);
  });
});
