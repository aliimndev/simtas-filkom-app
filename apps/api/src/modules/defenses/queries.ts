import { and, count, desc, eq, exists, sql } from "drizzle-orm";
import type { Db } from "@sims/db";
import { schema } from "@sims/db";
import { getDb } from "../../db";
import { loadConfig } from "../../config";
import { DefenseError, STAFF_ROLES, type DefenseActor, type DefenseListFilter } from "./types";

function db(): Db {
  return getDb(loadConfig().databaseUrl);
}

export function isUniqueViolation(error: unknown): boolean {
  return typeof error === "object" && error !== null && "code" in error && (error as { code?: string }).code === "23505";
}

export async function approvedLatestDefenseDocument(connection: Pick<Db, "select">, thesisId: string) {
  const rows = await connection
    .select({ status: schema.documents.status })
    .from(schema.documents)
    .where(and(eq(schema.documents.thesisId, thesisId as any), eq(schema.documents.documentType, "defense_doc")))
    .orderBy(desc(schema.documents.version), desc(schema.documents.createdAt))
    .limit(1);
  return rows[0]?.status === "approved";
}

export async function getThesis(connection: Pick<Db, "select">, thesisId: string) {
  const rows = await connection.select().from(schema.theses).where(eq(schema.theses.id, thesisId as any));
  return rows[0] ?? null;
}

async function canAccess(connection: Pick<Db, "select">, defense: any, actor: DefenseActor) {
  if (STAFF_ROLES.has(actor.role)) return true;
  if (actor.role === "MAHASISWA") return defense.studentId === actor.userId;
  if (actor.role === "DOSEN_PEMBIMBING") {
    const rows = await connection.select({ id: schema.thesisSupervisors.id })
      .from(schema.thesisSupervisors)
      .where(and(eq(schema.thesisSupervisors.thesisId, defense.thesisId), eq(schema.thesisSupervisors.supervisorId, actor.userId as any)))
      .limit(1);
    return rows.length > 0;
  }
  if (actor.role === "DOSEN_PENGUJI") {
    const rows = await connection.select({ id: schema.defenseExaminers.id })
      .from(schema.defenseExaminers)
      .where(and(eq(schema.defenseExaminers.defenseId, defense.id), eq(schema.defenseExaminers.examinerId, actor.userId as any)))
      .limit(1);
    return rows.length > 0;
  }
  return false;
}

function iso(value: unknown): string | null {
  if (!value) return null;
  return value instanceof Date ? value.toISOString() : new Date(String(value)).toISOString();
}

export async function hydrate(connection: Pick<Db, "select">, row: any) {
  const thesisRows = await connection
    .select({
      id: schema.theses.id,
      title: schema.theses.title,
      status: schema.theses.status,
      studentId: schema.theses.studentId,
      studentName: schema.users.fullName,
      studentNim: schema.users.nimNidn,
    })
    .from(schema.theses)
    .innerJoin(schema.users, eq(schema.users.id, schema.theses.studentId))
    .where(eq(schema.theses.id, row.thesisId as any));
  const thesis = thesisRows[0];
  const examiners = await connection
    .select({ id: schema.users.id, fullName: schema.users.fullName, nimNidn: schema.users.nimNidn })
    .from(schema.defenseExaminers)
    .innerJoin(schema.users, eq(schema.users.id, schema.defenseExaminers.examinerId))
    .where(eq(schema.defenseExaminers.defenseId, row.id as any));
  const scores = await connection
    .select({
      id: schema.defenseScores.id,
      examinerId: schema.defenseScores.examinerId,
      examinerName: schema.users.fullName,
      componentName: schema.defenseScores.componentName,
      componentWeight: schema.defenseScores.componentWeight,
      score: schema.defenseScores.score,
      notes: schema.defenseScores.notes,
    })
    .from(schema.defenseScores)
    .innerJoin(schema.users, eq(schema.users.id, schema.defenseScores.examinerId))
    .where(eq(schema.defenseScores.defenseId, row.id as any));

  return {
    id: row.id,
    thesisId: row.thesisId,
    thesisTitle: thesis?.title ?? null,
    thesisStatus: thesis?.status ?? null,
    thesis: thesis ? {
      id: thesis.id,
      title: thesis.title,
      status: thesis.status,
      student: { id: thesis.studentId, fullName: thesis.studentName, nim: thesis.studentNim ?? null },
    } : null,
    student: thesis ? { id: thesis.studentId, fullName: thesis.studentName, nim: thesis.studentNim ?? null } : null,
    status: row.status,
    scheduledAt: iso(row.scheduledAt),
    room: row.room ?? null,
    revisionNotes: row.revisionNotes ?? null,
    finalScore: row.finalScore == null ? null : Number(row.finalScore),
    examiners: examiners.map((examiner: any) => ({ id: examiner.id, fullName: examiner.fullName, nim: examiner.nimNidn ?? null })),
    scores: scores.map((score: any) => ({
      id: score.id,
      examinerId: score.examinerId,
      examiner: { id: score.examinerId, fullName: score.examinerName },
      componentName: score.componentName,
      componentWeight: Number(score.componentWeight),
      score: Number(score.score),
      notes: score.notes ?? null,
    })),
    createdAt: iso(row.createdAt),
    updatedAt: iso(row.updatedAt),
  };
}

function scopedWhere(actor: DefenseActor) {
  if (STAFF_ROLES.has(actor.role)) return undefined;
  if (actor.role === "MAHASISWA") {
    return exists(db().select({ one: sql`1` }).from(schema.theses).where(and(eq(schema.theses.id, schema.thesisDefenses.thesisId), eq(schema.theses.studentId, actor.userId as any))));
  }
  if (actor.role === "DOSEN_PEMBIMBING") {
    return exists(db().select({ one: sql`1` }).from(schema.thesisSupervisors).where(and(eq(schema.thesisSupervisors.thesisId, schema.thesisDefenses.thesisId), eq(schema.thesisSupervisors.supervisorId, actor.userId as any))));
  }
  if (actor.role === "DOSEN_PENGUJI") {
    return exists(db().select({ one: sql`1` }).from(schema.defenseExaminers).where(and(eq(schema.defenseExaminers.defenseId, schema.thesisDefenses.id), eq(schema.defenseExaminers.examinerId, actor.userId as any))));
  }
  return sql`false`;
}

export async function listDefenses(filter: DefenseListFilter, actor: DefenseActor) {
  const connection = db();
  const page = Math.max(1, Math.floor(filter.page || 1));
  const perPage = Math.min(100, Math.max(1, Math.floor(filter.perPage || 20)));
  const where = and(scopedWhere(actor), filter.thesisId ? eq(schema.thesisDefenses.thesisId, filter.thesisId as any) : undefined);
  const totalRows = await connection.select({ value: count() }).from(schema.thesisDefenses).where(where);
  const rows = await connection.select().from(schema.thesisDefenses).where(where).orderBy(desc(schema.thesisDefenses.createdAt)).limit(perPage).offset((page - 1) * perPage);
  return {
    data: await Promise.all(rows.map((row: any) => hydrate(connection, row))),
    meta: { page, perPage, total: Number(totalRows[0]?.value ?? 0), totalPages: Math.ceil(Number(totalRows[0]?.value ?? 0) / perPage) },
  };
}

export async function getDefense(id: string, actor: DefenseActor) {
  const connection = db();
  const rows = await connection.select().from(schema.thesisDefenses).where(eq(schema.thesisDefenses.id, id as any));
  const defense = rows[0];
  if (!defense) throw new DefenseError("NOT_FOUND", "Sidang tidak ditemukan", 404);
  const thesis = await getThesis(connection, defense.thesisId);
  if (!thesis) throw new DefenseError("NOT_FOUND", "Thesis tidak ditemukan", 404);
  if (!(await canAccess(connection, { ...defense, studentId: thesis.studentId }, actor))) {
    throw new DefenseError("FORBIDDEN", "Akses Sidang ditolak", 403);
  }
  return hydrate(connection, defense);
}
