import { Hono } from "hono";
import { z } from "zod";
import { Authenticate } from "../middleware/auth";
import { RequireRole } from "../middleware/rbac";
import {
  submitCreate,
  listRequests,
  getRequest,
  review,
  cancel,
  TitleChangeError,
} from "../services/titleChangeRequests";

export const titleChangeRequestsRoutes = new Hono();

const actorFrom = (c: any): { userId: string; role: string; ipAddress?: string | null; userAgent?: string | null } => {
  const u = c.get("user");
  return {
    userId: u.id,
    role: u.role,
    ipAddress: c.req.header("x-forwarded-for") ?? null,
    userAgent: c.req.header("user-agent") ?? null,
  };
};

const fail = (c: any, e: unknown) => {
  if (e instanceof TitleChangeError) {
    const status = { NOT_FOUND: 404, FORBIDDEN: 403, CONFLICT: 409, VALIDATION: 400 }[e.code] ?? 400;
    return c.json({ error: { code: e.code, message: e.message } }, status);
  }
  throw e;
};

const createSchema = z.object({
  thesisId: z.string().uuid(),
  requestedTitle: z.string().min(1),
  reason: z.string().optional(),
});

const reviewSchema = z.object({
  decision: z.enum(["APPROVED", "REJECTED"]),
  reviewNotes: z.string().optional(),
});

titleChangeRequestsRoutes.use("*", Authenticate());

titleChangeRequestsRoutes.get("/", async (c: any) => {
  const list = await listRequests(actorFrom(c));
  return c.json({ data: list }, 200);
});

titleChangeRequestsRoutes.get("/:id", async (c: any) => {
  try {
    const detail = await getRequest(c.req.param("id"), actorFrom(c));
    return c.json({ data: detail }, 200);
  } catch (e) {
    return fail(c, e);
  }
});

titleChangeRequestsRoutes.post("/", RequireRole("MAHASISWA"), async (c: any) => {
  const body = await c.req.json().catch(() => null);
  const parsed = createSchema.safeParse(body);
  if (!parsed.success) {
    return c.json({ error: { code: "VALIDATION", message: "requested_title wajib diisi" } }, 400);
  }
  try {
    const detail = await submitCreate({ ...parsed.data, actor: actorFrom(c) });
    return c.json({ data: detail }, 201);
  } catch (e) {
    return fail(c, e);
  }
});

titleChangeRequestsRoutes.patch("/:id/review", RequireRole("KAPRODI", "ADMIN_FAKULTAS"), async (c: any) => {
  const body = await c.req.json().catch(() => null);
  const parsed = reviewSchema.safeParse(body);
  if (!parsed.success) {
    return c.json({ error: { code: "VALIDATION", message: "decision harus APPROVED atau REJECTED" } }, 400);
  }
  try {
    const detail = await review({ id: c.req.param("id"), ...parsed.data, actor: actorFrom(c) });
    return c.json({ data: detail }, 200);
  } catch (e) {
    return fail(c, e);
  }
});

titleChangeRequestsRoutes.patch("/:id/cancel", RequireRole("MAHASISWA"), async (c: any) => {
  try {
    const detail = await cancel(c.req.param("id"), actorFrom(c));
    return c.json({ data: detail }, 200);
  } catch (e) {
    return fail(c, e);
  }
});
