import { eq, desc, and, sql } from "drizzle-orm";
import { getDb } from "../../db";
import { loadConfig } from "../../config";
import { schema } from "@sims/db";

export class AcademicYearError extends Error {
  code: string;
  status: number;
  constructor(code: string, message: string, status: number) {
    super(message);
    this.code = code;
    this.status = status;
  }
}

export const ErrAcademicYearNotFound = () =>
  new AcademicYearError("NOT_FOUND", "Tahun akademik tidak ditemukan", 404);
export const ErrInvalidSemester = () =>
  new AcademicYearError("VALIDATION", "Semester harus ganjil atau genap", 400);
export const ErrInvalidDateRange = () =>
  new AcademicYearError("VALIDATION", "Tanggal akhir harus setelah tanggal mulai", 400);
export const ErrAcademicYearInUse = () =>
  new AcademicYearError(
    "CONFLICT",
    "Tahun akademik aktif dengan skripsi berjalan tidak dapat diubah",
    409,
  );

export interface AcademicYearRequest {
  name: string;
  semester: string;
  startDate: string; // YYYY-MM-DD
  endDate: string; // YYYY-MM-DD
}

export async function list() {
  const db = getDb(loadConfig().databaseUrl);
  return db.select().from(schema.academicYears).orderBy(desc(schema.academicYears.startDate));
}

export async function get(id: string) {
  const db = getDb(loadConfig().databaseUrl);
  const rows = await db.select().from(schema.academicYears).where(eq(schema.academicYears.id, id));
  if (rows.length === 0) throw ErrAcademicYearNotFound();
  return rows[0];
}

function build(req: AcademicYearRequest) {
  if (req.semester !== "ganjil" && req.semester !== "genap") {
    throw ErrInvalidSemester();
  }
  const start = new Date(req.startDate);
  const end = new Date(req.endDate);
  if (Number.isNaN(start.getTime()) || Number.isNaN(end.getTime())) {
    throw new AcademicYearError("VALIDATION", "Format tanggal tidak valid (YYYY-MM-DD)", 400);
  }
  if (!end.getTime() || end.getTime() <= start.getTime()) {
    throw ErrInvalidDateRange();
  }
  return {
    name: req.name,
    semester: req.semester,
    startDate: req.startDate,
    endDate: req.endDate,
  };
}

export async function create(req: AcademicYearRequest) {
  const db = getDb(loadConfig().databaseUrl);
  const values = build(req);
  const [row] = await db.insert(schema.academicYears).values(values).returning();
  return row;
}

export async function update(id: string, req: AcademicYearRequest) {
  const db = getDb(loadConfig().databaseUrl);
  const existing = await get(id);

  if (existing.isActive) {
    const count = await countActiveTheses(id);
    if (count > 0) throw ErrAcademicYearInUse();
  }

  const values = build(req);
  const [row] = await db
    .update(schema.academicYears)
    .set({
      name: values.name,
      semester: values.semester,
      startDate: values.startDate,
      endDate: values.endDate,
      updatedAt: new Date(),
    })
    .where(eq(schema.academicYears.id, id))
    .returning();
  return row;
}

export async function activate(id: string) {
  const db = getDb(loadConfig().databaseUrl);
  // Existence check (mirrors Go: FindByID then Activate).
  await get(id);
  // Only one active year: deactivate all others, then activate this one.
  await db.transaction(async (tx) => {
    await tx
      .update(schema.academicYears)
      .set({ isActive: false, updatedAt: new Date() })
      .where(eq(schema.academicYears.isActive, true));
    await tx
      .update(schema.academicYears)
      .set({ isActive: true, updatedAt: new Date() })
      .where(eq(schema.academicYears.id, id));
  });
  return get(id);
}

async function countActiveTheses(academicYearId: string): Promise<number> {
  const db = getDb(loadConfig().databaseUrl);
  const rows = await db
    .select({ count: sql<number>`count(*)` })
    .from(schema.theses)
    .where(and(eq(schema.theses.academicYearId, academicYearId), sql`${schema.theses.status} <> 'graduated'`));
  return Number(rows[0]?.count ?? 0);
}
