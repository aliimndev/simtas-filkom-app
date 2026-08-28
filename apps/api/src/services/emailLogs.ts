import { and, desc, eq, sql } from "drizzle-orm";
import type { Db } from "@sims/db";
import { schema } from "@sims/db";

export interface ListEmailLogsOpts {
  eventType?: string;
  status?: string;
  limit: number;
  offset: number;
}

export interface EmailLogDetail {
  id: string;
  recipientEmail: string;
  eventType: string;
  subject: string | null;
  status: string;
  provider: string;
  errorMessage: string | null;
  attempts: number;
  createdAt: string;
}

// ponytail: Go clamps limit to [1,100] with default 20; mirror it.
export function normalizeLimit(raw: unknown): number {
  let n = typeof raw === "number" ? raw : Number(raw);
  if (!Number.isFinite(n) || n < 1) n = 20;
  if (n > 100) n = 100;
  return Math.floor(n);
}

export async function listEmailLogs(db: Db, opts: ListEmailLogsOpts): Promise<EmailLogDetail[]> {
  const conds = [];
  if (opts.eventType) conds.push(eq(schema.emailLogs.eventType, opts.eventType));
  if (opts.status) conds.push(eq(schema.emailLogs.status, opts.status));
  const where = conds.length ? and(...conds) : undefined;

  const rows: any = await db
    .select()
    .from(schema.emailLogs)
    .where(where)
    .orderBy(desc(schema.emailLogs.createdAt))
    .limit(opts.limit)
    .offset(opts.offset);
  return rows.map(toDetail);
}

export async function countEmailLogs(db: Db, opts: { eventType?: string; status?: string }): Promise<number> {
  const conds = [];
  if (opts.eventType) conds.push(eq(schema.emailLogs.eventType, opts.eventType));
  if (opts.status) conds.push(eq(schema.emailLogs.status, opts.status));
  const where = conds.length ? and(...conds) : undefined;

  const rows: any = await db.select({ count: sql<number>`count(*)` }).from(schema.emailLogs).where(where);
  return Number(rows[0]?.count ?? 0);
}

export async function getEmailLog(db: Db, id: string): Promise<EmailLogDetail | null> {
  const rows: any = await db.select().from(schema.emailLogs).where(eq(schema.emailLogs.id, id));
  return rows[0] ? toDetail(rows[0]) : null;
}

function toDetail(e: any): EmailLogDetail {
  return {
    id: e.id,
    recipientEmail: e.recipientEmail,
    eventType: e.eventType,
    subject: e.subject ?? null,
    status: e.status,
    provider: e.provider,
    errorMessage: e.errorMessage ?? null,
    attempts: Number(e.attempts ?? 0),
    createdAt: e.createdAt ? new Date(e.createdAt).toISOString() : "",
  };
}
