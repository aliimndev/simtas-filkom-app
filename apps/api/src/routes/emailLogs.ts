import { Hono } from "hono";
import { getDb } from "../db";
import { loadConfig } from "../config";
import { schema } from "@sims/db";
import { Authenticate } from "../middleware/auth";
import { RequireRole } from "../middleware/rbac";
import { countEmailLogs, getEmailLog, listEmailLogs, normalizeLimit } from "../services/emailLogs";

export const emailLogsRoutes = new Hono();

emailLogsRoutes.use("*", Authenticate());
emailLogsRoutes.use("*", RequireRole("ADMIN_FAKULTAS", "KAPRODI"));

emailLogsRoutes.get("/", async (c) => {
  const db = getDb(loadConfig().databaseUrl);

  const rawLimit = c.req.query("limit");
  const rawOffset = c.req.query("offset");
  const limit = normalizeLimit(rawLimit);
  const offset = (() => {
    const n = Number(rawOffset);
    return Number.isFinite(n) && n > 0 ? Math.floor(n) : 0;
  })();

  const eventType = c.req.query("event_type") ?? undefined;
  const status = c.req.query("status") ?? undefined;

  const [items, total] = await Promise.all([
    listEmailLogs(db, { eventType, status, limit, offset }),
    countEmailLogs(db, { eventType, status }),
  ]);

  return c.json({ data: items, total, limit, offset }, 200);
});

emailLogsRoutes.get("/:id", async (c) => {
  const db = getDb(loadConfig().databaseUrl);
  const id = c.req.param("id");
  const log = await getEmailLog(db, id);
  if (!log) return c.json({ error: { code: "NOT_FOUND", message: "Email log not found" } }, 404);
  return c.json({ data: log }, 200);
});
