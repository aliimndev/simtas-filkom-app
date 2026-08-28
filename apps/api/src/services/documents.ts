import { eq, and, desc, sql, isNull } from "drizzle-orm";
import { getDb } from "../db";
import { loadConfig } from "../config";
import { schema } from "@sims/db";

// Port of Go document module (backend/internal/usecase/document_usecase.go +
// document_repository_impl.go). Faithful on business logic; storage is stubbed
// (ponytail: no object store — callers pass file_url directly).

const DOC_TYPES = [
  "proposal",
  "draft_chapter",
  "seminar_doc",
  "defense_doc",
  "final_thesis",
  "revision_sheet",
  "endorsement_letter",
] as const;

const PENDING_REVIEW = "pending_review";

// Thesis statuses that allow document uploads (Go isDocumentEligible).
const ELIGIBLE_STATUSES = new Set([
  "in_progress",
  "seminar_ready",
  "seminar_done",
  "defense_ready",
  "defense_done",
  "graduated",
]);

export interface Actor {
  userId: string;
  ipAddress?: string | null;
  userAgent?: string | null;
}

export interface DocumentDetail {
  id: string;
  thesisId: string;
  documentType: string;
  chapterNumber: number | null;
  version: number;
  fileName: string;
  fileUrl: string;
  fileSize: number | null;
  status: string;
  uploadedBy: string;
  reviewerId: string | null;
  reviewerNotes: string | null;
  reviewedAt: string | null;
  createdAt: string;
  updatedAt: string;
}

class DocumentError extends Error {
  constructor(
    public code: string,
    message: string,
    public status: number,
  ) {
    super(message);
  }
}

function toDetail(d: any): DocumentDetail {
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

async function audit(params: {
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
    ipAddress: (params.ipAddress ?? null) as any,
    userAgent: params.userAgent ?? null,
  } as any);
}

async function getThesis(thesisId: string) {
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
async function canView(thesisId: string, userId: string, role: string): Promise<void> {
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

export interface UploadInput {
  thesisId: string;
  documentType: string;
  chapterNumber?: number | null;
  fileName: string;
  fileUrl: string;
  fileSize?: number | null;
  actor: Actor;
}

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
    ipAddress: input.actor.ipAddress,
    userAgent: input.actor.userAgent,
  });

  return toDetail(doc);
}

export interface ListFilter {
  thesisId: string;
  documentType?: string;
  status?: string;
  userId: string;
  role: string;
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

export interface ReviewInput {
  id: string;
  decision: string;
  notes?: string | null;
  actor: Actor;
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
    ipAddress: input.actor.ipAddress,
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

export { DocumentError };
