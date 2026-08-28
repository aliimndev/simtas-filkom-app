import { isIP } from "node:net";
import { eq, and } from "drizzle-orm";
import { getDb } from "../../db";
import { loadConfig } from "../../config";
import { schema } from "@sims/db";
import { DocumentError } from "./errors";
import type { DocumentDetail } from "./types";
import { PENDING_REVIEW } from "./constants";

export function toDetail(d: any): DocumentDetail {
  return {
    id: d.id,
    thesisId: d.thesisId,
    documentType: d.documentType,
    chapterNumber: d.chapterNumber ?? null,
    version: d.version,
    fileName: d.fileName,
    fileUrl: d.fileUrl,
    fileSize: d.fileSize ?? null,
    status: d.status,
    uploadedBy: d.uploadedBy,
    reviewerId: d.reviewerId ?? null,
    reviewerNotes: d.reviewerNotes ?? null,
    reviewedAt: d.reviewedAt ? new Date(d.reviewedAt).toISOString() : null,
    createdAt: new Date(d.createdAt).toISOString(),
    updatedAt: new Date(d.updatedAt).toISOString(),
  };
}

export function validIp(value?: string | null): string | null {
  const candidate = value?.split(",")[0]?.trim();
  return candidate && isIP(candidate) ? candidate : null;
}

export async function audit(params: {
  userId: string | null;
  action: string;
  entityId: string;
  newValue?: Record<string, unknown>;
  oldValue?: Record<string, unknown>;
  ipAddress?: string | null;
  userAgent?: string | null;
}) {
  const cfg = loadConfig();
  const db = getDb(cfg.databaseUrl);
  await db.insert(schema.auditLogs).values({
    userId: params.userId,
    action: params.action,
    entityType: "document",
    entityId: params.entityId,
    newValue: params.newValue ?? null,
    oldValue: params.oldValue ?? null,
    ipAddress: validIp(params.ipAddress),
    userAgent: params.userAgent ?? null,
  } as any);
}

export async function getThesis(thesisId: string) {
  const cfg = loadConfig();
  const db = getDb(cfg.databaseUrl);
  const rows = await db.select().from(schema.theses).where(eq(schema.theses.id, thesisId as any));
  return rows[0] as any | undefined;
}

async function isSupervisor(thesisId: string, userId: string): Promise<boolean> {
  const cfg = loadConfig();
  const db = getDb(cfg.databaseUrl);
  const rows = await db
    .select()
    .from(schema.thesisSupervisors)
    .where(
      and(
        eq(schema.thesisSupervisors.thesisId, thesisId as any),
        eq(schema.thesisSupervisors.supervisorId, userId as any),
      ),
    );
  return rows.length > 0;
}

async function isExaminer(thesisId: string, userId: string): Promise<boolean> {
  const cfg = loadConfig();
  const db = getDb(cfg.databaseUrl);
  const defenses = await db
    .select({ id: schema.thesisDefenses.id })
    .from(schema.thesisDefenses)
    .innerJoin(schema.defenseExaminers, eq(schema.defenseExaminers.defenseId, schema.thesisDefenses.id))
    .where(
      and(
        eq(schema.thesisDefenses.thesisId, thesisId as any),
        eq(schema.defenseExaminers.examinerId, userId as any),
      ),
    );
  if (defenses.length > 0) return true;
  const seminars = await db
    .select({ id: schema.seminars.id })
    .from(schema.seminars)
    .innerJoin(schema.seminarExaminers, eq(schema.seminarExaminers.seminarId, schema.seminars.id))
    .where(
      and(eq(schema.seminars.thesisId, thesisId as any), eq(schema.seminarExaminers.examinerId, userId as any)),
    );
  return seminars.length > 0;
}

// canView mirrors Go usecase.canView (owner + supervisor + examiner + admin + kaprodi).
export async function canView(thesisId: string, userId: string, role: string): Promise<void> {
  switch (role) {
    case "ADMIN_FAKULTAS":
    case "KAPRODI":
      return;
    case "MAHASISWA": {
      const t = await getThesis(thesisId);
      if (!t || t.studentId !== userId) throw new DocumentError("FORBIDDEN", "Akses ditolak", 403);
      return;
    }
    case "DOSEN_PEMBIMBING":
      if (!(await isSupervisor(thesisId, userId)))
        throw new DocumentError("FORBIDDEN", "Akses ditolak", 403);
      return;
    case "DOSEN_PENGUJI":
      if (!(await isExaminer(thesisId, userId)))
        throw new DocumentError("FORBIDDEN", "Akses ditolak", 403);
      return;
    default:
      throw new DocumentError("FORBIDDEN", "Akses ditolak", 403);
  }
}
