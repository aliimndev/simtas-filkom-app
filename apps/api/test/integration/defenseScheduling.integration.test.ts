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
let kaprodiId = "";
let adminId = "";
let examinerOneId = "";
let examinerTwoId = "";
let examinerThreeId = "";
let supervisorId = "";
let academicYearId = "";
let defenseId = "";
let noLetterDefenseId = "";
let scheduledDefenseId = "";
let conflictDefenseId = "";
let cancelledLetterDefenseId = "";
const userIds: string[] = [];
const thesisIds: string[] = [];
const defenseIds: string[] = [];
const letterIds: string[] = [];

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

async function defense(title: string): Promise<string> {
  const [thesis] = await db.insert(schema.theses).values({
    studentId,
    academicYearId,
    title,
    abstract: "Abstrak pengujian Penjadwalan Sidang.",
    thesisType: "skripsi",
    status: "defense_ready",
  } as any).returning({ id: schema.theses.id });
  thesisIds.push(thesis.id);
  await db.insert(schema.thesisSupervisors).values({ thesisId: thesis.id, supervisorId, assignedBy: kaprodiId } as any);
  const [row] = await db.insert(schema.thesisDefenses).values({ thesisId: thesis.id, status: "pending" } as any).returning({ id: schema.thesisDefenses.id });
  defenseIds.push(row.id);
  return row.id;
}

const auth = async (id: string, role: string) => ({
  authorization: `Bearer ${await signAccessToken(id, role, 0)}`,
});
const request = (path: string, init: Record<string, any> = {}) => app.request(path, {
  ...init,
  headers: { "x-forwarded-for": "127.0.0.1", ...(init.headers ?? {}) },
});
const scheduleAt = (days: number) => new Date(Date.now() + days * 86400000).toISOString();
const scheduleBody = (scheduledAt: string, room: string, examinerIds: string[]) => ({
  scheduled_at: scheduledAt,
  room,
  examiner_ids: examinerIds,
});
const letterBody = (number = `ST-${crypto.randomUUID().slice(0, 8)}`) => ({
  letter_number: number,
  issue_date: "2026-09-01",
  file_name: "surat-tugas.pdf",
  file_url: "https://storage.example/surat-tugas.pdf",
});

beforeAll(async () => {
  const [mahasiswa, kaprodi, admin, penguji, pembimbing] = await Promise.all([
    roleId("mahasiswa"),
    roleId("kaprodi"),
    roleId("admin_fakultas"),
    roleId("dosen_penguji"),
    roleId("dosen_pembimbing"),
  ]);
  const [year] = await db.insert(schema.academicYears).values({
    name: `2026/${crypto.randomUUID().slice(0, 4)}`,
    semester: "ganjil",
    startDate: "2026-08-01",
    endDate: "2026-12-31",
    isActive: false,
  } as any).returning({ id: schema.academicYears.id });
  academicYearId = year.id;

  [studentId, kaprodiId, adminId, examinerOneId, examinerTwoId, examinerThreeId, supervisorId] = await Promise.all([
    user("schedule-student", "Scheduling Student", mahasiswa),
    user("schedule-kaprodi", "Scheduling Kaprodi", kaprodi),
    user("schedule-admin", "Scheduling Admin", admin),
    user("schedule-examiner-one", "Scheduling Examiner One", penguji),
    user("schedule-examiner-two", "Scheduling Examiner Two", penguji),
    user("schedule-examiner-three", "Scheduling Examiner Three", penguji),
    user("schedule-supervisor", "Scheduling Supervisor", pembimbing),
  ]);

  defenseId = await defense("Primary Sidang Scheduling");
  noLetterDefenseId = await defense("No Letter Sidang");
  scheduledDefenseId = await defense("Scheduled Sidang");
  conflictDefenseId = await defense("Conflict Sidang");
  cancelledLetterDefenseId = await defense("Cancelled Letter Sidang");
});

afterAll(async () => {
  if (letterIds.length) await db.delete(schema.suratTugas).where(inArray(schema.suratTugas.id, letterIds as any));
  if (defenseIds.length) {
    await db.delete(schema.defenseExaminers).where(inArray(schema.defenseExaminers.defenseId, defenseIds as any));
    await db.delete(schema.auditLogs).where(inArray(schema.auditLogs.entityId, [...defenseIds, ...letterIds] as any));
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

describe("Surat Tugas and Sidang scheduling", () => {
  it("allows only Kaprodi/Admin to create and issue an immutable Surat Tugas", async () => {
    const forbidden = await request(`/api/v1/defenses/${defenseId}/surat-tugas`, {
      method: "POST",
      headers: { "content-type": "application/json", ...(await auth(studentId, "MAHASISWA")) },
      body: JSON.stringify(letterBody()),
    });
    expect(forbidden.status).toBe(403);

    const created = await request(`/api/v1/defenses/${defenseId}/surat-tugas`, {
      method: "POST",
      headers: { "content-type": "application/json", ...(await auth(kaprodiId, "KAPRODI")) },
      body: JSON.stringify(letterBody()),
    });
    expect(created.status).toBe(201);
    const draft = await created.json() as any;
    expect(draft.data.status).toBe("draft");
    letterIds.push(draft.data.id);

    const issued = await request(`/api/v1/surat-tugas/${draft.data.id}/issue`, {
      method: "POST",
      headers: await auth(adminId, "ADMIN_FAKULTAS"),
    });
    expect(issued.status).toBe(200);
    expect((await issued.json() as any).data.status).toBe("issued");

    const duplicate = await request(`/api/v1/defenses/${defenseId}/surat-tugas`, {
      method: "POST",
      headers: { "content-type": "application/json", ...(await auth(kaprodiId, "KAPRODI")) },
      body: JSON.stringify(letterBody("ST-DUPLICATE")),
    });
    expect(duplicate.status).toBe(409);
  });

  it("enforces Surat Tugas lifecycle reason and preserves cancelled history", async () => {
    const created = await request(`/api/v1/defenses/${cancelledLetterDefenseId}/surat-tugas`, {
      method: "POST",
      headers: { "content-type": "application/json", ...(await auth(kaprodiId, "KAPRODI")) },
      body: JSON.stringify(letterBody()),
    });
    const draft = await created.json() as any;
    letterIds.push(draft.data.id);

    const blank = await request(`/api/v1/surat-tugas/${draft.data.id}/cancel`, {
      method: "POST",
      headers: { "content-type": "application/json", ...(await auth(kaprodiId, "KAPRODI")) },
      body: JSON.stringify({ reason: "  " }),
    });
    expect(blank.status).toBe(422);

    const cancelled = await request(`/api/v1/surat-tugas/${draft.data.id}/cancel`, {
      method: "POST",
      headers: { "content-type": "application/json", ...(await auth(kaprodiId, "KAPRODI")) },
      body: JSON.stringify({ reason: "Perubahan penugasan" }),
    });
    expect(cancelled.status).toBe(200);
    expect((await cancelled.json() as any).data.status).toBe("cancelled");

    const reissue = await request(`/api/v1/surat-tugas/${draft.data.id}/issue`, {
      method: "POST",
      headers: await auth(kaprodiId, "KAPRODI"),
    });
    expect(reissue.status).toBe(409);

    const replacement = await request(`/api/v1/defenses/${cancelledLetterDefenseId}/surat-tugas`, {
      method: "POST",
      headers: { "content-type": "application/json", ...(await auth(kaprodiId, "KAPRODI")) },
      body: JSON.stringify(letterBody("ST-REPLACEMENT")),
    });
    expect(replacement.status).toBe(201);
    letterIds.push((await replacement.json() as any).data.id);
  });

  it("requires an issued letter and valid seven-day schedule with two active Penguji", async () => {
    const noLetter = await request(`/api/v1/defenses/${noLetterDefenseId}/schedule`, {
      method: "PUT",
      headers: { "content-type": "application/json", ...(await auth(kaprodiId, "KAPRODI")) },
      body: JSON.stringify(scheduleBody(scheduleAt(8), "R-SIDANG-1", [examinerOneId, examinerTwoId])),
    });
    expect(noLetter.status).toBe(409);

    const letter = await request(`/api/v1/defenses/${noLetterDefenseId}/surat-tugas`, {
      method: "POST",
      headers: { "content-type": "application/json", ...(await auth(kaprodiId, "KAPRODI")) },
      body: JSON.stringify(letterBody()),
    });
    const letterData = await letter.json() as any;
    letterIds.push(letterData.data.id);
    await request(`/api/v1/surat-tugas/${letterData.data.id}/issue`, { method: "POST", headers: await auth(kaprodiId, "KAPRODI") });

    for (const body of [
      scheduleBody(scheduleAt(3), "R-SIDANG-2", [examinerOneId, examinerTwoId]),
      scheduleBody(scheduleAt(8), "R-SIDANG-3", [examinerOneId]),
      scheduleBody(scheduleAt(8), "R-SIDANG-4", [examinerOneId, examinerOneId]),
    ]) {
      const response = await request(`/api/v1/defenses/${noLetterDefenseId}/schedule`, {
        method: "PUT",
        headers: { "content-type": "application/json", ...(await auth(kaprodiId, "KAPRODI")) },
        body: JSON.stringify(body),
      });
      expect(response.status).toBe(422);
    }
  });

  it("schedules atomically, audits, notifies, and rejects room or Penguji conflicts", async () => {
    const letter = await request(`/api/v1/defenses/${scheduledDefenseId}/surat-tugas`, {
      method: "POST",
      headers: { "content-type": "application/json", ...(await auth(kaprodiId, "KAPRODI")) },
      body: JSON.stringify(letterBody()),
    });
    const letterData = await letter.json() as any;
    letterIds.push(letterData.data.id);
    await request(`/api/v1/surat-tugas/${letterData.data.id}/issue`, { method: "POST", headers: await auth(kaprodiId, "KAPRODI") });

    const scheduledAt = scheduleAt(8);
    const scheduled = await request(`/api/v1/defenses/${scheduledDefenseId}/schedule`, {
      method: "PUT",
      headers: { "content-type": "application/json", ...(await auth(adminId, "ADMIN_FAKULTAS")) },
      body: JSON.stringify(scheduleBody(scheduledAt, "R-SIDANG-CONFLICT", [examinerOneId, examinerThreeId])),
    });
    expect(scheduled.status).toBe(200);
    expect((await scheduled.json() as any).data.status).toBe("scheduled");

    const conflictLetter = await request(`/api/v1/defenses/${conflictDefenseId}/surat-tugas`, {
      method: "POST",
      headers: { "content-type": "application/json", ...(await auth(kaprodiId, "KAPRODI")) },
      body: JSON.stringify(letterBody()),
    });
    const conflictLetterData = await conflictLetter.json() as any;
    letterIds.push(conflictLetterData.data.id);
    await request(`/api/v1/surat-tugas/${conflictLetterData.data.id}/issue`, { method: "POST", headers: await auth(kaprodiId, "KAPRODI") });

    const roomConflict = await request(`/api/v1/defenses/${conflictDefenseId}/schedule`, {
      method: "PUT",
      headers: { "content-type": "application/json", ...(await auth(kaprodiId, "KAPRODI")) },
      body: JSON.stringify(scheduleBody(scheduledAt, "R-SIDANG-CONFLICT", [examinerTwoId, examinerThreeId])),
    });
    expect(roomConflict.status).toBe(409);
    const examinerConflict = await request(`/api/v1/defenses/${conflictDefenseId}/schedule`, {
      method: "PUT",
      headers: { "content-type": "application/json", ...(await auth(kaprodiId, "KAPRODI")) },
      body: JSON.stringify(scheduleBody(scheduledAt, "R-SIDANG-OTHER", [examinerOneId, examinerTwoId])),
    });
    expect(examinerConflict.status).toBe(409);

    const [unchanged] = await db.select().from(schema.thesisDefenses).where(eq(schema.thesisDefenses.id, conflictDefenseId));
    expect(unchanged.status).toBe("pending");
    expect(unchanged.scheduledAt).toBeNull();
    const audits = await db.select().from(schema.auditLogs).where(eq(schema.auditLogs.entityId, scheduledDefenseId));
    expect(audits.some((row) => row.action === "defense_scheduled")).toBe(true);
    const notifications = await db.select().from(schema.notifications).where(eq(schema.notifications.userId, studentId));
    expect(notifications.some((row) => row.title === "Sidang Terjadwal")).toBe(true);
  });

  it("does not allow cancelling an issued letter already used by scheduled Sidang", async () => {
    const rows = await db.select().from(schema.suratTugas).where(eq(schema.suratTugas.defenseId, scheduledDefenseId));
    const cancel = await request(`/api/v1/surat-tugas/${rows[0].id}/cancel`, {
      method: "POST",
      headers: { "content-type": "application/json", ...(await auth(adminId, "ADMIN_FAKULTAS")) },
      body: JSON.stringify({ reason: "Tidak boleh setelah dijadwalkan" }),
    });
    expect(cancel.status).toBe(409);
  });
});
