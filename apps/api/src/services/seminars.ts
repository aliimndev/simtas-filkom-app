import { isIP } from "node:net";
import { and, count, desc, eq, exists, sql } from "drizzle-orm";
import type { Db } from "@sims/db";
import { schema } from "@sims/db";
import { getDb } from "../db";
import { loadConfig } from "../config";

const ACTIVE_SEMINAR_STATUSES = ["pending", "scheduled", "passed"] as const;
const STAFF_ROLES = new Set(["ADMIN_FAKULTAS", "KAPRODI"]);

export interface SeminarActor {
  userId: string;
  role: string;
  ipAddress?: string | null;
  userAgent?: string | null;
}

export class SeminarError extends Error {
  constructor(
    public code: string,
    message: string,
    public status: number,
  ) {
    super(message);
  }
}

function db(): Db {
  return getDb(loadConfig().databaseUrl);
}

function isUniqueViolation(error: unknown): boolean {
  return typeof error === "object" && error !== null && "code" in error && (error as { code?: string }).code === "23505";
}

function validIp(value?: string | null): string | null {
  const candidate = value?.split(",")[0]?.trim();
  return candidate && isIP(candidate) ? candidate : null;
}

function iso(value: unknown): string | null {
  if (!value) return null;
  return value instanceof Date ? value.toISOString() : new Date(String(value)).toISOString();
}

async function audit(
  connection: Pick<Db, "insert">,
  actor: SeminarActor,
  action: string,
  entityId: string,
  oldValue: unknown,
  newValue: unknown,
) {
  await connection.insert(schema.auditLogs).values({
    userId: actor.userId,
    action,
    entityType: "seminar",
    entityId,
    oldValue: oldValue as any,
    newValue: newValue as any,
    ipAddress: validIp(actor.ipAddress),
    userAgent: actor.userAgent ?? null,
  } as any);
}

async function approvedLatestSeminarDocument(connection: Pick<Db, "select">, thesisId: string) {
  const rows = await connection
    .select({ status: schema.documents.status })
    .from(schema.documents)
    .where(and(eq(schema.documents.thesisId, thesisId as any), eq(schema.documents.documentType, "seminar_doc")))
    .orderBy(desc(schema.documents.version), desc(schema.documents.createdAt))
    .limit(1);
  return rows[0]?.status === "approved";
}

async function getThesis(connection: Pick<Db, "select">, thesisId: string) {
  const rows = await connection.select().from(schema.theses).where(eq(schema.theses.id, thesisId as any));
  return rows[0] ?? null;
}

async function isSupervisor(connection: Pick<Db, "select">, thesisId: string, userId: string) {
  const rows = await connection
    .select({ id: schema.thesisSupervisors.id })
    .from(schema.thesisSupervisors)
    .where(and(eq(schema.thesisSupervisors.thesisId, thesisId as any), eq(schema.thesisSupervisors.supervisorId, userId as any)))
    .limit(1);
  return rows.length > 0;
}

async function canAccess(connection: Pick<Db, "select">, seminar: any, actor: SeminarActor) {
  if (STAFF_ROLES.has(actor.role)) return true;
  if (actor.role === "MAHASISWA") {
    return seminar.studentId === actor.userId;
  }
  if (actor.role === "DOSEN_PEMBIMBING") {
    return isSupervisor(connection, seminar.thesisId, actor.userId);
  }
  if (actor.role === "DOSEN_PENGUJI") {
    const rows = await connection
      .select({ id: schema.seminarExaminers.id })
      .from(schema.seminarExaminers)
      .where(and(eq(schema.seminarExaminers.seminarId, seminar.id), eq(schema.seminarExaminers.examinerId, actor.userId as any)))
      .limit(1);
    return rows.length > 0;
  }
  return false;
}

async function hydrate(connection: Pick<Db, "select">, row: any) {
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
    .from(schema.seminarExaminers)
    .innerJoin(schema.users, eq(schema.users.id, schema.seminarExaminers.examinerId))
    .where(eq(schema.seminarExaminers.seminarId, row.id));

  const scores = await connection
    .select({
      id: schema.seminarScores.id,
      examinerId: schema.seminarScores.examinerId,
      examinerName: schema.users.fullName,
      componentName: schema.seminarScores.componentName,
      componentWeight: schema.seminarScores.componentWeight,
      score: schema.seminarScores.score,
    })
    .from(schema.seminarScores)
    .innerJoin(schema.users, eq(schema.users.id, schema.seminarScores.examinerId))
    .where(eq(schema.seminarScores.seminarId, row.id));

  return {
    id: row.id,
    thesis: thesis
      ? {
          id: thesis.id,
          title: thesis.title,
          status: thesis.status,
          student: { id: thesis.studentId, fullName: thesis.studentName, nim: thesis.studentNim ?? null },
        }
      : null,
    thesisId: row.thesisId,
    thesisTitle: thesis?.title ?? null,
    thesisStatus: thesis?.status ?? null,
    student: thesis ? { id: thesis.studentId, fullName: thesis.studentName, nim: thesis.studentNim ?? null } : null,
    status: row.status,
    scheduledAt: iso(row.scheduledAt),
    room: row.room ?? null,
    notes: row.notes ?? null,
    finalScore: row.finalScore == null ? null : Number(row.finalScore),
    examiners: examiners.map((examiner: any) => ({ id: examiner.id, fullName: examiner.fullName, nim: examiner.nimNidn ?? null })),
    scores: scores.map((score: any) => ({
      id: score.id,
      examinerId: score.examinerId,
      examiner: { id: score.examinerId, fullName: score.examinerName },
      componentName: score.componentName,
      componentWeight: Number(score.componentWeight),
      score: Number(score.score),
    })),
    createdAt: iso(row.createdAt),
    updatedAt: iso(row.updatedAt),
  };
}

export async function submitSeminar(thesisId: string, actor: SeminarActor) {
  if (actor.role !== "MAHASISWA") throw new SeminarError("FORBIDDEN", "Hanya Mahasiswa yang dapat mengajukan Seminar", 403);

  const connection = db();
  let result: any;
  try {
    result = await connection.transaction(async (tx) => {
      const thesis = await getThesis(tx, thesisId);
      if (!thesis) throw new SeminarError("NOT_FOUND", "Thesis tidak ditemukan", 404);
      if (thesis.studentId !== actor.userId) throw new SeminarError("FORBIDDEN", "Bukan pemilik Thesis", 403);

      const active = await tx
        .select({ id: schema.seminars.id })
        .from(schema.seminars)
        .where(and(eq(schema.seminars.thesisId, thesisId as any), sql`${schema.seminars.status} IN ('pending', 'scheduled', 'passed')`))
        .limit(1);
      if (active.length > 0) throw new SeminarError("CONFLICT", "Thesis ini sudah memiliki Seminar aktif", 409);

      if (thesis.status !== "in_progress") {
        throw new SeminarError("VALIDATION", "Thesis harus berstatus in_progress untuk mengajukan Seminar", 422);
      }

      if (!(await approvedLatestSeminarDocument(tx, thesisId))) {
        throw new SeminarError("GATE_NOT_MET", "Dokumen Seminar terbaru belum disetujui", 422);
      }

      const [seminar] = await tx
        .insert(schema.seminars)
        .values({ thesisId: thesisId as any, status: "pending" })
        .returning();
      const updated = await tx
        .update(schema.theses)
        .set({ status: "seminar_ready", updatedAt: new Date() })
        .where(and(eq(schema.theses.id, thesisId as any), eq(schema.theses.status, "in_progress")))
        .returning({ id: schema.theses.id });
      if (updated.length === 0) throw new SeminarError("CONFLICT", "Status Thesis berubah, ulangi pengajuan", 409);

      await audit(tx, actor, "seminar_submitted", seminar.id, { status: thesis.status }, { status: "pending", thesisStatus: "seminar_ready" });
      return seminar;
    });
  } catch (error) {
    if (isUniqueViolation(error)) throw new SeminarError("CONFLICT", "Thesis ini sudah memiliki Seminar aktif", 409);
    throw error;
  }

  await notifyKaprodi(result.id, thesisId, actor).catch(() => undefined);
  return getSeminar(result.id, actor);
}

async function notifyKaprodi(seminarId: string, thesisId: string, actor: SeminarActor) {
  const connection = db();
  const roles = await connection.select({ id: schema.roles.id }).from(schema.roles).where(eq(schema.roles.name, "kaprodi"));
  if (!roles[0]) return;
  const kaprodi = await connection.select({ id: schema.users.id }).from(schema.users).where(eq(schema.users.roleId, roles[0].id));
  if (kaprodi.length === 0) return;
  await connection.insert(schema.notifications).values(
    kaprodi.map((user: any) => ({
      userId: user.id,
      title: "Pengajuan Seminar Baru",
      message: "Mahasiswa mengajukan Seminar dan menunggu penjadwalan.",
      type: "seminar",
      link: `/dashboard/seminars/${seminarId}`,
    })) as any,
  );
}

export interface SeminarListFilter {
  status?: string;
  page: number;
  perPage: number;
}

function scopedWhere(actor: SeminarActor) {
  if (STAFF_ROLES.has(actor.role)) return undefined;
  if (actor.role === "MAHASISWA") {
    return exists(
      db().select({ one: sql`1` }).from(schema.theses).where(and(eq(schema.theses.id, schema.seminars.thesisId), eq(schema.theses.studentId, actor.userId as any))),
    );
  }
  if (actor.role === "DOSEN_PEMBIMBING") {
    return exists(
      db().select({ one: sql`1` }).from(schema.thesisSupervisors).where(and(eq(schema.thesisSupervisors.thesisId, schema.seminars.thesisId), eq(schema.thesisSupervisors.supervisorId, actor.userId as any))),
    );
  }
  if (actor.role === "DOSEN_PENGUJI") {
    return exists(
      db().select({ one: sql`1` }).from(schema.seminarExaminers).where(and(eq(schema.seminarExaminers.seminarId, schema.seminars.id), eq(schema.seminarExaminers.examinerId, actor.userId as any))),
    );
  }
  return sql`false`;
}

export async function listSeminars(filter: SeminarListFilter, actor: SeminarActor) {
  const connection = db();
  const page = Math.max(1, Math.floor(filter.page || 1));
  const perPage = Math.min(100, Math.max(1, Math.floor(filter.perPage || 20)));
  const where = and(scopedWhere(actor), filter.status ? eq(schema.seminars.status, filter.status) : undefined);
  const totalRows = await connection.select({ value: count() }).from(schema.seminars).where(where);
  const total = Number(totalRows[0]?.value ?? 0);
  const rows = await connection
    .select()
    .from(schema.seminars)
    .where(where)
    .orderBy(desc(schema.seminars.createdAt))
    .limit(perPage)
    .offset((page - 1) * perPage);
  return {
    data: await Promise.all(rows.map((row: any) => hydrate(connection, row))),
    meta: { page, perPage, total, totalPages: Math.ceil(total / perPage) },
  };
}

export async function getSeminar(id: string, actor: SeminarActor) {
  const connection = db();
  const rows = await connection.select().from(schema.seminars).where(eq(schema.seminars.id, id as any));
  const seminar = rows[0];
  if (!seminar) throw new SeminarError("NOT_FOUND", "Seminar tidak ditemukan", 404);
  const thesis = await getThesis(connection, seminar.thesisId);
  if (!thesis) throw new SeminarError("NOT_FOUND", "Thesis tidak ditemukan", 404);
  const access = await canAccess(connection, { ...seminar, studentId: thesis.studentId }, actor);
  if (!access) throw new SeminarError("FORBIDDEN", "Akses Seminar ditolak", 403);
  return hydrate(connection, seminar);
}

export { ACTIVE_SEMINAR_STATUSES };
