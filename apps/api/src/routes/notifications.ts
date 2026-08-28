import { Hono } from "hono";
import { z } from "zod";
import { getDb } from "../db";
import { loadConfig } from "../config";
import { schema } from "@sims/db";
import { Authenticate } from "../middleware/auth";
import {
  getNotification,
  listNotifications,
  markAllRead,
  markRead,
  normalizeLimit,
  unreadCount,
} from "../services/notifications";

export const notificationsRoutes = new Hono();
notificationsRoutes.use("*", Authenticate());

const listQuery = z.object({
  limit: z.coerce.number().optional(),
  offset: z.coerce.number().optional(),
  unread: z
    .string()
    .optional()
    .transform((v) => v === "true" || v === "1"),
});

// Register /unread-count before /:id so Hono doesn't route it into the param.
notificationsRoutes.get("/unread-count", async (c: any) => {
  const u = c.get("user") as any;
  const db = getDb(loadConfig().databaseUrl);
  const count = await unreadCount(db, u.id);
  return c.json({ unread_count: count }, 200);
});

notificationsRoutes.get("/", async (c: any) => {
  const u = c.get("user") as any;
  const parsed = listQuery.safeParse(Object.fromEntries(new URL(c.req.url).searchParams));
  if (!parsed.success) {
    return c.json({ error: { code: "VALIDATION", message: "Invalid query" } }, 400);
  }
  const db = getDb(loadConfig().databaseUrl);
  const notifs = await listNotifications(db, u.id, {
    limit: normalizeLimit(parsed.data.limit),
    offset: Math.max(0, Math.floor(parsed.data.offset ?? 0)),
    unread: parsed.data.unread,
  });
  return c.json(notifs, 200);
});

notificationsRoutes.get("/:id", async (c: any) => {
  const u = c.get("user") as any;
  const id = c.req.param("id");
  const db = getDb(loadConfig().databaseUrl);
  const notif = await getNotification(db, u.id, id);
  if (!notif) return c.json({ error: { code: "NOT_FOUND", message: "Notifikasi tidak ditemukan" } }, 404);
  return c.json(notif, 200);
});

notificationsRoutes.patch("/:id/read", async (c: any) => {
  const u = c.get("user") as any;
  const id = c.req.param("id");
  const db = getDb(loadConfig().databaseUrl);
  const ok = await markRead(db, u.id, id);
  if (!ok) return c.json({ error: { code: "NOT_FOUND", message: "Notifikasi tidak ditemukan" } }, 404);
  return c.json({ ok: true }, 200);
});

notificationsRoutes.post("/read-all", async (c: any) => {
  const u = c.get("user") as any;
  const db = getDb(loadConfig().databaseUrl);
  await markAllRead(db, u.id);
  return c.json({ ok: true }, 200);
});
