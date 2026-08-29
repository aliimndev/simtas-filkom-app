import { and, eq, inArray, sql } from "drizzle-orm";
import type { Db } from "@sims/db";
import { schema } from "@sims/db";
import { getDb } from "../../db";
import { loadConfig } from "../../config";
import { recordAudit } from "./audit";
import { notifyFinalization, notifySchedule, notifySubmission } from "./notifications";
import { approvedLatestDefenseDocument, getDefense, isUniqueViolation } from "./queries";
import { DefenseError, STAFF_ROLES, type DefenseActor, type DefenseFinalizeInput, type DefenseScoreInput, type ScheduleDefenseInput, type SuratTugasInput } from "./types";

export { DefenseError } from "./types";
export { getDefense, listDefenses } from "./queries";

function db(): Db {
  return getDb(loadConfig().databaseUrl);
}

function validateLetterInput(input: SuratTugasInput) {
  const letterNumber = input.letterNumber.trim();
  const fileName = input.fileName.trim();
  const fileUrl = input.fileUrl.trim();
  const issueDate = new Date(input.issueDate);
  if (!letterNumber || !fileName || !fileUrl || Number.isNaN(issueDate.getTime())) {
    throw new DefenseError("VALIDATION", "letter_number, issue_date, file_name, dan file_url wajib diisi", 422);
  }
  return { letterNumber, issueDate: input.issueDate, fileName, fileUrl };
}

function validateScheduleInput(input: ScheduleDefenseInput) {
  const scheduledAt = new Date(input.scheduledAt);
  const room = input.room.trim();
  const examinerIds = input.examinerIds.map((id) => String(id).trim());
  if (Number.isNaN(scheduledAt.getTime()) || scheduledAt.getTime() < Date.now() + 7 * 86400000) {
    throw new DefenseError("VALIDATION", "Jadwal Sidang harus minimal 7 hari dari sekarang", 422);
  }
  if (!room) throw new DefenseError("VALIDATION", "room wajib diisi", 422);
  if (examinerIds.length < 2 || new Set(examinerIds).size !== examinerIds.length) {
    throw new DefenseError("VALIDATION", "Sidang membutuhkan minimal 2 Penguji yang berbeda", 422);
  }
  return { scheduledAt, room, examinerIds };
}

export async function submitDefense(thesisId: string, actor: DefenseActor) {
  if (actor.role !== "MAHASISWA") throw new DefenseError("FORBIDDEN", "Hanya Mahasiswa yang dapat mengajukan Sidang", 403);

  let result: any;
  try {
    result = await db().transaction(async (tx) => {
      const thesisRows = await tx.select().from(schema.theses).where(eq(schema.theses.id, thesisId as any)).for("update");
      const thesis = thesisRows[0];
      if (!thesis) throw new DefenseError("NOT_FOUND", "Thesis tidak ditemukan", 404);
      if (thesis.studentId !== actor.userId) throw new DefenseError("FORBIDDEN", "Bukan pemilik Thesis", 403);
      if (thesis.status !== "defense_ready") throw new DefenseError("VALIDATION", "Thesis harus berstatus defense_ready untuk mengajukan Sidang", 422);
      if (!(await approvedLatestDefenseDocument(tx, thesisId))) throw new DefenseError("GATE_NOT_MET", "Dokumen Sidang terbaru belum disetujui", 422);

      const active = await tx.select({ id: schema.thesisDefenses.id })
        .from(schema.thesisDefenses)
        .where(and(eq(schema.thesisDefenses.thesisId, thesisId as any), sql`${schema.thesisDefenses.status} IN ('pending', 'scheduled')`))
        .limit(1);
      if (active.length > 0) throw new DefenseError("CONFLICT", "Thesis ini sudah memiliki Sidang aktif", 409);

      const [defense] = await tx.insert(schema.thesisDefenses).values({ thesisId: thesisId as any, status: "pending" }).returning();
      await recordAudit(tx, actor, "defense_submitted", defense.id, { thesisStatus: thesis.status }, { status: "pending", thesisStatus: thesis.status });
      return { defense, studentId: thesis.studentId };
    });
  } catch (error) {
    if (isUniqueViolation(error)) throw new DefenseError("CONFLICT", "Thesis ini sudah memiliki Sidang aktif", 409);
    throw error;
  }

  await notifySubmission(result.defense.id, result.studentId).catch(() => undefined);
  return getDefense(result.defense.id, actor);
}

export async function createSuratTugas(defenseId: string, input: SuratTugasInput, actor: DefenseActor) {
  if (!STAFF_ROLES.has(actor.role)) throw new DefenseError("FORBIDDEN", "Hanya Kaprodi atau Admin Fakultas yang dapat membuat Surat Tugas", 403);
  const letter = validateLetterInput(input);
  try {
    return await db().transaction(async (tx) => {
      const defenses = await tx.select().from(schema.thesisDefenses).where(eq(schema.thesisDefenses.id, defenseId as any)).for("update");
      const defense = defenses[0];
      if (!defense) throw new DefenseError("NOT_FOUND", "Sidang tidak ditemukan", 404);
      if (defense.status !== "pending") throw new DefenseError("CONFLICT", "Surat Tugas hanya dapat dibuat untuk Sidang pending", 409);
      const [created] = await tx.insert(schema.suratTugas).values({
        defenseId: defenseId as any,
        letterNumber: letter.letterNumber,
        issueDate: letter.issueDate,
        fileName: letter.fileName,
        fileUrl: letter.fileUrl,
        issuerId: actor.userId,
        status: "draft",
      } as any).returning();
      await recordAudit(tx, actor, "surat_tugas_created", created.id, null, { defense_id: defenseId, status: "draft", letter_number: letter.letterNumber });
      return created;
    });
  } catch (error) {
    if (isUniqueViolation(error)) throw new DefenseError("CONFLICT", "Sidang ini sudah memiliki Surat Tugas aktif atau nomor surat sudah digunakan", 409);
    throw error;
  }
}

export async function issueSuratTugas(id: string, actor: DefenseActor) {
  if (!STAFF_ROLES.has(actor.role)) throw new DefenseError("FORBIDDEN", "Hanya Kaprodi atau Admin Fakultas yang dapat menerbitkan Surat Tugas", 403);
  return db().transaction(async (tx) => {
    const rows = await tx.select().from(schema.suratTugas).where(eq(schema.suratTugas.id, id as any)).for("update");
    const letter = rows[0];
    if (!letter) throw new DefenseError("NOT_FOUND", "Surat Tugas tidak ditemukan", 404);
    if (letter.status !== "draft") throw new DefenseError("CONFLICT", "Surat Tugas tidak dapat diterbitkan dari status ini", 409);
    const [updated] = await tx.update(schema.suratTugas).set({ status: "issued", updatedAt: new Date() }).where(eq(schema.suratTugas.id, id as any)).returning();
    await recordAudit(tx, actor, "surat_tugas_issued", id, { status: "draft" }, { status: "issued" });
    return updated;
  });
}

export async function cancelSuratTugas(id: string, reason: string, actor: DefenseActor) {
  if (!STAFF_ROLES.has(actor.role)) throw new DefenseError("FORBIDDEN", "Hanya Kaprodi atau Admin Fakultas yang dapat membatalkan Surat Tugas", 403);
  const cancellationReason = reason.trim();
  if (!cancellationReason) throw new DefenseError("VALIDATION", "reason wajib diisi", 422);
  return db().transaction(async (tx) => {
    const rows = await tx.select().from(schema.suratTugas).where(eq(schema.suratTugas.id, id as any)).for("update");
    const letter = rows[0];
    if (!letter) throw new DefenseError("NOT_FOUND", "Surat Tugas tidak ditemukan", 404);
    if (letter.status === "cancelled") throw new DefenseError("CONFLICT", "Surat Tugas sudah dibatalkan", 409);
    if (letter.status === "issued") {
      const defenses = await tx.select({ status: schema.thesisDefenses.status }).from(schema.thesisDefenses).where(eq(schema.thesisDefenses.id, letter.defenseId)).for("update");
      if (defenses[0]?.status === "scheduled") throw new DefenseError("CONFLICT", "Surat Tugas yang sudah digunakan untuk Sidang terjadwal tidak dapat dibatalkan langsung", 409);
    }
    const [updated] = await tx.update(schema.suratTugas).set({ status: "cancelled", cancellationReason, updatedAt: new Date() }).where(eq(schema.suratTugas.id, id as any)).returning();
    await recordAudit(tx, actor, "surat_tugas_cancelled", id, { status: letter.status }, { status: "cancelled", cancellation_reason: cancellationReason });
    return updated;
  });
}

export async function scheduleDefense(id: string, input: ScheduleDefenseInput, actor: DefenseActor) {
  if (!STAFF_ROLES.has(actor.role)) throw new DefenseError("FORBIDDEN", "Hanya Kaprodi atau Admin Fakultas yang dapat menjadwalkan Sidang", 403);
  const schedule = validateScheduleInput(input);
  const result = await db().transaction(async (tx) => {
    const rows = await tx.select().from(schema.thesisDefenses).where(eq(schema.thesisDefenses.id, id as any)).for("update");
    const defense = rows[0];
    if (!defense) throw new DefenseError("NOT_FOUND", "Sidang tidak ditemukan", 404);
    if (defense.status !== "pending") throw new DefenseError("CONFLICT", "Sidang tidak dapat dijadwalkan dari status ini", 409);
    const letters = await tx.select().from(schema.suratTugas).where(and(eq(schema.suratTugas.defenseId, id as any), eq(schema.suratTugas.status, "issued"))).limit(1);
    if (!letters[0]) throw new DefenseError("CONFLICT", "Sidang membutuhkan Surat Tugas yang sudah diterbitkan", 409);

    const roles = await tx.select({ id: schema.roles.id }).from(schema.roles).where(eq(schema.roles.name, "dosen_penguji"));
    if (!roles[0]) throw new DefenseError("VALIDATION", "Role Dosen Penguji belum tersedia", 422);
    const examiners = await tx.select({ id: schema.users.id }).from(schema.users).where(and(
      inArray(schema.users.id, schedule.examinerIds as any),
      eq(schema.users.roleId, roles[0].id),
      eq(schema.users.isActive, true),
    ));
    if (examiners.length !== schedule.examinerIds.length) throw new DefenseError("VALIDATION", "Semua Penguji harus merupakan Dosen Penguji aktif", 422);

    const sameSlot = await tx.select({ id: schema.thesisDefenses.id, room: schema.thesisDefenses.room })
      .from(schema.thesisDefenses)
      .where(and(eq(schema.thesisDefenses.status, "scheduled"), eq(schema.thesisDefenses.scheduledAt, schedule.scheduledAt), sql`${schema.thesisDefenses.id} <> ${id}`));
    if (sameSlot.some((candidate) => candidate.room === schedule.room)) throw new DefenseError("CONFLICT", "Ruangan sudah digunakan pada waktu tersebut", 409);
    const otherIds = sameSlot.map((candidate) => candidate.id);
    if (otherIds.length > 0) {
      const conflict = await tx.select({ id: schema.defenseExaminers.id }).from(schema.defenseExaminers).where(and(inArray(schema.defenseExaminers.defenseId, otherIds as any), inArray(schema.defenseExaminers.examinerId, schedule.examinerIds as any))).limit(1);
      if (conflict.length > 0) throw new DefenseError("CONFLICT", "Penguji sudah memiliki Sidang pada waktu tersebut", 409);
    }

    await tx.delete(schema.defenseExaminers).where(eq(schema.defenseExaminers.defenseId, id as any));
    await tx.insert(schema.defenseExaminers).values(schedule.examinerIds.map((examinerId) => ({ defenseId: id, examinerId, assignedBy: actor.userId })) as any);
    const [updated] = await tx.update(schema.thesisDefenses).set({ status: "scheduled", scheduledAt: schedule.scheduledAt, room: schedule.room, updatedAt: new Date() }).where(eq(schema.thesisDefenses.id, id as any)).returning();
    await recordAudit(tx, actor, "defense_scheduled", id, { status: defense.status }, { status: "scheduled", scheduled_at: schedule.scheduledAt.toISOString(), room: schedule.room, examiner_ids: schedule.examinerIds });
    return { studentId: (await tx.select({ studentId: schema.theses.studentId }).from(schema.theses).where(eq(schema.theses.id, defense.thesisId)))[0]?.studentId ?? "", examinerIds: schedule.examinerIds, defense: updated };
  });

  await notifySchedule(id, result.studentId, result.examinerIds).catch(() => undefined);
  return getDefense(id, actor);
}

const DEFENSE_RUBRIC = new Map([
  ["Presentasi", 30],
  ["Penguasaan Materi", 30],
  ["Kualitas Naskah", 25],
  ["Kemampuan Menjawab", 15],
]);

function validateDefenseScores(scores: DefenseScoreInput[]) {
  if (scores.length !== DEFENSE_RUBRIC.size || new Set(scores.map((score) => score.componentName)).size !== DEFENSE_RUBRIC.size) {
    throw new DefenseError("VALIDATION", "Sidang harus memiliki tepat empat komponen rubric yang unik", 422);
  }
  for (const score of scores) {
    const expectedWeight = DEFENSE_RUBRIC.get(score.componentName);
    if (expectedWeight === undefined) throw new DefenseError("VALIDATION", "Komponen rubric Sidang tidak valid", 422);
    if (score.componentWeight !== expectedWeight) throw new DefenseError("VALIDATION", "Bobot rubric Sidang tidak dapat diubah", 422);
    if (!Number.isFinite(score.score) || score.score < 0 || score.score > 100) {
      throw new DefenseError("VALIDATION", "Nilai Sidang harus berada pada rentang 0-100", 422);
    }
    if (score.notes != null && typeof score.notes !== "string") {
      throw new DefenseError("VALIDATION", "Catatan Penguji harus berupa teks", 422);
    }
  }
}

export async function saveDefenseScores(id: string, scores: DefenseScoreInput[], actor: DefenseActor) {
  if (actor.role !== "DOSEN_PENGUJI") throw new DefenseError("FORBIDDEN", "Hanya Penguji yang dapat mengisi nilai Sidang", 403);
  validateDefenseScores(scores);

  await db().transaction(async (tx) => {
    const rows = await tx.select().from(schema.thesisDefenses).where(eq(schema.thesisDefenses.id, id as any)).for("update");
    const defense = rows[0];
    if (!defense) throw new DefenseError("NOT_FOUND", "Sidang tidak ditemukan", 404);
    if (defense.status !== "scheduled") throw new DefenseError("CONFLICT", "Sidang tidak sedang dalam status terjadwal", 409);
    const assignment = await tx.select({ id: schema.defenseExaminers.id })
      .from(schema.defenseExaminers)
      .where(and(eq(schema.defenseExaminers.defenseId, id as any), eq(schema.defenseExaminers.examinerId, actor.userId as any)));
    if (assignment.length === 0) throw new DefenseError("FORBIDDEN", "Penguji tidak ditugaskan pada Sidang ini", 403);

    for (const score of scores) {
      await tx.insert(schema.defenseScores).values({
        defenseId: id as any,
        examinerId: actor.userId as any,
        componentName: score.componentName,
        componentWeight: String(score.componentWeight),
        score: String(score.score),
        notes: score.notes ?? null,
      } as any).onConflictDoUpdate({
        target: [schema.defenseScores.defenseId, schema.defenseScores.examinerId, schema.defenseScores.componentName],
        set: {
          componentWeight: String(score.componentWeight),
          score: String(score.score),
          notes: score.notes ?? null,
          updatedAt: new Date(),
        } as any,
      });
    }

    await recordAudit(tx, actor, "defense_scores_updated", id, null, {
      examiner_id: actor.userId,
      components: scores.map((score) => ({ component_name: score.componentName, score: score.score })),
    });
  });

  return getDefense(id, actor);
}

export async function finalizeDefense(id: string, input: DefenseFinalizeInput, actor: DefenseActor) {
  if (!STAFF_ROLES.has(actor.role)) throw new DefenseError("FORBIDDEN", "Hanya Kaprodi atau Admin Fakultas yang dapat memfinalisasi Sidang", 403);

  const result = await db().transaction(async (tx) => {
    const rows = await tx.select().from(schema.thesisDefenses).where(eq(schema.thesisDefenses.id, id as any)).for("update");
    const defense = rows[0];
    if (!defense) throw new DefenseError("NOT_FOUND", "Sidang tidak ditemukan", 404);
    if (defense.status !== "scheduled") throw new DefenseError("CONFLICT", "Sidang hanya dapat difinalisasi dari status scheduled", 409);

    const thesisRows = await tx.select().from(schema.theses).where(eq(schema.theses.id, defense.thesisId)).for("update");
    const thesis = thesisRows[0];
    if (!thesis) throw new DefenseError("NOT_FOUND", "Thesis tidak ditemukan", 404);
    const examiners = await tx.select({ examinerId: schema.defenseExaminers.examinerId })
      .from(schema.defenseExaminers)
      .where(eq(schema.defenseExaminers.defenseId, id as any));
    const scores = await tx.select().from(schema.defenseScores).where(eq(schema.defenseScores.defenseId, id as any));
    const expected = examiners.length * DEFENSE_RUBRIC.size;
    if (examiners.length < 2 || scores.length !== expected) {
      throw new DefenseError("VALIDATION", "Sidang membutuhkan minimal 2 Penguji dengan seluruh komponen rubric", 422);
    }

    const assigned = new Set(examiners.map((examiner) => examiner.examinerId));
    const byExaminer = new Map<string, any[]>();
    for (const score of scores) {
      const expectedWeight = DEFENSE_RUBRIC.get(score.componentName);
      const numericScore = Number(score.score);
      if (!assigned.has(score.examinerId) || expectedWeight === undefined || Number(score.componentWeight) !== expectedWeight || !Number.isFinite(numericScore) || numericScore < 0 || numericScore > 100) {
        throw new DefenseError("VALIDATION", "Data nilai Sidang tidak lengkap atau tidak valid", 422);
      }
      const current = byExaminer.get(score.examinerId) ?? [];
      current.push(score);
      byExaminer.set(score.examinerId, current);
    }
    if (byExaminer.size !== examiners.length || [...byExaminer.values()].some((items) => items.length !== DEFENSE_RUBRIC.size || new Set(items.map((item) => item.componentName)).size !== DEFENSE_RUBRIC.size)) {
      throw new DefenseError("VALIDATION", "Semua Penguji harus mengisi seluruh komponen rubric sebelum finalisasi", 422);
    }

    const examinerTotals = [...byExaminer.values()].map((items) => items.reduce((total, item) => total + Number(item.score) * Number(item.componentWeight) / 100, 0));
    const finalScore = Number((examinerTotals.reduce((total, value) => total + value, 0) / examinerTotals.length).toFixed(2));
    const status: "passed" | "revision_required" | "failed" = finalScore >= 75 ? "passed" : finalScore >= 60 ? "revision_required" : "failed";
    const revisionNotes = input.revisionNotes?.trim() || null;
    if (status === "revision_required" && !revisionNotes) {
      throw new DefenseError("VALIDATION", "revision_notes wajib diisi untuk hasil revision_required", 422);
    }

    const updated = await tx.update(schema.thesisDefenses)
      .set({ status, finalScore: String(finalScore), revisionNotes, updatedAt: new Date() } as any)
      .where(and(eq(schema.thesisDefenses.id, id as any), eq(schema.thesisDefenses.status, "scheduled")))
      .returning({ id: schema.thesisDefenses.id });
    if (updated.length === 0) throw new DefenseError("CONFLICT", "Sidang sudah difinalisasi oleh request lain", 409);

    const thesisStatus = status === "failed" ? thesis.status : "defense_done";
    if (status !== "failed") {
      const thesisUpdated = await tx.update(schema.theses)
        .set({ status: "defense_done", updatedAt: new Date() })
        .where(and(eq(schema.theses.id, thesis.id), eq(schema.theses.status, "defense_ready")))
        .returning({ id: schema.theses.id });
      if (thesisUpdated.length === 0) throw new DefenseError("CONFLICT", "Status Thesis tidak dapat dipindahkan ke defense_done", 409);
    }

    await recordAudit(tx, actor, "defense_finalized", id, { status: defense.status, thesis_status: thesis.status }, {
      status,
      final_score: finalScore,
      revision_notes: revisionNotes,
      thesis_status: thesisStatus,
    });
    return { studentId: thesis.studentId, examinerIds: examiners.map((examiner) => examiner.examinerId), status, finalScore };
  });

  await notifyFinalization(id, result.studentId, result.examinerIds, result.status, result.finalScore).catch(() => undefined);
  return getDefense(id, actor);
}
