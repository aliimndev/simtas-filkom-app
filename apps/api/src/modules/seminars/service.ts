import { and, eq, inArray, sql } from "drizzle-orm";
import type { Db } from "@sims/db";
import { schema } from "@sims/db";
import { getDb } from "../../db";
import { loadConfig } from "../../config";
import { recordAudit } from "./audit";
import { notifyKaprodi, notifySchedule } from "./notifications";
import { approvedLatestSeminarDocument, getSeminar, getThesis, isUniqueViolation, listSeminars } from "./queries";
import { ACTIVE_SEMINAR_STATUSES, STAFF_ROLES, SeminarError, type SeminarActor, type SeminarListFilter, type ScheduleSeminarInput } from "./types";

export { ACTIVE_SEMINAR_STATUSES, SeminarError } from "./types";
export type { SeminarActor, SeminarListFilter, ScheduleSeminarInput } from "./types";
export { getSeminar, listSeminars } from "./queries";

function db(): Db {
  return getDb(loadConfig().databaseUrl);
}

export async function submitSeminar(thesisId: string, actor: SeminarActor) {
  if (actor.role !== "MAHASISWA") throw new SeminarError("FORBIDDEN", "Hanya Mahasiswa yang dapat mengajukan Seminar", 403);

  let result: any;
  try {
    result = await db().transaction(async (tx) => {
      const thesis = await getThesis(tx, thesisId);
      if (!thesis) throw new SeminarError("NOT_FOUND", "Thesis tidak ditemukan", 404);
      if (thesis.studentId !== actor.userId) throw new SeminarError("FORBIDDEN", "Bukan pemilik Thesis", 403);

      const active = await tx.select({ id: schema.seminars.id }).from(schema.seminars).where(and(
        eq(schema.seminars.thesisId, thesisId as any),
        sql`${schema.seminars.status} IN ('pending', 'scheduled', 'passed')`,
      )).limit(1);
      if (active.length > 0) throw new SeminarError("CONFLICT", "Thesis ini sudah memiliki Seminar aktif", 409);
      if (thesis.status !== "in_progress") throw new SeminarError("VALIDATION", "Thesis harus berstatus in_progress untuk mengajukan Seminar", 422);
      if (!(await approvedLatestSeminarDocument(tx, thesisId))) throw new SeminarError("GATE_NOT_MET", "Dokumen Seminar terbaru belum disetujui", 422);

      const [seminar] = await tx.insert(schema.seminars).values({ thesisId: thesisId as any, status: "pending" }).returning();
      const updated = await tx.update(schema.theses)
        .set({ status: "seminar_ready", updatedAt: new Date() })
        .where(and(eq(schema.theses.id, thesisId as any), eq(schema.theses.status, "in_progress")))
        .returning({ id: schema.theses.id });
      if (updated.length === 0) throw new SeminarError("CONFLICT", "Status Thesis berubah, ulangi pengajuan", 409);
      await recordAudit(tx, actor, "seminar_submitted", seminar.id, { status: thesis.status }, { status: "pending", thesisStatus: "seminar_ready" });
      return seminar;
    });
  } catch (error) {
    if (isUniqueViolation(error)) throw new SeminarError("CONFLICT", "Thesis ini sudah memiliki Seminar aktif", 409);
    throw error;
  }

  await notifyKaprodi(result.id).catch(() => undefined);
  return getSeminar(result.id, actor);
}

const MIN_SCHEDULE_LEAD_MS = 3 * 24 * 60 * 60 * 1000;

function validateScheduleInput(input: ScheduleSeminarInput) {
  const scheduledAt = new Date(input.scheduledAt);
  if (!input.scheduledAt || Number.isNaN(scheduledAt.getTime()) || scheduledAt.getTime() < Date.now() + MIN_SCHEDULE_LEAD_MS) {
    throw new SeminarError("VALIDATION", "Jadwal Seminar harus minimal 3 hari dari sekarang", 422);
  }
  const room = input.room.trim();
  if (!room) throw new SeminarError("VALIDATION", "room wajib diisi", 422);
  if (!Array.isArray(input.examinerIds) || input.examinerIds.length < 2) throw new SeminarError("VALIDATION", "Seminar membutuhkan minimal 2 Penguji", 422);
  const examinerIds = input.examinerIds.map((id) => String(id).trim());
  if (examinerIds.some((id) => !id) || new Set(examinerIds).size !== examinerIds.length) throw new SeminarError("VALIDATION", "examiner_ids harus berisi ID unik", 422);
  return { scheduledAt, room, examinerIds };
}

export async function scheduleSeminar(id: string, input: ScheduleSeminarInput, actor: SeminarActor) {
  if (!STAFF_ROLES.has(actor.role)) throw new SeminarError("FORBIDDEN", "Hanya Kaprodi atau Admin Fakultas yang dapat menjadwalkan Seminar", 403);
  const schedule = validateScheduleInput(input);
  const result = await db().transaction(async (tx) => {
    const rows = await tx.select().from(schema.seminars).where(eq(schema.seminars.id, id as any));
    const seminar = rows[0];
    if (!seminar) throw new SeminarError("NOT_FOUND", "Seminar tidak ditemukan", 404);
    if (seminar.status !== "pending" && seminar.status !== "scheduled") throw new SeminarError("CONFLICT", "Seminar tidak dapat dijadwalkan dari status ini", 409);
    const thesis = await getThesis(tx, seminar.thesisId);
    if (!thesis) throw new SeminarError("NOT_FOUND", "Thesis tidak ditemukan", 404);

    const roles = await tx.select({ id: schema.roles.id }).from(schema.roles).where(eq(schema.roles.name, "dosen_penguji"));
    if (!roles[0]) throw new SeminarError("VALIDATION", "Role Dosen Penguji belum tersedia", 422);
    const examiners = await tx.select({ id: schema.users.id }).from(schema.users).where(and(
      inArray(schema.users.id, schedule.examinerIds as any),
      eq(schema.users.roleId, roles[0].id),
      eq(schema.users.isActive, true),
    ));
    if (examiners.length !== schedule.examinerIds.length) throw new SeminarError("VALIDATION", "Semua Penguji harus merupakan Dosen Penguji aktif", 422);

    const sameSlot = await tx.select({ id: schema.seminars.id, room: schema.seminars.room }).from(schema.seminars).where(and(
      eq(schema.seminars.status, "scheduled"),
      eq(schema.seminars.scheduledAt, schedule.scheduledAt),
      sql`${schema.seminars.id} <> ${id}`,
    ));
    if (sameSlot.some((candidate) => candidate.room === schedule.room)) throw new SeminarError("CONFLICT", "Ruangan sudah digunakan pada waktu tersebut", 409);
    const otherSlotIds = sameSlot.map((candidate) => candidate.id);
    if (otherSlotIds.length > 0) {
      const conflict = await tx.select({ id: schema.seminarExaminers.id }).from(schema.seminarExaminers).where(and(
        inArray(schema.seminarExaminers.seminarId, otherSlotIds as any),
        inArray(schema.seminarExaminers.examinerId, schedule.examinerIds as any),
      )).limit(1);
      if (conflict.length > 0) throw new SeminarError("CONFLICT", "Penguji sudah memiliki Seminar pada waktu tersebut", 409);
    }

    await tx.delete(schema.seminarExaminers).where(eq(schema.seminarExaminers.seminarId, id as any));
    await tx.insert(schema.seminarExaminers).values(schedule.examinerIds.map((examinerId) => ({ seminarId: id, examinerId, assignedBy: actor.userId })) as any);
    const [updated] = await tx.update(schema.seminars)
      .set({ status: "scheduled", scheduledAt: schedule.scheduledAt, room: schedule.room, updatedAt: new Date() } as any)
      .where(eq(schema.seminars.id, id as any)).returning();
    await recordAudit(tx, actor, "seminar_scheduled", id,
      { status: seminar.status, scheduled_at: seminar.scheduledAt, room: seminar.room },
      { status: "scheduled", scheduled_at: schedule.scheduledAt.toISOString(), room: schedule.room, examiner_ids: schedule.examinerIds },
    );
    return { studentId: thesis.studentId, seminar: updated };
  });

  await notifySchedule(id, result.studentId, schedule.examinerIds).catch(() => undefined);
  return getSeminar(id, actor);
}
