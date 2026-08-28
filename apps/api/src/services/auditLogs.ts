import { and, desc, eq, sql } from "drizzle-orm";
import { getDb } from "../db";
import { loadConfig } from "../config";
import { schema } from "@sims/db";

export interface AuditFilter {
  action?: string;
  entityType?: string;
  entityId?: string;
  page?: number;
  perPage?: number;
}

export interface AuditLogDetail {
  id: string;
  userId: string | null;
  action: string;
  entityType: string | null;
  entityId: string | null;
  oldValue: unknown;
  newValue: unknown;
  ipAddress: string | null;
  userAgent: string | null;
  createdAt: string;
}

function toDetail(row: any): AuditLogDetail {
  return {
    id: row.id,
    userId: row.userId,
    action: row.action,
    entityType: row.entityType,
    entityId: row.entityId,
    oldValue: row.oldValue,
    newValue: row.newValue,
    ipAddress: row.ipAddress,
    userAgent: row.userAgent,
    createdAt: row.createdAt?.toISOString?.() ?? row.createdAt,
  };
}

export async function listAuditLogs(filter: AuditFilter): Promise<{
  data: AuditLogDetail[];
  total: number;
  page: number;
  perPage: number;
  totalPages: number;
}> {
  const cfg = loadConfig();
  const db = getDb(cfg.databaseUrl);

  let page = filter.page ?? 1;
  let perPage = filter.perPage ?? 50;
  if (page < 1) page = 1;
  if (perPage < 1) perPage = 50;
  if (perPage > 200) perPage = 200;

  const conditions: any[] = [];
  if (filter.action) conditions.push(eq(schema.auditLogs.action, filter.action));
  if (filter.entityType) conditions.push(eq(schema.auditLogs.entityType, filter.entityType));
  if (filter.entityId) conditions.push(eq(schema.auditLogs.entityId, filter.entityId));
  const where = conditions.length ? and(...conditions) : undefined;

  const totalRows = await db
    .select({ count: sql<number>`count(*)` })
    .from(schema.auditLogs)
    .where(where);
  const total = Number(totalRows[0]?.count ?? 0);

  const rows = await db
    .select()
    .from(schema.auditLogs)
    .where(where)
    .orderBy(desc(schema.auditLogs.createdAt))
    .limit(perPage)
    .offset((page - 1) * perPage);

  const totalPages = perPage > 0 ? Math.ceil(total / perPage) : 0;
  return { data: rows.map(toDetail), total, page, perPage, totalPages };
}

export async function getAuditLog(id: string): Promise<AuditLogDetail | null> {
  const cfg = loadConfig();
  const db = getDb(cfg.databaseUrl);
  const rows = await db.select().from(schema.auditLogs).where(eq(schema.auditLogs.id, id));
  if (!rows[0]) return null;
  return toDetail(rows[0]);
}
