import { and, eq } from "drizzle-orm";
import type { Db } from "@sims/db";
import { schema } from "@sims/db";
import { getDb } from "../../db";
import { loadConfig } from "../../config";
import { recordAudit } from "./audit";
import { notifyFinalization } from "./notifications";
import { getSeminar, getThesis } from "./queries";
import { STAFF_ROLES, SeminarError, type SeminarActor, type SeminarScoreInput } from "./types";

const RUBRIC = new Map([
  ["Presentasi", 30],
  ["Penguasaan Materi", 30],
  ["Kualitas Naskah", 25],
  ["Kemampuan Menjawab", 15],
]);

function db(): Db {
  return getDb(loadConfig().databaseUrl);
}

function validateScores(scores: SeminarScoreInput[]) {
  if (scores.length !== RUBRIC.size || new Set(scores.map((score) => score.componentName)).size !== RUBRIC.size) {
    throw new SeminarError("VALIDATION", "Seminar harus memiliki tepat empat komponen rubric yang unik", 422);
  }

  for (const score of scores) {
    const expectedWeight = RUBRIC.get(score.componentName);
    if (expectedWeight === undefined) throw new SeminarError("VALIDATION", "Komponen rubric Seminar tidak valid", 422);
    if (score.componentWeight !== expectedWeight) throw new SeminarError("VALIDATION", "Bobot rubric Seminar tidak dapat diubah", 422);
    if (!Number.isFinite(score.score) || score.score < 0 || score.score > 100) {
      throw new SeminarError("VALIDATION", "Nilai Seminar harus berada pada rentang 0-100", 422);
    }
    if (score.notes != null && typeof score.notes !== "string") {
      throw new SeminarError("VALIDATION", "Catatan Penguji harus berupa teks", 422);
    }
  }
}

export async function saveSeminarScores(id: string, scores: SeminarScoreInput[], actor: SeminarActor) {
  if (actor.role !== "DOSEN_PENGUJI") throw new SeminarError("FORBIDDEN", "Hanya Penguji yang dapat mengisi nilai Seminar", 403);
  validateScores(scores);

  await db().transaction(async (tx) => {
    const rows = await tx.select().from(schema.seminars).where(eq(schema.seminars.id, id as any)).for("update");
    const seminar = rows[0];
    if (!seminar) throw new SeminarError("NOT_FOUND", "Seminar tidak ditemukan", 404);
    if (seminar.status !== "scheduled") throw new SeminarError("CONFLICT", "Seminar tidak sedang dalam status terjadwal", 409);

    const assignment = await tx.select({ id: schema.seminarExaminers.id })
      .from(schema.seminarExaminers)
      .where(and(eq(schema.seminarExaminers.seminarId, id as any), eq(schema.seminarExaminers.examinerId, actor.userId as any)));
    if (assignment.length === 0) throw new SeminarError("FORBIDDEN", "Penguji tidak ditugaskan pada Seminar ini", 403);

    for (const score of scores) {
      await tx.insert(schema.seminarScores).values({
        seminarId: id as any,
        examinerId: actor.userId as any,
        componentName: score.componentName,
        componentWeight: String(score.componentWeight),
        score: String(score.score),
        notes: score.notes ?? null,
      } as any).onConflictDoUpdate({
        target: [schema.seminarScores.seminarId, schema.seminarScores.examinerId, schema.seminarScores.componentName],
        set: {
          componentWeight: String(score.componentWeight),
          score: String(score.score),
          notes: score.notes ?? null,
          updatedAt: new Date(),
        } as any,
      });
    }

    await recordAudit(tx, actor, "seminar_scores_updated", id, null, {
      examiner_id: actor.userId,
      components: scores.map((score) => ({ component_name: score.componentName, score: score.score })),
    });
  });

  return getSeminar(id, actor);
}

export async function finalizeSeminar(id: string, actor: SeminarActor) {
  if (!STAFF_ROLES.has(actor.role)) throw new SeminarError("FORBIDDEN", "Hanya Kaprodi atau Admin Fakultas yang dapat memfinalisasi Seminar", 403);

  let result: { studentId: string; examinerIds: string[]; status: "passed" | "failed"; finalScore: number };
  try {
    result = await db().transaction(async (tx) => {
      const rows = await tx.select().from(schema.seminars).where(eq(schema.seminars.id, id as any)).for("update");
      const seminar = rows[0];
      if (!seminar) throw new SeminarError("NOT_FOUND", "Seminar tidak ditemukan", 404);
      if (seminar.status !== "scheduled") throw new SeminarError("CONFLICT", "Seminar hanya dapat difinalisasi dari status scheduled", 409);

      const thesis = await getThesis(tx, seminar.thesisId);
      if (!thesis) throw new SeminarError("NOT_FOUND", "Thesis tidak ditemukan", 404);
      const examiners = await tx.select({ examinerId: schema.seminarExaminers.examinerId })
        .from(schema.seminarExaminers)
        .where(eq(schema.seminarExaminers.seminarId, id as any));
      const scores = await tx.select().from(schema.seminarScores).where(eq(schema.seminarScores.seminarId, id as any));
      const expected = examiners.length * RUBRIC.size;
      if (examiners.length < 2 || scores.length !== expected) {
        throw new SeminarError("VALIDATION", "Seminar membutuhkan minimal 2 Penguji dengan seluruh komponen rubric", 422);
      }

      const assigned = new Set(examiners.map((examiner) => examiner.examinerId));
      if (assigned.size !== examiners.length) {
        throw new SeminarError("VALIDATION", "Penugasan Penguji Seminar tidak valid", 422);
      }
      const byExaminer = new Map<string, any[]>();
      for (const score of scores) {
        const expectedWeight = RUBRIC.get(score.componentName);
        const numericScore = Number(score.score);
        if (
          !assigned.has(score.examinerId)
          || expectedWeight === undefined
          || Number(score.componentWeight) !== expectedWeight
          || !Number.isFinite(numericScore)
          || numericScore < 0
          || numericScore > 100
        ) {
          throw new SeminarError("VALIDATION", "Data nilai Seminar tidak lengkap atau tidak valid", 422);
        }
        const current = byExaminer.get(score.examinerId) ?? [];
        current.push(score);
        byExaminer.set(score.examinerId, current);
      }
      if (byExaminer.size !== examiners.length || [...byExaminer.values()].some((items) => items.length !== RUBRIC.size || new Set(items.map((item) => item.componentName)).size !== RUBRIC.size)) {
        throw new SeminarError("VALIDATION", "Semua Penguji harus mengisi seluruh komponen rubric sebelum finalisasi", 422);
      }

      const examinerTotals = [...byExaminer.values()].map((items) => items.reduce((total, item) => total + Number(item.score) * Number(item.componentWeight) / 100, 0));
      const finalScore = Number((examinerTotals.reduce((total, value) => total + value, 0) / examinerTotals.length).toFixed(2));
      const status = finalScore >= 60 ? "passed" : "failed";
      const updated = await tx.update(schema.seminars)
        .set({ status, finalScore: String(finalScore), updatedAt: new Date() } as any)
        .where(and(eq(schema.seminars.id, id as any), eq(schema.seminars.status, "scheduled")))
        .returning({ id: schema.seminars.id });
      if (updated.length === 0) throw new SeminarError("CONFLICT", "Seminar sudah difinalisasi oleh request lain", 409);

      if (status === "passed") {
        const thesisUpdated = await tx.update(schema.theses)
          .set({ status: "defense_ready", updatedAt: new Date() })
          .where(and(eq(schema.theses.id, thesis.id), eq(schema.theses.status, "seminar_ready")))
          .returning({ id: schema.theses.id });
        if (thesisUpdated.length === 0) throw new SeminarError("CONFLICT", "Status Thesis tidak dapat dipindahkan ke defense_ready", 409);
      }

      await recordAudit(tx, actor, "seminar_finalized", id, { status: seminar.status, thesis_status: thesis.status }, {
        status,
        final_score: finalScore,
        thesis_status: status === "passed" ? "defense_ready" : thesis.status,
        milestone: status === "passed" ? "seminar_done" : undefined,
      });
      return { studentId: thesis.studentId, examinerIds: examiners.map((examiner) => examiner.examinerId), status, finalScore };
    });
  } catch (error) {
    throw error;
  }

  await notifyFinalization(id, result.studentId, result.examinerIds, result.status, result.finalScore).catch(() => undefined);
  return getSeminar(id, actor);
}
