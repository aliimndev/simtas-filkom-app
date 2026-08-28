import { and, desc, eq, sql } from "drizzle-orm";
import type { Db } from "@sims/db";
import { schema } from "@sims/db";

export interface ListOpts {
  limit: number;
  offset: number;
  unread?: boolean;
}

export interface NotificationDetail {
  id: string;
  title: string;
  message: string;
  type: string;
  link: string | null;
  isRead: boolean;
  readAt: string | null;
  createdAt: string;
}

// ponytail: Go clamps limit to [1,100] with default 20; mirror it.
export function normalizeLimit(raw: unknown): number {
  let n = typeof raw === "number" ? raw : Number(raw);
  if (!Number.isFinite(n) || n < 1) n = 20;
  if (n > 100) n = 100;
  return Math.floor(n);
}

export async function listNotifications(db: Db, userId: string, opts: ListOpts): Promise<NotificationDetail[]> {
  const where = opts.unread
    ? and(eq(schema.notifications.userId, userId), eq(schema.notifications.isRead, false))
    : eq(schema.notifications.userId, userId);
  const rows: any = await db
    .select()
    .from(schema.notifications)
    .where(where)
    .orderBy(desc(schema.notifications.createdAt))
    .limit(opts.limit)
    .offset(opts.offset);
  return rows.map(toDetail);
}

export async function getNotification(db: Db, userId: string, id: string): Promise<NotificationDetail | null> {
  const rows: any = await db
    .select()
    .from(schema.notifications)
    .where(and(eq(schema.notifications.userId, userId), eq(schema.notifications.id, id)));
  return rows[0] ? toDetail(rows[0]) : null;
}

export async function unreadCount(db: Db, userId: string): Promise<number> {
  const rows: any = await db
    .select({ count: sql<number>`count(*)` })
    .from(schema.notifications)
    .where(and(eq(schema.notifications.userId, userId), eq(schema.notifications.isRead, false)));
  return Number(rows[0]?.count ?? 0);
}

// MarkRead returns false when no owned row matched (404 in Go).
export async function markRead(db: Db, userId: string, id: string): Promise<boolean> {
  const res = await db
    .update(schema.notifications)
    .set({ isRead: true, readAt: new Date() } as any)
    .where(and(eq(schema.notifications.userId, userId), eq(schema.notifications.id, id)));
  return (res.rowCount ?? 0) > 0;
}

export async function markAllRead(db: Db, userId: string): Promise<void> {
  await db
    .update(schema.notifications)
    .set({ isRead: true, readAt: new Date() } as any)
    .where(and(eq(schema.notifications.userId, userId), eq(schema.notifications.isRead, false)));
}

function toDetail(n: any): NotificationDetail {
  return {
    id: n.id,
    title: n.title,
    message: n.message,
    type: n.type,
    link: n.link ?? null,
    isRead: Boolean(n.isRead),
    readAt: n.readAt ? new Date(n.readAt).toISOString() : null,
    createdAt: n.createdAt ? new Date(n.createdAt).toISOString() : "",
  };
}
