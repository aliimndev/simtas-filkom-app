import { isIP } from "node:net";
import { getDb } from "../db";
import { loadConfig } from "../config";
import { schema } from "@sims/db";
import { eq, and, sql, exists, desc } from "drizzle-orm";

// ── error helper ──
export class ThesesError extends Error {
  constructor(public code: string, message: string, public status: number) {
    super(message);
  }
}

const db_ = () => getDb(loadConfig().databaseUrl);

// ── state machine (mirror backend/pkg/statemachine/thesis_state.go) ──
const VALID_TRANSITIONS: Record<string, string[]> = {
  submitted: ["approved", "rejected", "cancelled"],
  approved: ["in_progress", "cancelled"],
  in_progress: ["seminar_ready", "cancelled"],
  seminar_ready: ["seminar_done"],
  seminar_done: ["defense_ready"],
  defense_ready: ["defense_done"],
  defense_done: ["graduated"],
  rejected: [],
  graduated: [],
  cancelled: [],
};

function canTransition(from: string, to: string): boolean {
  if (to === "cancelled") return from !== "cancelled";
  return (VALID_TRANSITIONS[from] ?? []).includes(to);
}

// ── types ──
export interface ThesisDetail {
  id: string;
  title: string;
  abstract: string | null;
  fieldOfStudy: string | null;
  thesisType: string;
  status: string;
  kaprodiNotes: string | null;
  student: { id: string; fullName: string; nim: string | null };
  supervisors: { id: string; fullName: string }[];
  academicYear: { name: string; semester: string };
  submittedAt: string;
  approvedAt: string | null;
}

export interface Actor {
  userId: string;
  role: string;
  ipAddress: string;
  userAgent: string;
}

async function getDetail(db: any, thesis: any): Promise<ThesisDetail> {
  const student = (await db.select().from(schema.users).where(eq(schema.users.id, thesis.studentId)))[0];
  const ay = (await db.select().from(schema.academicYears).where(eq(schema.academicYears.id, thesis.academicYearId)))[0];
  const supervisors = await db
    .select({ id: schema.users.id, fullName: schema.users.fullName })
    .from(schema.users)
    .innerJoin(schema.thesisSupervisors, eq(schema.thesisSupervisors.supervisorId, schema.users.id))
    .where(eq(schema.thesisSupervisors.thesisId, thesis.id));
  return {
    id: thesis.id,
    title: thesis.title,
    abstract: thesis.abstract,
    fieldOfStudy: thesis.fieldOfStudy,
    thesisType: thesis.thesisType,
    status: thesis.status,
    kaprodiNotes: thesis.kaprodiNotes,
    student: { id: thesis.studentId, fullName: student?.fullName ?? "", nim: student?.nimNidn ?? null },
    supervisors,
    academicYear: { name: ay?.name ?? "", semester: ay?.semester ?? "" },
    submittedAt: thesis.submittedAt?.toISOString?.() ?? String(thesis.submittedAt),
    approvedAt: thesis.approvedAt ? (thesis.approvedAt.toISOString?.() ?? String(thesis.approvedAt)) : null,
  };
}

function validIp(value?: string | null): string | null {
  const candidate = value?.split(",")[0]?.trim();
  return candidate && isIP(candidate) ? candidate : null;
}

async function logAudit(db: any, actor: Actor, action: string, entityId: string, oldValue?: any, newValue?: any) {
  await db.insert(schema.auditLogs).values({
    userId: actor.userId,
    action,
    entityType: "thesis",
    entityId,
    oldValue: oldValue ?? null,
    newValue: newValue ?? null,
    ipAddress: validIp(actor.ipAddress),
    userAgent: actor.userAgent || null,
  });
}

// ── create ──
export async function createThesis(input: any, studentId: string, actor: Actor): Promise<ThesisDetail> {
  const db = db_();
  const title = (input.title ?? "").trim();
  const abstract = (input.abstract ?? "").trim();
  const fieldOfStudy = (input.fieldOfStudy ?? "").trim();
  const thesisType = (input.thesisType ?? "").trim();

  if (!title) throw new ThesesError("VALIDATION", "Title is required", 400);
  if (title.split(/\s+/).length < 10) throw new ThesesError("VALIDATION", "Title must be at least 10 words", 400);
  if (title.length > 500) throw new ThesesError("VALIDATION", "Title max 500 chars", 400);
  if (abstract.split(/\s+/).filter(Boolean).length < 100) throw new ThesesError("VALIDATION", "Abstract must be at least 100 words", 400);
  if (thesisType !== "skripsi" && thesisType !== "tugas_akhir")
    throw new ThesesError("VALIDATION", "thesis_type must be skripsi or tugas_akhir", 400);

  // one active thesis per student
  const active = await db
    .select()
    .from(schema.theses)
    .where(and(eq(schema.theses.studentId, studentId), sql`status NOT IN ('cancelled','graduated')`, sql`deleted_at IS NULL`));
  if (active.length > 0) throw new ThesesError("CONFLICT", "Mahasiswa already has an active thesis", 409);

  const year = await db.select().from(schema.academicYears).where(eq(schema.academicYears.isActive, true));
  if (year.length === 0) throw new ThesesError("VALIDATION", "No active academic year", 400);

  const [thesis] = await db
    .insert(schema.theses)
    .values({
      studentId,
      academicYearId: year[0].id,
      title,
      abstract: abstract || null,
      fieldOfStudy: fieldOfStudy || null,
      thesisType,
      status: "submitted",
    })
    .returning();

  // notify all kaprodi (non-fatal). The inserted row only has studentId, so
  // resolve the student's full name for the notification body.
  const studentRow = (await db
    .select({ fullName: schema.users.fullName })
    .from(schema.users)
    .where(eq(schema.users.id, studentId)))[0];
  const studentName = studentRow?.fullName ?? "";
  const kaprodiRole = (await db.select().from(schema.roles).where(eq(schema.roles.name, "kaprodi")))[0];
  if (kaprodiRole) {
    const kap = await db.select().from(schema.users).where(eq(schema.users.roleId, kaprodiRole.id));
    for (const k of kap) {
      await db.insert(schema.notifications).values({
        userId: k.id,
        title: "Pengajuan Judul Skripsi Baru",
        message: `${studentName} mengajukan judul skripsi baru.`,
        type: "thesis",
        link: `/theses/${thesis.id}`,
      });
    }
  }

  await logAudit(db, actor, "thesis_submitted", thesis.id, null, { title: thesis.title, status: thesis.status });
  return getDetail(db, thesis);
}

// ── list (role-scoped) ──
export async function listTheses(userId: string, role: string, query: any) {
  const db = db_();
  const page = Math.max(1, Number(query.page ?? 1));
  const perPage = Math.min(100, Math.max(1, Number(query.perPage ?? 20)));

  const conditions: any[] = [sql`theses.deleted_at IS NULL`];
  if (role === "MAHASISWA") conditions.push(eq(schema.theses.studentId, userId));
  else if (role === "DOSEN_PEMBIMBING")
    conditions.push(
      exists(
        db.select({ v: sql`1` }).from(schema.thesisSupervisors).where(
          and(eq(schema.thesisSupervisors.thesisId, schema.theses.id), eq(schema.thesisSupervisors.supervisorId, userId)),
        ),
      ),
    );
  // ADMIN_FAKULTAS, KAPRODI => all
  if (query.status) conditions.push(eq(schema.theses.status, query.status));
  if (query.academicYearId) conditions.push(eq(schema.theses.academicYearId, query.academicYearId));
  if (query.fieldOfStudy) conditions.push(eq(schema.theses.fieldOfStudy, query.fieldOfStudy));

  const where = conditions.length ? and(...conditions) : undefined;
  const totalRows = await db.select({ c: sql`count(*)` }).from(schema.theses).where(where);
  const total = Number(totalRows[0]?.c ?? 0);

  const rows = await db
    .select()
    .from(schema.theses)
    .where(where)
    .orderBy(desc(schema.theses.submittedAt))
    .limit(perPage)
    .offset((page - 1) * perPage);

  const data = await Promise.all(rows.map((t: any) => getDetail(db, t)));
  return { data, total, page, perPage };
}

// ── get (role-scoped) ──
export async function getThesis(id: string, userId: string, role: string): Promise<ThesisDetail> {
  const db = db_();
  const rows = await db.select().from(schema.theses).where(and(eq(schema.theses.id, id), sql`deleted_at IS NULL`));
  if (rows.length === 0) throw new ThesesError("NOT_FOUND", "Thesis not found", 404);
  const thesis = rows[0];

  if (!(await canAccessThesis(db, thesis, userId, role)))
    throw new ThesesError("FORBIDDEN", "You do not have access to this thesis", 403);
  return getDetail(db, thesis);
}

async function canAccessThesis(db: any, thesis: any, userId: string, role: string): Promise<boolean> {
  switch (role) {
    case "ADMIN_FAKULTAS":
    case "KAPRODI":
      return true;
    case "MAHASISWA":
      return thesis.studentId === userId;
    case "DOSEN_PEMBIMBING": {
      const sup = await db
        .select({ v: sql`1` })
        .from(schema.thesisSupervisors)
        .where(and(eq(schema.thesisSupervisors.thesisId, thesis.id), eq(schema.thesisSupervisors.supervisorId, userId)));
      return sup.length > 0;
    }
    default:
      return false;
  }
}

// ── review (kaprodi approve/reject) ──
export async function reviewThesis(id: string, decision: string, notes: string, actor: Actor): Promise<ThesisDetail> {
  const db = db_();
  if (decision !== "approved" && decision !== "rejected")
    throw new ThesesError("VALIDATION", "Decision must be approved or rejected", 400);

  const rows = await db.select().from(schema.theses).where(and(eq(schema.theses.id, id), sql`deleted_at IS NULL`));
  if (rows.length === 0) throw new ThesesError("NOT_FOUND", "Thesis not found", 404);
  const thesis = rows[0];

  if (!canTransition(thesis.status, decision))
    throw new ThesesError("VALIDATION", "Invalid status transition", 400);

  const oldStatus = thesis.status;
  const update: any = { status: decision };
  if (notes) update.kaprodiNotes = notes;
  if (decision === "approved") update.approvedAt = new Date();
  update.updatedAt = new Date();

  await db.update(schema.theses).set(update).where(eq(schema.theses.id, id));

  const [updated] = await db.select().from(schema.theses).where(eq(schema.theses.id, id));

  await db.insert(schema.notifications).values({
    userId: updated.studentId,
    title: decision === "approved" ? "Judul Skripsi Disetujui" : "Judul Skripsi Perlu Revisi",
    message:
      decision === "approved"
        ? "Selamat, judul skripsi Anda telah disetujui oleh Kaprodi."
        : "Judul skripsi Anda belum dapat disetujui. Periksa catatan Kaprodi.",
    type: "thesis",
    link: `/theses/${id}`,
  });

  await logAudit(
    db,
    actor,
    decision === "approved" ? "thesis_approved" : "thesis_rejected",
    id,
    { status: oldStatus },
    { status: decision, notes },
  );
  return getDetail(db, updated);
}

// ── assign supervisors ──
export async function assignSupervisors(id: string, supervisorIds: string[], actor: Actor): Promise<ThesisDetail> {
  const db = db_();
  const unique = Array.from(new Set(supervisorIds));
  if (unique.length < 1 || unique.length > 2)
    throw new ThesesError("VALIDATION", "Supervisor count must be 1-2", 400);

  const rows = await db.select().from(schema.theses).where(and(eq(schema.theses.id, id), sql`deleted_at IS NULL`));
  if (rows.length === 0) throw new ThesesError("NOT_FOUND", "Thesis not found", 404);
  const thesis = rows[0];

  if (!canTransition(thesis.status, "in_progress"))
    throw new ThesesError("VALIDATION", "Invalid status transition", 400);

  const lectRole = (await db.select().from(schema.roles).where(eq(schema.roles.name, "dosen_pembimbing")))[0];
  if (lectRole) {
    for (const sid of unique) {
      const sup = await db
        .select()
        .from(schema.users)
        .where(and(eq(schema.users.id, sid), eq(schema.users.roleId, lectRole.id), eq(schema.users.isActive, true)));
      if (sup.length === 0) throw new ThesesError("VALIDATION", "Supervisor must be an active dosen_pembimbing", 400);
    }
  }

  // avoid duplicates: skip supervisor_ids already assigned
  const existing = await db.select().from(schema.thesisSupervisors).where(eq(schema.thesisSupervisors.thesisId, id));
  const existingSet = new Set(existing.map((e: any) => e.supervisorId));
  const toInsert = unique.filter((s) => !existingSet.has(s));

  for (const sid of toInsert) {
    await db.insert(schema.thesisSupervisors).values({ thesisId: id, supervisorId: sid, assignedBy: actor.userId });
  }
  if (thesis.status !== "in_progress") {
    await db.update(schema.theses).set({ status: "in_progress", updatedAt: new Date() }).where(eq(schema.theses.id, id));
  }

  const [updated] = await db.select().from(schema.theses).where(eq(schema.theses.id, id));
  const notifyIds = [updated.studentId, ...unique];
  await db.insert(schema.notifications).values({
    userId: updated.studentId,
    title: "Dosen Pembimbing Ditetapkan",
    message: `Dosen pembimbing telah ditetapkan untuk skripsi "${updated.title}".`,
    type: "thesis",
    link: `/theses/${id}`,
  });
  for (const sid of unique) {
    await db.insert(schema.notifications).values({
      userId: sid,
      title: "Dosen Pembimbing Ditetapkan",
      message: `Anda ditetapkan sebagai pembimbing skripsi "${updated.title}".`,
      type: "thesis",
      link: `/theses/${id}`,
    });
  }

  await logAudit(db, actor, "supervisor_assigned", id, null, { status: "in_progress", supervisor_ids: unique });
  return getDetail(db, updated);
}

// ── soft delete (kaprodi/admin) ──
export async function softDeleteThesis(id: string, actor: Actor): Promise<void> {
  const db = db_();
  const rows = await db.select().from(schema.theses).where(and(eq(schema.theses.id, id), sql`deleted_at IS NULL`));
  if (rows.length === 0) throw new ThesesError("NOT_FOUND", "Thesis not found", 404);
  await db.update(schema.theses).set({ deletedAt: new Date(), updatedAt: new Date() }).where(eq(schema.theses.id, id));
  await logAudit(db, actor, "thesis_deleted", id, { status: rows[0].status }, { deleted: true });
}
