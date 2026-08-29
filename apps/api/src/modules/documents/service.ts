import { eq, and, desc } from "drizzle-orm";
import { getDb } from "../../db";
import { loadConfig } from "../../config";
import { schema } from "@sims/db";
import { DOC_TYPES, PENDING_REVIEW, ELIGIBLE_STATUSES } from "./constants";
import { audit, canView, getThesis, toDetail, validIp } from "./helpers";
import type { Actor, DocumentDetail, ListFilter, ReviewInput, UploadInput } from "./types";
import { DocumentError } from "./errors";

export { DocumentError };

// Document service business rules are implemented in this TypeScript API;
// storage remains injectable so an object store can be added without changing routes.

export async function upload(input: UploadInput): Promise<DocumentDetail> {
  const documentType = (input.documentType ?? "").trim();
  if (!DOC_TYPES.includes(documentType as any))
    throw new DocumentError("VALIDATION", "document_type tidak valid", 400);

  let chapterNumber: number | null = input.chapterNumber ?? null;
  if (documentType === "draft_chapter" && chapterNumber == null)
    throw new DocumentError("VALIDATION", "chapter_number wajib diisi untuk draft_chapter", 400);
  if (chapterNumber != null && (chapterNumber < 1 || chapterNumber > 5))
    throw new DocumentError("VALIDATION", "chapter_number harus 1-5 untuk draft_chapter", 400);

  if (!input.fileName || !input.fileUrl)
    throw new DocumentError("VALIDATION", "file_name dan file_url wajib diisi", 400);

  const cfg = loadConfig();
  const db = getDb(cfg.databaseUrl);

  const thesis = await getThesis(input.thesisId);
  if (!thesis) throw new DocumentError("NOT_FOUND", "Thesis tidak ditemukan", 404);
  if (!ELIGIBLE_STATUSES.has(thesis.status))
    throw new DocumentError("VALIDATION", "thesis harus berstatus in_progress atau lebih lanjut", 400);
  if (documentType === "final_thesis" && thesis.status !== "defense_done")
    throw new DocumentError("CONFLICT", "final_thesis hanya dapat diunggah setelah Sidang selesai", 409);
  // Only the owning mahasiswa uploads.
  if (thesis.studentId !== input.actor.userId)
    throw new DocumentError("FORBIDDEN", "Bukan pemilik thesis", 403);

  // Version increments per (thesis, document_type) — for draft_chapter, per chapter too.
  const conditions = [
    eq(schema.documents.thesisId, input.thesisId as any),
    eq(schema.documents.documentType, documentType),
  ];
  if (documentType === "draft_chapter" && chapterNumber != null)
    conditions.push(eq(schema.documents.chapterNumber, chapterNumber));
  const latest = await db
    .select()
    .from(schema.documents)
    .where(and(...conditions))
    .orderBy(desc(schema.documents.version))
    .limit(1);
  const version = latest.length ? latest[0].version + 1 : 1;

  const [doc] = await db
    .insert(schema.documents)
    .values({
      thesisId: input.thesisId as any,
      uploadedBy: input.actor.userId as any,
      documentType,
      chapterNumber,
      version,
      fileName: input.fileName,
      fileUrl: input.fileUrl,
      fileSize: input.fileSize ?? null,
      status: PENDING_REVIEW,
    } as any)
    .returning();

  await audit({
    userId: input.actor.userId,
    action: "document_uploaded",
    entityId: doc.id,
    newValue: { document_type: documentType, version, file_name: input.fileName },
    ipAddress: validIp(input.actor.ipAddress),
    userAgent: input.actor.userAgent,
  });

  return toDetail(doc);
}

// List returns the active (latest version per document_type/chapter) documents.
export async function list(filter: ListFilter): Promise<{ items: DocumentDetail[]; total: number }> {
  await canView(filter.thesisId, filter.userId, filter.role);

  const cfg = loadConfig();
  const db = getDb(cfg.databaseUrl);
  const conditions = [eq(schema.documents.thesisId, filter.thesisId as any)];
  if (filter.documentType) conditions.push(eq(schema.documents.documentType, filter.documentType));
  if (filter.status) conditions.push(eq(schema.documents.status, filter.status));

  const rows = await db
    .select()
    .from(schema.documents)
    .where(and(...conditions))
    .orderBy(desc(schema.documents.version));

  // Keep the latest version per (document_type, chapter_number) — mirrors Go DISTINCT ON.
  const latest = new Map<string, any>();
  for (const r of rows) {
    const key = `${r.documentType}::${r.chapterNumber ?? 0}`;
    if (!latest.has(key)) latest.set(key, r);
  }
  const items = Array.from(latest.values())
    .sort((a, b) => new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime())
    .map(toDetail);
  return { items, total: items.length };
}

export async function getById(id: string, userId: string, role: string): Promise<DocumentDetail> {
  const cfg = loadConfig();
  const db = getDb(cfg.databaseUrl);
  const rows = await db.select().from(schema.documents).where(eq(schema.documents.id, id as any));
  const doc = rows[0];
  if (!doc) throw new DocumentError("NOT_FOUND", "Dokumen tidak ditemukan", 404);

  await canView(doc.thesisId, userId, role);
  return toDetail(doc);
}

export async function review(input: ReviewInput): Promise<DocumentDetail> {
  const decision = (input.decision ?? "").toLowerCase();
  // Task mapping: APPROVED -> approved; REJECTED/revision_required -> rejected.
  const status = decision === "approved" ? "approved" : decision === "rejected" || decision === "revision_required" ? "rejected" : null;
  if (!status) throw new DocumentError("VALIDATION", "decision harus approved atau rejected", 400);

  const cfg = loadConfig();
  const db = getDb(cfg.databaseUrl);
  const rows = await db.select().from(schema.documents).where(eq(schema.documents.id, input.id as any));
  const doc = rows[0];
  if (!doc) throw new DocumentError("NOT_FOUND", "Dokumen tidak ditemukan", 404);
  if (doc.status !== PENDING_REVIEW)
    throw new DocumentError("CONFLICT", "Dokumen tidak dalam status menunggu review", 409);

  if (doc.documentType === "final_thesis") {
    if (input.actor.role === "DOSEN_PEMBIMBING") {
      const related = await db
        .select({ id: schema.thesisSupervisors.id })
        .from(schema.thesisSupervisors)
        .where(and(
          eq(schema.thesisSupervisors.thesisId, doc.thesisId),
          eq(schema.thesisSupervisors.supervisorId, input.actor.userId as any),
        ))
        .limit(1);
      if (related.length === 0) throw new DocumentError("FORBIDDEN", "Hanya Dosen Pembimbing terkait yang dapat mereview final_thesis", 403);
    } else if (input.actor.role !== "KAPRODI" && input.actor.role !== "ADMIN_FAKULTAS") {
      throw new DocumentError("FORBIDDEN", "Reviewer final_thesis tidak berwenang", 403);
    }
  }

  const updated = await db
    .update(schema.documents)
    .set({
      status,
      reviewerId: input.actor.userId as any,
      reviewedAt: new Date(),
      reviewerNotes: input.notes ?? null,
    } as any)
    .where(and(eq(schema.documents.id, input.id as any), eq(schema.documents.status, PENDING_REVIEW)))
    .returning();
  if (updated.length === 0)
    throw new DocumentError("CONFLICT", "Dokumen tidak dalam status menunggu review", 409);

  const action = status === "approved" ? "document_approved" : "document_revision";
  await audit({
    userId: input.actor.userId,
    action,
    entityId: doc.id,
    oldValue: { status: PENDING_REVIEW },
    newValue: { status, notes: input.notes ?? null },
    ipAddress: validIp(input.actor.ipAddress),
    userAgent: input.actor.userAgent,
  });

  return toDetail(updated[0]);
}

export async function downloadUrl(id: string, userId: string, role: string): Promise<{ fileUrl: string }> {
  const cfg = loadConfig();
  const db = getDb(cfg.databaseUrl);
  const rows = await db.select().from(schema.documents).where(eq(schema.documents.id, id as any));
  const doc = rows[0];
  if (!doc) throw new DocumentError("NOT_FOUND", "Dokumen tidak ditemukan", 404);

  await canView(doc.thesisId, userId, role);

  // ponytail: no real storage/signing — Go would GeneratePresignedURL here; we return the stored URL.
  await audit({
    userId,
    action: "document_downloaded",
    entityId: doc.id,
    newValue: { document_type: doc.documentType, version: doc.version },
    ipAddress: null,
    userAgent: null,
  });

  return { fileUrl: doc.fileUrl };
}
