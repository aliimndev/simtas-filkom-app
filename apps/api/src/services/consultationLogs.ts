import { and, desc, eq, sql, count } from "drizzle-orm";
import { getDb } from "../db";
import { loadConfig } from "../config";
import { schema } from "@sims/db";

export interface ConsultationActor {
  id: string;
  role: string;
}

interface ReqMeta {
  ip?: string;
  userAgent?: string;
}

export class ConsultationError extends Error {
  code: string;
  status: number;
  constructor(code: string, message: string, status: number) {
    super(message);
    this.code = code;
    this.status = status;
  }
}

const fail = (code: string, message: string, status: number): never => {
  throw new ConsultationError(code, message, status);
};

const ADMIN_ROLES = ["ADMIN_FAKULTAS", "KAPRODI"];
const ELIGIBLE_STATUSES = [
  "in_progress",
  "seminar_ready",
  "seminar_done",
  "defense_ready",
  "defense_done",
  "graduated",
];

function db() {
  return getDb(loadConfig().databaseUrl);
}

async function getThesis(thesisId: string) {
  const rows = await db().select().from(schema.theses).where(eq(schema.theses.id, thesisId as any));
  return rows[0] ?? null;
}

async function isSupervisor(thesisId: string, userId: string): Promise<boolean> {
  const rows = await db()
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

// canView mirrors Go usecase.canView: admin/kaprodi always; mahasiswa must own;
// dosen_pembimbing must supervise; anything else is forbidden.
async function canView(thesisId: string, actor: ConsultationActor): Promise<void> {
  const role = (actor.role ?? "").toUpperCase();
  if (ADMIN_ROLES.includes(role)) return;
  if (role === "MAHASISWA") {
    const t = await getThesis(thesisId);
    if (t && t.studentId === actor.id) return;
    fail("FORBIDDEN", "Akses ditolak", 403);
  }
  if (role === "DOSEN_PEMBIMBING") {
    if (await isSupervisor(thesisId, actor.id)) return;
    fail("FORBIDDEN", "Akses ditolak", 403);
  }
  fail("FORBIDDEN", "Akses ditolak", 403);
}

function isValidInet(v?: string): boolean {
  if (!v) return false;
  // ponytail: minimal inet validation; tests send a non-IP xff, which we coerce to null.
  return /^(?:\d{1,3}\.){3}\d{1,3}$/.test(v) || v.includes(":");
}

async function audit(params: {
  userId?: string;
  action: string;
  entityId?: string;
  newValue?: unknown;
  ip?: string;
  userAgent?: string;
}) {
  await db()
    .insert(schema.auditLogs)
    .values({
      userId: params.userId ? (params.userId as any) : null,
      action: params.action,
      entityType: "consultation",
      entityId: params.entityId ? (params.entityId as any) : null,
      newValue: params.newValue as any,
      ipAddress: isValidInet(params.ip) ? (params.ip as any) : null,
      userAgent: params.userAgent ?? null,
    } as any);
}

function parseDate(s: string): string {
  if (!/^\d{4}-\d{2}-\d{2}$/.test(s)) fail("VALIDATION", "format tanggal tidak valid, gunakan YYYY-MM-DD", 400);
  const today = new Date().toISOString().slice(0, 10);
  if (s > today) fail("VALIDATION", "tanggal konsultasi tidak boleh di masa depan", 400);
  return s;
}

async function toDetail(rows: any[]) {
  const ids = new Set<string>();
  for (const r of rows) {
    if (r.createdBy) ids.add(r.createdBy);
    if (r.approvedBy) ids.add(r.approvedBy);
  }
  const userMap = new Map<string, any>();
  if (ids.size) {
    const list = Array.from(ids);
    const users = await db()
      .select({ id: schema.users.id, fullName: schema.users.fullName })
      .from(schema.users)
      .where(sql`${schema.users.id} = ANY(${list}::uuid[])`);
    for (const u of users) userMap.set(u.id, u);
  }
  return rows.map((r) => ({
    id: r.id,
    thesisId: r.thesisId,
    createdBy: r.createdBy,
    creator: r.createdBy && userMap.get(r.createdBy) ? { id: r.createdBy, fullName: userMap.get(r.createdBy).fullName } : undefined,
    consultationDate: r.consultationDate,
    topicsDiscussed: r.topicsDiscussed,
    notes: r.notes ?? undefined,
    followUp: r.followUp ?? undefined,
    attachmentUrl: r.attachmentUrl ?? undefined,
    status: r.status,
    approvedBy: r.approvedBy ?? undefined,
    approver:
      r.approvedBy && userMap.get(r.approvedBy) ? { id: r.approvedBy, fullName: userMap.get(r.approvedBy).fullName } : undefined,
    approvedAt: r.approvedAt ?? undefined,
    createdAt: r.createdAt,
    updatedAt: r.updatedAt,
  }));
}

export interface CreateConsultationInput {
  thesisId: string;
  consultationDate: string;
  topicsDiscussed: string;
  notes?: string;
  followUp?: string;
  attachmentUrl?: string;
}

export async function createConsultation(input: CreateConsultationInput, actor: ConsultationActor, meta: ReqMeta) {
  const topicsDiscussed = (input.topicsDiscussed ?? "").trim();
  if (!topicsDiscussed) fail("VALIDATION", "topics_discussed tidak boleh kosong", 400);
  const consultationDate = parseDate((input.consultationDate ?? "").trim());

  const thesis = await getThesis(input.thesisId);
  if (!thesis) fail("NOT_FOUND", "Thesis tidak ditemukan", 404);
  if (!ELIGIBLE_STATUSES.includes(thesis.status)) {
    fail("VALIDATION", "thesis harus berstatus in_progress atau lebih lanjut", 400);
  }

  const owner = thesis.studentId === actor.id;
  const sup = await isSupervisor(input.thesisId, actor.id);
  if (!owner && !sup) fail("FORBIDDEN", "Akses ditolak", 403);

  const [row] = await db()
    .insert(schema.consultationLogs)
    .values({
      thesisId: input.thesisId as any,
      createdBy: actor.id as any,
      consultationDate,
      topicsDiscussed,
      notes: input.notes ?? null,
      followUp: input.followUp ?? null,
      attachmentUrl: input.attachmentUrl ?? null,
      status: "pending",
    } as any)
    .returning();

  await audit({
    userId: actor.id,
    action: "CONSULTATION_CREATED",
    entityId: row.id,
    newValue: { thesis_id: row.thesisId, consultation_date: row.consultationDate, topics_discussed: row.topicsDiscussed, status: row.status },
    ip: meta.ip,
    userAgent: meta.userAgent,
  });

  return (await toDetail([row]))[0];
}

export interface ListConsultationFilter {
  status?: string;
  dateFrom?: string;
  dateTo?: string;
  page: number;
  perPage: number;
}

export async function listConsultations(thesisId: string, filter: ListConsultationFilter, actor: ConsultationActor) {
  await canView(thesisId, actor);

  const page = Math.max(1, filter.page);
  const perPage = Math.min(100, Math.max(1, filter.perPage));

  const base = and(
    eq(schema.consultationLogs.thesisId, thesisId as any),
    filter.status ? eq(schema.consultationLogs.status, filter.status as any) : undefined,
    filter.dateFrom ? sql`${schema.consultationLogs.consultationDate} >= ${filter.dateFrom}` : undefined,
    filter.dateTo ? sql`${schema.consultationLogs.consultationDate} <= ${filter.dateTo}` : undefined,
  );

  const totalRows = await db().select({ value: count() }).from(schema.consultationLogs).where(base);
  const total = Number(totalRows[0]?.value ?? 0);

  const rows = await db()
    .select()
    .from(schema.consultationLogs)
    .where(base)
    .orderBy(desc(schema.consultationLogs.consultationDate), desc(schema.consultationLogs.createdAt))
    .limit(perPage)
    .offset((page - 1) * perPage);

  const totalPages = perPage > 0 ? Math.ceil(total / perPage) : 0;
  const consultations = await toDetail(rows as any[]);
  const summary = await summarize(thesisId);
  return {
    consultations,
    summary,
    meta: { page, perPage, total, totalPages },
  };
}

export async function getConsultation(thesisId: string, id: string, actor: ConsultationActor) {
  await canView(thesisId, actor);
  const rows = await db().select().from(schema.consultationLogs).where(eq(schema.consultationLogs.id, id as any));
  const row = rows[0];
  if (!row || row.thesisId !== thesisId) fail("NOT_FOUND", "Log konsultasi tidak ditemukan", 404);
  return (await toDetail([row]))[0];
}

export async function approveConsultation(thesisId: string, id: string, actor: ConsultationActor, meta: ReqMeta) {
  const rows = await db().select().from(schema.consultationLogs).where(eq(schema.consultationLogs.id, id as any));
  const row = rows[0];
  if (!row || row.thesisId !== thesisId) fail("NOT_FOUND", "Log konsultasi tidak ditemukan", 404);
  if (row.status !== "pending") fail("CONSULTATION_ALREADY_DONE", "Log konsultasi sudah disetujui", 422);

  if (!(await isSupervisor(thesisId, actor.id))) {
    fail("FORBIDDEN", "Hanya dosen pembimbing thesis ini yang dapat menyetujui", 403);
  }

  await db()
    .update(schema.consultationLogs)
    .set({ status: "approved", approvedBy: actor.id as any, approvedAt: new Date(), updatedAt: new Date() } as any)
    .where(eq(schema.consultationLogs.id, id as any));

  const approved = await db().select().from(schema.consultationLogs).where(eq(schema.consultationLogs.id, id as any));
  const approvedCountRows = await db()
    .select({ value: count() })
    .from(schema.consultationLogs)
    .where(and(eq(schema.consultationLogs.thesisId, thesisId as any), eq(schema.consultationLogs.status, "approved")));

  await audit({
    userId: actor.id,
    action: "CONSULTATION_APPROVED",
    entityId: id,
    newValue: { status: "approved", approved_by: actor.id, approved_count: Number(approvedCountRows[0]?.value ?? 0) },
    ip: meta.ip,
    userAgent: meta.userAgent,
  });

  return (await toDetail(approved as any[]))[0];
}

export async function summarizeConsultations(thesisId: string, actor: ConsultationActor) {
  await canView(thesisId, actor);
  return summarize(thesisId);
}

async function summarize(thesisId: string) {
  const all = await db().select().from(schema.consultationLogs).where(eq(schema.consultationLogs.thesisId, thesisId as any));
  const s = {
    totalConsultations: all.length,
    approvedCount: 0,
    pendingCount: 0,
    lastConsultationDate: undefined as string | undefined,
    averageIntervalDays: undefined as number | undefined,
    consultationsThisMonth: 0,
  };
  const now = new Date();
  const dates: number[] = [];
  let last = 0;
  for (const l of all) {
    if (l.status === "approved") s.approvedCount++;
    else s.pendingCount++;
    const t = new Date(l.consultationDate + "T00:00:00Z").getTime();
    if (!Number.isNaN(t)) {
      if (t > last) last = t;
      dates.push(t);
    }
    const d = new Date(l.consultationDate + "T00:00:00Z");
    if (d.getUTCFullYear() === now.getUTCFullYear() && d.getUTCMonth() === now.getUTCMonth()) s.consultationsThisMonth++;
  }
  if (last) s.lastConsultationDate = new Date(last).toISOString().slice(0, 10);
  dates.sort((a, b) => a - b);
  if (dates.length >= 2) {
    let sum = 0;
    for (let i = 1; i < dates.length; i++) sum += Math.floor((dates[i] - dates[i - 1]) / 86_400_000);
    s.averageIntervalDays = Math.floor(sum / (dates.length - 1));
  }
  return s;
}
