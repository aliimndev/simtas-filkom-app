import { and, desc, eq, exists, inArray, or, sql } from "drizzle-orm";
import type { Db } from "@sims/db";
import { schema } from "@sims/db";
import { getDb } from "../../db";
import { loadConfig } from "../../config";
import { notifyGraduation } from "./notifications";
import { ArchiveError } from "./errors";
import type { ArchiveActor, ArchiveInput } from "./types";

function db(): Db {
  return getDb(loadConfig().databaseUrl);
}

function uniqueViolation(error: unknown): boolean {
  return typeof error === "object" && error !== null && "code" in error && (error as { code?: string }).code === "23505";
}

function validateInput(input: ArchiveInput) {
  const fileUrl = input.fileUrl.trim();
  const fileName = input.fileName.trim();
  const abstractId = input.abstractId.trim();
  const abstractEn = input.abstractEn?.trim() || null;
  const keywords = Array.from(new Set((input.keywords ?? []).map((keyword) => keyword.trim()).filter(Boolean)));
  if (!fileUrl || !fileName || !abstractId || !Number.isInteger(input.graduationYear) || input.graduationYear < 2000 || input.graduationYear > 2200) {
    throw new ArchiveError("VALIDATION", "file_url, file_name, abstract_id, dan graduation_year wajib valid", 422);
  }
  if (keywords.length === 0) throw new ArchiveError("VALIDATION", "keywords wajib berisi minimal satu kata kunci", 422);
  return { fileUrl, fileName, abstractId, abstractEn, keywords, graduationYear: input.graduationYear };
}

function toDetail(row: any) {
  return {
    id: row.id,
    thesisId: row.thesisId,
    fileUrl: row.fileUrl,
    fileName: row.fileName,
    abstractId: row.abstractId,
    abstractEn: row.abstractEn ?? null,
    keywords: row.keywords ?? [],
    graduationYear: row.graduationYear,
    archivedBy: row.archivedBy,
    archivedAt: row.archivedAt instanceof Date ? row.archivedAt.toISOString() : new Date(row.archivedAt).toISOString(),
    createdAt: row.createdAt instanceof Date ? row.createdAt.toISOString() : new Date(row.createdAt).toISOString(),
    updatedAt: row.updatedAt instanceof Date ? row.updatedAt.toISOString() : new Date(row.updatedAt).toISOString(),
  };
}

async function canAccess(connection: Pick<Db, "select">, thesisId: string, actor: ArchiveActor) {
  if (actor.role === "KAPRODI" || actor.role === "ADMIN_FAKULTAS") return true;
  if (actor.role === "MAHASISWA") {
    const rows = await connection.select({ id: schema.theses.id })
      .from(schema.theses)
      .where(and(eq(schema.theses.id, thesisId as any), eq(schema.theses.studentId, actor.userId as any)))
      .limit(1);
    return rows.length > 0;
  }
  if (actor.role === "DOSEN_PEMBIMBING") {
    const rows = await connection.select({ id: schema.thesisSupervisors.id })
      .from(schema.thesisSupervisors)
      .where(and(eq(schema.thesisSupervisors.thesisId, thesisId as any), eq(schema.thesisSupervisors.supervisorId, actor.userId as any)))
      .limit(1);
    return rows.length > 0;
  }
  return false;
}

async function archiveAudit(connection: Pick<Db, "insert">, actor: ArchiveActor, action: string, archiveId: string, oldValue: unknown, newValue: unknown) {
  await connection.insert(schema.auditLogs).values({
    userId: actor.userId,
    action,
    entityType: "archive",
    entityId: archiveId,
    oldValue: oldValue as any,
    newValue: newValue as any,
    ipAddress: actor.ipAddress ?? null,
    userAgent: actor.userAgent ?? null,
  } as any);
}

export async function createArchive(thesisId: string, input: ArchiveInput, actor: ArchiveActor) {
  if (actor.role !== "KAPRODI" && actor.role !== "ADMIN_FAKULTAS") {
    throw new ArchiveError("FORBIDDEN", "Hanya Kaprodi atau Admin Fakultas yang dapat membuat Arsip", 403);
  }
  const archive = validateInput(input);
  let result: { archive: any; studentId: string; supervisorIds: string[] };
  try {
    result = await db().transaction(async (tx) => {
      const thesisRows = await tx.select().from(schema.theses).where(eq(schema.theses.id, thesisId as any)).for("update");
      const thesis = thesisRows[0];
      if (!thesis) throw new ArchiveError("NOT_FOUND", "Thesis tidak ditemukan", 404);
      if (thesis.status !== "defense_done") throw new ArchiveError("CONFLICT", "Thesis harus berstatus defense_done sebelum diarsipkan", 409);

      const defenseRows = await tx.select({ status: schema.thesisDefenses.status })
        .from(schema.thesisDefenses)
        .where(eq(schema.thesisDefenses.thesisId, thesisId as any))
        .orderBy(desc(schema.thesisDefenses.createdAt))
        .limit(1);
      if (!defenseRows[0] || !["passed", "revision_required"].includes(defenseRows[0].status)) {
        throw new ArchiveError("CONFLICT", "Sidang terbaru belum menghasilkan outcome yang eligible untuk graduation", 409);
      }

      const finalDocuments = await tx.select().from(schema.documents)
        .where(and(eq(schema.documents.thesisId, thesisId as any), eq(schema.documents.documentType, "final_thesis")))
        .orderBy(desc(schema.documents.version), desc(schema.documents.createdAt))
        .limit(1);
      const finalDocument = finalDocuments[0];
      if (!finalDocument || finalDocument.status !== "approved") {
        throw new ArchiveError("GATE_NOT_MET", "final_thesis terbaru harus berstatus approved", 422);
      }

      const [created] = await tx.insert(schema.thesisArchives).values({
        thesisId: thesisId as any,
        fileUrl: archive.fileUrl,
        fileName: archive.fileName,
        abstractId: archive.abstractId,
        abstractEn: archive.abstractEn,
        keywords: archive.keywords,
        graduationYear: archive.graduationYear,
        archivedBy: actor.userId,
      } as any).returning();
      const updated = await tx.update(schema.theses)
        .set({ status: "graduated", graduatedAt: new Date(), updatedAt: new Date() })
        .where(and(eq(schema.theses.id, thesisId as any), eq(schema.theses.status, "defense_done")))
        .returning({ id: schema.theses.id });
      if (updated.length === 0) throw new ArchiveError("CONFLICT", "Thesis sudah berubah dan tidak dapat diarsipkan", 409);

      const supervisors = await tx.select({ id: schema.thesisSupervisors.supervisorId })
        .from(schema.thesisSupervisors)
        .where(eq(schema.thesisSupervisors.thesisId, thesisId as any));
      await archiveAudit(tx, actor, "archive_created", created.id, { thesis_status: thesis.status }, { thesis_status: "graduated", graduation_year: archive.graduationYear });
      return { archive: created, studentId: thesis.studentId, supervisorIds: supervisors.map((row) => row.id) };
    });
  } catch (error) {
    if (uniqueViolation(error)) throw new ArchiveError("CONFLICT", "Thesis sudah memiliki Arsip", 409);
    throw error;
  }

  await notifyGraduation(thesisId, result.studentId, result.supervisorIds).catch(() => undefined);
  return toDetail(result.archive);
}

async function getArchiveRow(id: string) {
  const rows = await db().select().from(schema.thesisArchives).where(eq(schema.thesisArchives.id, id as any));
  return rows[0] ?? null;
}

export async function getArchive(id: string, actor: ArchiveActor) {
  const connection = db();
  const row = await getArchiveRow(id);
  if (!row) throw new ArchiveError("NOT_FOUND", "Arsip tidak ditemukan", 404);
  if (!(await canAccess(connection, row.thesisId, actor))) throw new ArchiveError("FORBIDDEN", "Akses Arsip ditolak", 403);
  return toDetail(row);
}

export async function listArchives(actor: ArchiveActor) {
  const connection = db();
  let rows: any[];
  if (actor.role === "KAPRODI" || actor.role === "ADMIN_FAKULTAS") {
    rows = await connection.select().from(schema.thesisArchives).orderBy(desc(schema.thesisArchives.archivedAt));
  } else if (actor.role === "MAHASISWA") {
    rows = await connection.select({ archive: schema.thesisArchives })
      .from(schema.thesisArchives)
      .innerJoin(schema.theses, eq(schema.thesisArchives.thesisId, schema.theses.id))
      .where(eq(schema.theses.studentId, actor.userId as any))
      .orderBy(desc(schema.thesisArchives.archivedAt))
      .then((items) => items.map((item) => item.archive));
  } else if (actor.role === "DOSEN_PEMBIMBING") {
    rows = await connection.select({ archive: schema.thesisArchives })
      .from(schema.thesisArchives)
      .innerJoin(schema.thesisSupervisors, eq(schema.thesisArchives.thesisId, schema.thesisSupervisors.thesisId))
      .where(eq(schema.thesisSupervisors.supervisorId, actor.userId as any))
      .orderBy(desc(schema.thesisArchives.archivedAt))
      .then((items) => items.map((item) => item.archive));
  } else {
    rows = [];
  }
  return { data: rows.map(toDetail), total: rows.length };
}
