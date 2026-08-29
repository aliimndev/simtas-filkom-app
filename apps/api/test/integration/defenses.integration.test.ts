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
let assignedExaminerId = "";
let unrelatedExaminerId = "";
let kaprodiId = "";
let adminId = "";
let academicYearId = "";
let readyThesisId = "";
let noDocumentThesisId = "";
let wrongStateThesisId = "";
let failedThesisId = "";
let cancelledThesisId = "";
let activeDefenseId = "";
let failedDefenseId = "";
let cancelledDefenseId = "";
let raceThesisId = "";
const userIds: string[] = [];
const thesisIds: string[] = [];
const defenseIds: string[] = [];

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

async function thesis(student: string, status: string, withDefenseDocument = true) {
  const [row] = await db.insert(schema.theses).values({
    studentId: student,
    academicYearId,
    title: `Pengujian Sidang ${crypto.randomUUID()}`,
    abstract: "Abstrak pengujian lifecycle Sidang.",
    thesisType: "skripsi",
    status,
  } as any).returning({ id: schema.theses.id });
  thesisIds.push(row.id);
  await db.insert(schema.thesisSupervisors).values({ thesisId: row.id, supervisorId, assignedBy: kaprodiId } as any);
  if (withDefenseDocument) {
    await db.insert(schema.documents).values({
      thesisId: row.id,
      uploadedBy: student,
      documentType: "defense_doc",
      version: 1,
      fileName: "defense.pdf",
      fileUrl: "https://storage.example/defense.pdf",
      status: "approved",
      reviewerId: supervisorId,
      reviewedAt: new Date(),
    } as any);
  }
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

  [studentId, otherStudentId, supervisorId, assignedExaminerId, unrelatedExaminerId, kaprodiId, adminId] = await Promise.all([
    user("defense-student", "Defense Student", mahasiswa),
    user("defense-other", "Other Student", mahasiswa),
    user("defense-supervisor", "Defense Supervisor", pembimbing),
    user("defense-examiner", "Defense Examiner", penguji),
    user("defense-unrelated-examiner", "Unrelated Examiner", penguji),
    user("defense-kaprodi", "Defense Kaprodi", kaprodi),
    user("defense-admin", "Defense Admin", admin),
  ]);

  readyThesisId = await thesis(studentId, "defense_ready");
  const [active] = await db.insert(schema.thesisDefenses).values({ thesisId: readyThesisId, status: "pending" } as any).returning({ id: schema.thesisDefenses.id });
  activeDefenseId = active.id;
  defenseIds.push(activeDefenseId);
  await db.insert(schema.defenseExaminers).values({ defenseId: activeDefenseId, examinerId: assignedExaminerId, assignedBy: kaprodiId } as any);

  failedThesisId = await thesis(studentId, "defense_ready");
  const [failed] = await db.insert(schema.thesisDefenses).values({ thesisId: failedThesisId, status: "failed", finalScore: "45" } as any).returning({ id: schema.thesisDefenses.id });
  failedDefenseId = failed.id;
  defenseIds.push(failedDefenseId);

  cancelledThesisId = await thesis(studentId, "defense_ready");
  const [cancelled] = await db.insert(schema.thesisDefenses).values({ thesisId: cancelledThesisId, status: "cancelled" } as any).returning({ id: schema.thesisDefenses.id });
  cancelledDefenseId = cancelled.id;
  defenseIds.push(cancelledDefenseId);

  noDocumentThesisId = await thesis(studentId, "defense_ready", false);
  wrongStateThesisId = await thesis(studentId, "seminar_ready");
  raceThesisId = await thesis(studentId, "defense_ready");
});

afterAll(async () => {
  if (defenseIds.length > 0) {
    await db.delete(schema.defenseScores).where(inArray(schema.defenseScores.defenseId, defenseIds as any));
    await db.delete(schema.defenseExaminers).where(inArray(schema.defenseExaminers.defenseId, defenseIds as any));
    await db.delete(schema.auditLogs).where(and(eq(schema.auditLogs.entityType, "defense"), inArray(schema.auditLogs.entityId, defenseIds as any)));
    await db.delete(schema.thesisDefenses).where(inArray(schema.thesisDefenses.id, defenseIds as any));
  }
  if (thesisIds.length > 0) {
    await db.delete(schema.documents).where(inArray(schema.documents.thesisId, thesisIds as any));
    await db.delete(schema.thesisSupervisors).where(inArray(schema.thesisSupervisors.thesisId, thesisIds as any));
    await db.delete(schema.theses).where(inArray(schema.theses.id, thesisIds as any));
  }
  await db.delete(schema.notifications).where(inArray(schema.notifications.userId, userIds as any));
  await db.delete(schema.users).where(inArray(schema.users.id, userIds as any));
  await db.delete(schema.academicYears).where(eq(schema.academicYears.id, academicYearId));
});

describe("Sidang submission and attempt lifecycle", () => {
  it("requires an owned defense-ready Thesis and the latest approved defense document", async () => {
    const forbidden = await request(`/api/v1/theses/${readyThesisId}/defenses`, {
      method: "POST",
      headers: { "content-type": "application/json", ...(await auth(otherStudentId, "MAHASISWA")) },
    });
    expect(forbidden.status).toBe(403);

    const noDocument = await request(`/api/v1/theses/${noDocumentThesisId}/defenses`, {
      method: "POST",
      headers: { "content-type": "application/json", ...(await auth(studentId, "MAHASISWA")) },
    });
    expect(noDocument.status).toBe(422);
    expect((await noDocument.json() as any).error.code).toBe("GATE_NOT_MET");

    const wrongState = await request(`/api/v1/theses/${wrongStateThesisId}/defenses`, {
      method: "POST",
      headers: { "content-type": "application/json", ...(await auth(studentId, "MAHASISWA")) },
    });
    expect(wrongState.status).toBe(422);
  });

  it("reads role-scoped defense history and rejects unrelated users", async () => {
    const owner = await request(`/api/v1/defenses/${activeDefenseId}`, { headers: await auth(studentId, "MAHASISWA") });
    expect(owner.status).toBe(200);
    expect((await owner.json() as any).data.status).toBe("pending");

    const supervisor = await request(`/api/v1/defenses/${activeDefenseId}`, { headers: await auth(supervisorId, "DOSEN_PEMBIMBING") });
    expect(supervisor.status).toBe(200);

    const examiner = await request(`/api/v1/defenses/${activeDefenseId}`, { headers: await auth(assignedExaminerId, "DOSEN_PENGUJI") });
    expect(examiner.status).toBe(200);

    const kaprodi = await request(`/api/v1/defenses/${activeDefenseId}`, { headers: await auth(kaprodiId, "KAPRODI") });
    expect(kaprodi.status).toBe(200);

    const admin = await request(`/api/v1/defenses/${activeDefenseId}`, { headers: await auth(adminId, "ADMIN_FAKULTAS") });
    expect(admin.status).toBe(200);

    const unrelated = await request(`/api/v1/defenses/${activeDefenseId}`, { headers: await auth(unrelatedExaminerId, "DOSEN_PENGUJI") });
    expect(unrelated.status).toBe(403);

    const list = await request("/api/v1/defenses", { headers: await auth(studentId, "MAHASISWA") });
    expect(list.status).toBe(200);
    expect((await list.json() as any).data.some((row: any) => row.id === activeDefenseId)).toBe(true);

    const cancelledHistory = await request(`/api/v1/defenses/${cancelledDefenseId}`, { headers: await auth(studentId, "MAHASISWA") });
    expect(cancelledHistory.status).toBe(200);
    expect((await cancelledHistory.json() as any).data.status).toBe("cancelled");
  });

  it("allows a failed retry, preserves history, audits, and notifies Kaprodi", async () => {
    const response = await request(`/api/v1/theses/${failedThesisId}/defenses`, {
      method: "POST",
      headers: { "content-type": "application/json", ...(await auth(studentId, "MAHASISWA")) },
    });
    expect(response.status).toBe(201);
    const body = await response.json() as any;
    expect(body.data.status).toBe("pending");
    expect(body.data.thesisStatus).toBe("defense_ready");
    defenseIds.push(body.data.id);

    const list = await request(`/api/v1/defenses?thesis_id=${failedThesisId}`, { headers: await auth(studentId, "MAHASISWA") });
    expect((await list.json() as any).data.map((row: any) => row.status)).toEqual(expect.arrayContaining(["failed", "pending"]));

    const audits = await db.select().from(schema.auditLogs).where(eq(schema.auditLogs.entityId, body.data.id));
    expect(audits.some((row) => row.action === "defense_submitted")).toBe(true);
    const notifications = await db.select().from(schema.notifications).where(eq(schema.notifications.userId, kaprodiId));
    expect(notifications.some((row) => row.title === "Pengajuan Sidang Baru")).toBe(true);
  });

  it("rejects a second active attempt and protects concurrent submission", async () => {
    const duplicate = await request(`/api/v1/theses/${readyThesisId}/defenses`, {
      method: "POST",
      headers: { "content-type": "application/json", ...(await auth(studentId, "MAHASISWA")) },
    });
    expect(duplicate.status).toBe(409);

    const responses = await Promise.all([
      request(`/api/v1/theses/${raceThesisId}/defenses`, { method: "POST", headers: { "content-type": "application/json", ...(await auth(studentId, "MAHASISWA")) } }),
      request(`/api/v1/theses/${raceThesisId}/defenses`, { method: "POST", headers: { "content-type": "application/json", ...(await auth(studentId, "MAHASISWA")) } }),
    ]);
    expect(responses.filter((response) => response.status === 201)).toHaveLength(1);
    expect(responses.filter((response) => response.status === 409)).toHaveLength(1);
    for (const response of responses) {
      if (response.status === 201) defenseIds.push((await response.json() as any).data.id);
    }
  });
});
