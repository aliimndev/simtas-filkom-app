import { Hono } from "hono";
import { z } from "zod";
import { getDb } from "../db";
import { loadConfig } from "../config";
import { schema } from "@sims/db";
import { Authenticate } from "../middleware/auth";
import {
  ConsultationError,
  createConsultation,
  listConsultations,
  getConsultation,
  approveConsultation,
  summarizeConsultations,
} from "../services/consultationLogs";

export const consultationLogsRoutes = new Hono();

const createSchema = z.object({
  thesisId: z.string().uuid(),
  consultationDate: z.string().regex(/^\d{4}-\d{2}-\d{2}$/, "YYYY-MM-DD"),
  topicsDiscussed: z.string().min(1),
  notes: z.string().optional(),
  followUp: z.string().optional(),
  attachmentUrl: z.string().optional(),
});

const handle = (c: any, fn: () => Promise<any>) =>
  fn().catch((e) => {
    if (e instanceof ConsultationError) {
      return c.json({ error: { code: e.code, message: e.message } }, e.status);
    }
    throw e;
  });

const actorOf = (c: any) => {
  const u = c.get("user");
  return { id: u.id, role: u.role };
};
const metaOf = (c: any) => ({
  ip: c.req.header("x-forwarded-for") ?? undefined,
  userAgent: c.req.header("user-agent") ?? undefined,
});

consultationLogsRoutes.use("*", Authenticate());

// GET /summary?thesisId= — unpaginated counts (mirrors Go Summary). Registered
// before /:id so it isn't shadowed by the id param route.
consultationLogsRoutes.get("/summary", (c) =>
  handle(c, async () => {
    const thesisId = c.req.query("thesisId");
    if (!thesisId) return c.json({ error: { code: "VALIDATION", message: "thesisId wajib diisi" } }, 400);
    const summary = await summarizeConsultations(thesisId, actorOf(c));
    return c.json({ data: summary }, 200);
  }),
);

// GET /?thesisId=&status=&date_from=&date_to=&page=&per_page=
consultationLogsRoutes.get("/", (c) =>
  handle(c, async () => {
    const thesisId = c.req.query("thesisId");
    if (!thesisId) return c.json({ error: { code: "VALIDATION", message: "thesisId wajib diisi" } }, 400);
    const page = Number(c.req.query("page") ?? 1) || 1;
    const perPage = Number(c.req.query("per_page") ?? 20) || 20;
    const result = await listConsultations(
      thesisId,
      { status: c.req.query("status") ?? undefined, dateFrom: c.req.query("date_from") ?? undefined, dateTo: c.req.query("date_to") ?? undefined, page, perPage },
      actorOf(c),
    );
    return c.json({ data: result.consultations, summary: result.summary, meta: result.meta }, 200);
  }),
);

// GET /:id
consultationLogsRoutes.get("/:id", (c) =>
  handle(c, async () => {
    const thesisId = c.req.query("thesisId");
    if (!thesisId) return c.json({ error: { code: "VALIDATION", message: "thesisId wajib diisi" } }, 400);
    const detail = await getConsultation(thesisId, c.req.param("id"), actorOf(c));
    return c.json({ data: detail }, 200);
  }),
);

// POST /
consultationLogsRoutes.post("/", (c) =>
  handle(c, async () => {
    const body = await c.req.json().catch(() => null);
    const parsed = createSchema.safeParse(body);
    if (!parsed.success) {
      return c.json({ error: { code: "VALIDATION", message: "consultation_date & topics_discussed wajib diisi" } }, 400);
    }
    const detail = await createConsultation(parsed.data, actorOf(c), metaOf(c));
    return c.json({ data: detail }, 201);
  }),
);

// PATCH /:id/approve
consultationLogsRoutes.patch("/:id/approve", (c) =>
  handle(c, async () => {
    const thesisId = c.req.query("thesisId");
    if (!thesisId) return c.json({ error: { code: "VALIDATION", message: "thesisId wajib diisi" } }, 400);
    const detail = await approveConsultation(thesisId, c.req.param("id"), actorOf(c), metaOf(c));
    return c.json({ data: detail }, 200);
  }),
);
