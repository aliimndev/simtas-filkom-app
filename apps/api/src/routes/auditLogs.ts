import { Hono } from "hono";
import { getDb } from "../db";
import { loadConfig } from "../config";
import { schema } from "@sims/db";
import { Authenticate } from "../middleware/auth";
import { RequireRole } from "../middleware/rbac";
import { listAuditLogs, getAuditLog } from "../services/auditLogs";

export const auditLogsRoutes = new Hono();
auditLogsRoutes.use("*", Authenticate());
auditLogsRoutes.use("*", RequireRole("ADMIN_FAKULTAS", "KAPRODI"));

auditLogsRoutes.get("/", async (c: any) => {
  const q = c.req.query();
  const page = q.page ? Number(q.page) : 1;
  const perPage = q.per_page ? Number(q.per_page) : 50;
  const result = await listAuditLogs({
    action: q.action,
    entityType: q.entity_type,
    entityId: q.entity_id,
    page,
    perPage,
  });
  return c.json(
    {
      success: true,
      data: result.data,
      meta: {
        page: result.page,
        perPage: result.perPage,
        total: result.total,
        totalPages: result.totalPages,
      },
    },
    200,
  );
});

auditLogsRoutes.get("/:id", async (c: any) => {
  const log = await getAuditLog(c.req.param("id"));
  if (!log) {
    return c.json({ error: { code: "NOT_FOUND", message: "Audit log not found" } }, 404);
  }
  return c.json({ success: true, data: log }, 200);
});
