import { isIP } from "node:net";
import { eq, and, desc } from "drizzle-orm";
import { getDb } from "../db";
import { loadConfig } from "../config";
import { schema } from "@sims/db";
import type { Db } from "@sims/db";

// Status values mirror the Go constants (kept in sync with the DB constraint).
export const TitleChangeStatus = {
  Pending: "PENDING",
  Approved: "APPROVED",
  Rejected: "REJECTED",
  Cancelled: "CANCELLED",
} as const;

export class TitleChangeError extends Error {
  constructor(public code: string, message: string) {
    super(message);
  }
}

const ErrTitleChangeNotFound = () => new TitleChangeError("NOT_FOUND", "permintaan perubahan judul tidak ditemukan");
const ErrTitleChangeForbidden = () => new TitleChangeError("FORBIDDEN", "akses ditolak");
const ErrTitleChangeNotEligible = () =>
  new TitleChangeError("VALIDATION", "perubahan judul hanya dapat diajukan pada status approved atau in_progress");
const ErrNoSupervisorAssigned = () => new TitleChangeError("VALIDATION", "thesis belum memiliki pembimbing aktif");
const ErrPendingTitleChangeExists = () =>
  new TitleChangeError("CONFLICT", "sudah ada permintaan perubahan judul yang sedang diproses");
const ErrTitleChangeNotPending = () => new TitleChangeError("CONFLICT", "permintaan perubahan judul tidak dalam status pending");
const ErrTitleChangeTitleTooShort = () => new TitleChangeError("VALIDATION", "judul minimal 10 kata");
const ErrTitleChangeTitleTooLong = () => new TitleChangeError("VALIDATION", "judul maksimal 500 karakter");
const ErrTitleChangeReviewNotesReq = () => new TitleChangeError("VALIDATION", "catatan penolakan wajib diisi");

export interface Actor {
  userId: string;
  role: string;
  ipAddress?: string | null;
  userAgent?: string | null;
}

function validateTitleChange(title: string): void {
  const words = title.trim().split(/\s+/).filter(Boolean);
  if (words.length < 10) throw ErrTitleChangeTitleTooShort();
  if (title.length > 500) throw ErrTitleChangeTitleTooLong();
}

function validIp(value?: string | null): string | null {
  const candidate = value?.split(",")[0]?.trim();
  return candidate && isIP(candidate) ? candidate : null;
}

async function logAudit(
  db: Db,
  actor: Actor,
  action: string,
  entityId: string,
  oldValue?: Record<string, unknown>,
  newValue?: Record<string, unknown>,
): Promise<void> {
  await db.insert(schema.auditLogs).values({
    userId: actor.userId,
    action,
    entityType: "title_change_request",
    entityId,
    oldValue: oldValue ?? null,
    newValue: newValue ?? null,
    ipAddress: validIp(actor.ipAddress),
    userAgent: actor.userAgent ?? null,
  });
}

function toDetail(r: any) {
  return {
    id: r.id,
    thesisId: r.thesisId,
    previousTitle: r.previousTitle,
    requestedTitle: r.requestedTitle,
    reason: r.reason ?? null,
    status: r.status,
    requestedById: r.requestedById,
    reviewedById: r.reviewedById ?? null,
    reviewedAt: r.reviewedAt ?? null,
    reviewNotes: r.reviewNotes ?? null,
    cancelledById: r.cancelledById ?? null,
    cancelledAt: r.cancelledAt ?? null,
    createdAt: r.createdAt,
    updatedAt: r.updatedAt,
  };
}

// Submit creates a PENDING title change request (Mahasiswa pemilik only).
export async function submitCreate(input: {
  thesisId: string;
  requestedTitle: string;
  reason?: string;
  actor: Actor;
}): Promise<any> {
  const db = getDb(loadConfig().databaseUrl);
  const requestedTitle = input.requestedTitle.trim();
  validateTitleChange(requestedTitle);

  const theses = await db.select().from(schema.theses).where(eq(schema.theses.id, input.thesisId));
  const thesis = theses[0];
  if (!thesis) throw ErrTitleChangeNotFound();

  if (thesis.studentId !== input.actor.userId) throw ErrTitleChangeForbidden();

  if (thesis.status !== "approved" && thesis.status !== "in_progress") throw ErrTitleChangeNotEligible();

  const supers = await db
    .select()
    .from(schema.thesisSupervisors)
    .where(eq(schema.thesisSupervisors.thesisId, input.thesisId));
  if (supers.length === 0) throw ErrNoSupervisorAssigned();

  const pending = await db
    .select()
    .from(schema.titleChangeRequests)
    .where(and(eq(schema.titleChangeRequests.thesisId, input.thesisId), eq(schema.titleChangeRequests.status, "PENDING")));
  if (pending.length > 0) throw ErrPendingTitleChangeExists();

  const [created] = await db
    .insert(schema.titleChangeRequests)
    .values({
      thesisId: input.thesisId,
      requestedById: input.actor.userId,
      previousTitle: thesis.title,
      requestedTitle,
      reason: input.reason ?? null,
      status: TitleChangeStatus.Pending,
    })
    .returning();

  await logAudit(db, input.actor, "TITLE_CHANGE_REQUESTED", created.id, undefined, {
    previous_title: created.previousTitle,
    requested_title: created.requestedTitle,
  });

  return toDetail(created);
}

// List returns title change requests, role-scoped (mahasiswa owns; kaprodi/admin all).
export async function listRequests(actor: Actor): Promise<any[]> {
  const db = getDb(loadConfig().databaseUrl);
  const isStaff = actor.role === "KAPRODI" || actor.role === "ADMIN_FAKULTAS";
  const rows = isStaff
    ? await db.select().from(schema.titleChangeRequests).orderBy(desc(schema.titleChangeRequests.createdAt))
    : await db
        .select()
        .from(schema.titleChangeRequests)
        .where(eq(schema.titleChangeRequests.requestedById, actor.userId))
        .orderBy(desc(schema.titleChangeRequests.createdAt));
  return rows.map(toDetail);
}

// Get returns a single request if the actor may read it.
export async function getRequest(id: string, actor: Actor): Promise<any> {
  const db = getDb(loadConfig().databaseUrl);
  const rows = await db.select().from(schema.titleChangeRequests).where(eq(schema.titleChangeRequests.id, id));
  const r = rows[0];
  if (!r) throw ErrTitleChangeNotFound();

  const isStaff = actor.role === "KAPRODI" || actor.role === "ADMIN_FAKULTAS";
  if (!isStaff && r.requestedById !== actor.userId) throw ErrTitleChangeForbidden();

  return toDetail(r);
}

// Review approves or rejects a PENDING request (Kaprodi/Admin).
export async function review(input: {
  id: string;
  decision: "APPROVED" | "REJECTED";
  reviewNotes?: string;
  actor: Actor;
}): Promise<any> {
  const db = getDb(loadConfig().databaseUrl);
  const rows = await db.select().from(schema.titleChangeRequests).where(eq(schema.titleChangeRequests.id, input.id));
  const r = rows[0];
  if (!r) throw ErrTitleChangeNotFound();
  if (r.status !== TitleChangeStatus.Pending) throw ErrTitleChangeNotPending();

  if (input.decision === "REJECTED") {
    const notes = input.reviewNotes?.trim();
    if (!notes) throw ErrTitleChangeReviewNotesReq();
    const [updated] = await db
      .update(schema.titleChangeRequests)
      .set({
        status: TitleChangeStatus.Rejected,
        reviewedById: input.actor.userId,
        reviewedAt: new Date(),
        reviewNotes: notes,
        updatedAt: new Date(),
      })
      .where(eq(schema.titleChangeRequests.id, input.id))
      .returning();

    await logAudit(db, input.actor, "TITLE_CHANGE_REJECTED", input.id, { status: "PENDING" }, {
      status: "REJECTED",
      review_notes: notes,
    });
    return toDetail(updated);
  }

  // APPROVED: transition the request and update the thesis title atomically.
  const now = new Date();
  const [updated] = await db.transaction(async (tx) => {
    const [u] = await tx
      .update(schema.titleChangeRequests)
      .set({
        status: TitleChangeStatus.Approved,
        reviewedById: input.actor.userId,
        reviewedAt: now,
        reviewNotes: input.reviewNotes ?? null,
        updatedAt: now,
      })
      .where(eq(schema.titleChangeRequests.id, input.id))
      .returning();
    await tx
      .update(schema.theses)
      .set({ title: r.requestedTitle, updatedAt: now } as any)
      .where(eq(schema.theses.id, r.thesisId));
    return [u];
  });

  await logAudit(db, input.actor, "TITLE_CHANGE_APPROVED", input.id, { status: "PENDING", previous_title: r.previousTitle }, {
    status: "APPROVED",
    requested_title: r.requestedTitle,
  });
  await logAudit(db, input.actor, "THESIS_TITLE_UPDATED", r.thesisId, { title: r.previousTitle }, { title: r.requestedTitle });

  return toDetail(updated);
}

// Cancel retracts a PENDING request (Mahasiswa pemilik only).
export async function cancel(id: string, actor: Actor): Promise<any> {
  const db = getDb(loadConfig().databaseUrl);
  const rows = await db.select().from(schema.titleChangeRequests).where(eq(schema.titleChangeRequests.id, id));
  const r = rows[0];
  if (!r) throw ErrTitleChangeNotFound();
  if (r.requestedById !== actor.userId) throw ErrTitleChangeForbidden();
  if (r.status !== TitleChangeStatus.Pending) throw ErrTitleChangeNotPending();

  const [updated] = await db
    .update(schema.titleChangeRequests)
    .set({
      status: TitleChangeStatus.Cancelled,
      cancelledById: actor.userId,
      cancelledAt: new Date(),
      updatedAt: new Date(),
    })
    .where(eq(schema.titleChangeRequests.id, id))
    .returning();

  await logAudit(db, actor, "TITLE_CHANGE_CANCELLED", id, undefined, { status: "CANCELLED" });
  return toDetail(updated);
}
