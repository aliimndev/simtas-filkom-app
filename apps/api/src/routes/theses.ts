import { Hono } from "hono";
import { z } from "zod";
import { Authenticate } from "../middleware/auth";
import { RequireRole } from "../middleware/rbac";
import { ThesesError, createThesis, listTheses, getThesis, reviewThesis, assignSupervisors, softDeleteThesis } from "../services/theses";
import type { Actor } from "../services/theses";

export const thesesRoutes = new Hono();

const actorFrom = (c: any): Actor => ({
  userId: c.get("user").id,
  role: c.get("user").role,
  ipAddress: (c.req.header("x-forwarded-for") ?? "").split(",")[0].trim(),
  userAgent: c.req.header("user-agent") ?? "",
});

thesesRoutes.use("*", Authenticate());

// GET / — role-scoped list
thesesRoutes.get("/", async (c: any) => {
  try {
    const u = c.get("user");
    const result = await listTheses(u.id, u.role, c.req.query());
    return c.json(result, 200);
  } catch (e) {
    return err(c, e);
  }
});

// GET /:id — role-scoped detail
thesesRoutes.get("/:id", async (c: any) => {
  try {
    const u = c.get("user");
    const detail = await getThesis(c.req.param("id"), u.id, u.role);
    return c.json(detail, 200);
  } catch (e) {
    return err(c, e);
  }
});

// POST / — mahasiswa creates thesis
const createSchema = z.object({
  title: z.string().min(1),
  abstract: z.string().min(1),
  fieldOfStudy: z.string().optional(),
  thesisType: z.string().min(1),
});
thesesRoutes.post("/", RequireRole("MAHASISWA"), async (c: any) => {
  try {
    const body = await c.req.json().catch(() => null);
    const parsed = createSchema.safeParse(body);
    if (!parsed.success) return c.json({ error: { code: "VALIDATION", message: "Invalid body" } }, 400);
    const u = c.get("user");
    const detail = await createThesis(parsed.data, u.id, actorFrom(c));
    return c.json(detail, 201);
  } catch (e) {
    return err(c, e);
  }
});

// PUT /:id — kaprodi review (approve/reject)
const reviewSchema = z.object({
  decision: z.enum(["approved", "rejected"]),
  notes: z.string().optional(),
});
thesesRoutes.put("/:id", RequireRole("KAPRODI"), async (c) => {
  try {
    const body = await c.req.json().catch(() => null);
    const parsed = reviewSchema.safeParse(body);
    if (!parsed.success) return c.json({ error: { code: "VALIDATION", message: "decision must be approved or rejected" } }, 400);
    const detail = await reviewThesis(c.req.param("id"), parsed.data.decision, parsed.data.notes ?? "", actorFrom(c));
    return c.json(detail, 200);
  } catch (e) {
    return err(c, e);
  }
});

// POST /:id/supervisors — kaprodi/admin assign
const supervisorSchema = z.object({ supervisorIds: z.array(z.string()).min(1).max(2) });
thesesRoutes.post("/:id/supervisors", RequireRole("KAPRODI", "ADMIN_FAKULTAS"), async (c) => {
  try {
    const body = await c.req.json().catch(() => null);
    const parsed = supervisorSchema.safeParse(body);
    if (!parsed.success) return c.json({ error: { code: "VALIDATION", message: "supervisorIds required (1-2)" } }, 400);
    const detail = await assignSupervisors(c.req.param("id"), parsed.data.supervisorIds, actorFrom(c));
    return c.json(detail, 200);
  } catch (e) {
    return err(c, e);
  }
});

// DELETE /:id — kaprodi/admin soft delete
thesesRoutes.delete("/:id", RequireRole("KAPRODI", "ADMIN_FAKULTAS"), async (c) => {
  try {
    await softDeleteThesis(c.req.param("id"), actorFrom(c));
    return c.body(null, 204);
  } catch (e) {
    return err(c, e);
  }
});

function err(c: any, e: unknown) {
  if (e instanceof ThesesError) return c.json({ error: { code: e.code, message: e.message } }, e.status);
  return c.json({ error: { code: "INTERNAL", message: "Internal server error" } }, 500);
}
