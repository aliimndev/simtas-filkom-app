import { and, desc, eq, ilike, or, sql } from "drizzle-orm";
import { getDb } from "../db";
import { loadConfig } from "../config";
import { schema } from "@sims/db";

const MinKeywords = 3;
const MinAbstractWords = 50;

// ArchiveError carries an API error code + HTTP status so handlers can return
// the exact envelope Go would (e.g. 422 for a non-graduated thesis).
export class ArchiveError extends Error {
  code: string;
  status: number;
  constructor(code: string, message: string, status: number) {
    super(message);
    this.code = code;
    this.status = status;
  }
}

// wordCount mirrors the Go whitespace-split word count for abstract validation.
function wordCount(s: string): number {
  const t = s.trim();
  if (t === "") return 0;
  return t.split(/\s+/).length;
}

export interface CreateArchiveInput {
  thesisId: string;
  fileUrl: string;
  fileName: string;
  abstractId: string;
  abstractEn?: string;
  keywords: string[];
  graduationYear: number;
  archivedBy: string;
}

export interface PatchArchiveInput {
  fileUrl?: string;
  fileName?: string;
  abstractEn?: string;
  keywords?: string[];
  graduationYear?: number;
}

interface JoinedRow {
  archive: any;
  thesis: any;
  student: any;
}

// toDetail maps a joined archive row into the API response shape. file_url is
// returned directly (ponytail: no presigning — just hand back the stored URL).
function toDetail(row: JoinedRow) {
  const a = row.archive;
  const t = row.thesis;
  const s = row.student;
  return {
    id: a.id,
    thesisId: a.thesisId,
    title: t?.title ?? null,
    abstractId: a.abstractId,
    abstractEn: a.abstractEn ?? null,
    keywords: a.keywords ?? [],
    graduationYear: a.graduationYear,
    fileName: a.fileName,
    fileUrl: a.fileUrl, // ponytail: return file_url directly, no presigned URL
    fieldOfStudy: t?.fieldOfStudy ?? null,
    studyProgram: s?.studyProgram ?? null,
    student: s ? { fullName: s.fullName, nim: s.nimNidn } : null,
    archivedBy: a.archivedBy,
    archivedAt: a.archivedAt,
  };
}

function joinedSelect(db: any) {
  return db
    .select({
      archive: schema.thesisArchives,
      thesis: schema.theses,
      student: schema.users,
    })
    .from(schema.thesisArchives)
    .innerJoin(schema.theses, eq(schema.thesisArchives.thesisId, schema.theses.id))
    .innerJoin(schema.users, eq(schema.theses.studentId, schema.users.id));
}

async function findJoinedById(id: string): Promise<JoinedRow | null> {
  const cfg = loadConfig();
  const db = getDb(cfg.databaseUrl);
  const rows = await joinedSelect(db).where(eq(schema.thesisArchives.id, id));
  return rows[0] ?? null;
}

export async function createArchive(input: CreateArchiveInput) {
  const cfg = loadConfig();
  const db = getDb(cfg.databaseUrl);

  const abstractId = input.abstractId.trim();
  if (wordCount(abstractId) < MinAbstractWords) {
    throw new ArchiveError("VALIDATION", "abstract_id minimal 50 kata", 400);
  }
  if (input.keywords.length < MinKeywords) {
    throw new ArchiveError("VALIDATION", "minimal 3 kata kunci", 400);
  }
  const year = input.graduationYear;
  if (!Number.isInteger(year) || year < 2000 || year > new Date().getFullYear()) {
    throw new ArchiveError("VALIDATION", "graduation_year tidak valid", 400);
  }
  if (!input.fileUrl || !input.fileName) {
    throw new ArchiveError("VALIDATION", "file wajib diunggah", 400);
  }

  const thesisRows = await db.select().from(schema.theses).where(eq(schema.theses.id, input.thesisId));
  const thesis = thesisRows[0];
  if (!thesis) {
    throw new ArchiveError("NOT_FOUND", "thesis tidak ditemukan", 404);
  }
  if (thesis.status !== "graduated") {
    throw new ArchiveError("VALIDATION", "hanya thesis berstatus graduated yang dapat diarsipkan", 422);
  }

  const existing = await db
    .select()
    .from(schema.thesisArchives)
    .where(eq(schema.thesisArchives.thesisId, input.thesisId));
  if (existing[0]) {
    throw new ArchiveError("CONFLICT", "thesis ini sudah memiliki arsip", 409);
  }

  const [created] = await db
    .insert(schema.thesisArchives)
    .values({
      thesisId: input.thesisId,
      fileUrl: input.fileUrl,
      fileName: input.fileName,
      abstractId,
      abstractEn: input.abstractEn ?? null,
      keywords: input.keywords,
      graduationYear: year,
      archivedBy: input.archivedBy,
    })
    .returning();

  // Audit: mirror Go's archive.created log (search_vector is DB-generated — skipped).
  await db.insert(schema.auditLogs).values({
    userId: input.archivedBy,
    action: "archive.created",
    entityType: "archive",
    entityId: created.id,
    newValue: { thesis_id: input.thesisId, graduation_year: year, file_name: input.fileName },
    ipAddress: null,
    userAgent: null,
  });

  const joined = await findJoinedById(created.id);
  return toDetail(joined!);
}

export async function listArchives(role: string, userId: string) {
  const cfg = loadConfig();
  const db = getDb(cfg.databaseUrl);
  const conds: any[] = [];
  if (role === "MAHASISWA") conds.push(eq(schema.theses.studentId, userId));
  const rows = await joinedSelect(db)
    .where(conds.length ? and(...conds) : undefined)
    .orderBy(desc(schema.thesisArchives.archivedAt));
  return rows.map(toDetail);
}

export async function searchArchives(role: string, userId: string, q: string) {
  const cfg = loadConfig();
  const db = getDb(cfg.databaseUrl);
  const conds: any[] = [];
  if (role === "MAHASISWA") conds.push(eq(schema.theses.studentId, userId));
  const term = q.trim();
  if (term) {
    const like = `%${term}%`;
    // Mirror Go: search title/abstract via the joined theses, plus keywords array.
    conds.push(
      or(
        ilike(schema.theses.title, like),
        ilike(schema.theses.abstract, like),
        sql`exists (select 1 from unnest(${schema.thesisArchives.keywords}) kw where kw ilike ${like})`,
      ),
    );
  }
  const rows = await joinedSelect(db)
    .where(conds.length ? and(...conds) : undefined)
    .orderBy(desc(schema.thesisArchives.archivedAt));
  return rows.map(toDetail);
}

export async function getArchive(role: string, userId: string, id: string) {
  const joined = await findJoinedById(id);
  if (!joined) throw new ArchiveError("NOT_FOUND", "arsip tidak ditemukan", 404);
  if (role === "MAHASISWA" && joined.thesis.studentId !== userId) {
    throw new ArchiveError("FORBIDDEN", "akses ditolak", 403);
  }
  return toDetail(joined);
}

export async function updateArchive(id: string, patch: PatchArchiveInput) {
  const cfg = loadConfig();
  const db = getDb(cfg.databaseUrl);

  const current = await findJoinedById(id);
  if (!current) throw new ArchiveError("NOT_FOUND", "arsip tidak ditemukan", 404);

  const set: any = {};
  if (patch.fileUrl !== undefined) set.fileUrl = patch.fileUrl;
  if (patch.fileName !== undefined) set.fileName = patch.fileName;
  if (patch.abstractEn !== undefined) set.abstractEn = patch.abstractEn;
  if (patch.keywords !== undefined) set.keywords = patch.keywords;
  if (patch.graduationYear !== undefined) {
    const year = patch.graduationYear;
    if (!Number.isInteger(year) || year < 2000 || year > new Date().getFullYear()) {
      throw new ArchiveError("VALIDATION", "graduation_year tidak valid", 400);
    }
    set.graduationYear = year;
  }
  if (Object.keys(set).length === 0) return toDetail(current);

  await db.update(schema.thesisArchives).set(set).where(eq(schema.thesisArchives.id, id));
  const updated = await findJoinedById(id);
  return toDetail(updated!);
}
