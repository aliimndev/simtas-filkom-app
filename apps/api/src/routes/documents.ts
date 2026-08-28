import { Hono } from "hono";
import { z } from "zod";
import { loadConfig } from "../config";
import { schema } from "@sims/db";
import { Authenticate } from "../middleware/auth";
import { RequireRole } from "../middleware/rbac";
import { DocumentError, upload, list, getById, review, downloadUrl } from "../services/documents";

export const documentsRoutes = new Hono();

documentsRoutes.use("*", Authenticate());

const actorFrom = (c: any): { userId: string; ipAddress?: string | null; userAgent?: string | null } => {
  const u = c.get("user") as any;
  return {
    userId: u.id,
    ipAddress: c.req.header("x-forwarded-for") ?? null,
    userAgent: c.req.header("user-agent") ?? null,
  };
};

const uploadSchema = z.object({
  thesisId: z.string().uuid(),
  documentType: z.string().min(1),
  chapterNumber: z.number().int().optional().nullable(),
  fileName: z.string().min(1),
  fileUrl: z.string().min(1),
  fileSize: z.number().int().optional().nullable(),
});

const reviewSchema = z.object({
  decision: z.string().min(1),
  notes: z.string().optional().nullable(),
});

function fail(c: any, e: unknown) {
  if (e instanceof DocumentError) {
    return c.json({ error: { code: e.code, message: e.message } }, e.status as any);
  }
  throw e;
}

documentsRoutes.get("/", async (c: any) => {
  try {
    const u = c.get("user") as any;
    const thesisId = c.req.query("thesisId");
    if (!thesisId) return c.json({ error: { code: "VALIDATION", message: "thesisId wajib diisi" } }, 400);
    const { items, total } = await list({
      thesisId,
      documentType: c.req.query("document_type") ?? undefined,
      status: c.req.query("status") ?? undefined,
      userId: u.id,
      role: u.role,
    });
    return c.json({ data: items, total }, 200);
  } catch (e) {
    return fail(c, e);
  }
});

// Register /:id/download before /:id so Hono matches the more specific path first.
documentsRoutes.get("/:id/download", async (c: any) => {
  try {
    const u = c.get("user") as any;
    const result = await downloadUrl(c.req.param("id"), u.id, u.role);
    return c.json(result, 200);
  } catch (e) {
    return fail(c, e);
  }
});

documentsRoutes.get("/:id", async (c: any) => {
  try {
    const u = c.get("user") as any;
    const detail = await getById(c.req.param("id"), u.id, u.role);
    return c.json({ data: detail }, 200);
  } catch (e) {
    return fail(c, e);
  }
});

documentsRoutes.post("/", RequireRole("MAHASISWA"), async (c: any) => {
  try {
    const body = await c.req.json().catch(() => null);
    const parsed = uploadSchema.safeParse(body);
    if (!parsed.success)
      return c.json({ error: { code: "VALIDATION", message: "Invalid body" } }, 400);
    const detail = await upload({ ...parsed.data, actor: actorFrom(c) });
    return c.json({ data: detail }, 201);
  } catch (e) {
    return fail(c, e);
  }
});

documentsRoutes.patch("/:id/review", RequireRole("KAPRODI", "ADMIN_FAKULTAS", "DOSEN_PEMBIMBING"), async (c: any) => {
  try {
    const body = await c.req.json().catch(() => null);
    const parsed = reviewSchema.safeParse(body);
    if (!parsed.success)
      return c.json({ error: { code: "VALIDATION", message: "decision wajib diisi" } }, 400);
    const detail = await review({ id: c.req.param("id"), ...parsed.data, actor: actorFrom(c) });
    return c.json({ data: detail }, 200);
  } catch (e) {
    return fail(c, e);
  }
});
