import { Hono } from "hono";
import { z } from "zod";
import { Authenticate } from "../middleware/auth";
import { RequireRole } from "../middleware/rbac";
import * as svc from "../services/academicYears";

export const academicYearsRoutes = new Hono();

const bodySchema = z.object({
  name: z.string().min(1),
  semester: z.string().min(1),
  startDate: z.string().min(1),
  endDate: z.string().min(1),
});

academicYearsRoutes.use("*", Authenticate(), RequireRole("ADMIN_FAKULTAS", "KAPRODI"));

// GET /api/v1/academic-years → list (newest start_date first)
academicYearsRoutes.get("/", async (c) => {
  try {
    const years = await svc.list();
    return c.json(years, 200);
  } catch (e: any) {
    return c.json({ error: { code: e.code ?? "INTERNAL", message: e.message } }, e.status ?? 500);
  }
});

// GET /api/v1/academic-years/:id → get one
academicYearsRoutes.get("/:id", async (c) => {
  try {
    const year = await svc.get(c.req.param("id"));
    return c.json(year, 200);
  } catch (e: any) {
    return c.json({ error: { code: e.code ?? "INTERNAL", message: e.message } }, e.status ?? 500);
  }
});

// POST /api/v1/academic-years → create
academicYearsRoutes.post("/", async (c) => {
  const body = await c.req.json().catch(() => null);
  const parsed = bodySchema.safeParse(body);
  if (!parsed.success) {
    return c.json(
      { error: { code: "VALIDATION", message: "Request tidak valid: name, semester, start_date, end_date wajib diisi" } },
      400,
    );
  }
  try {
    const year = await svc.create(parsed.data);
    return c.json(year, 201);
  } catch (e: any) {
    return c.json({ error: { code: e.code ?? "INTERNAL", message: e.message } }, e.status ?? 500);
  }
});

// PUT /api/v1/academic-years/:id → update
academicYearsRoutes.put("/:id", async (c) => {
  const body = await c.req.json().catch(() => null);
  const parsed = bodySchema.safeParse(body);
  if (!parsed.success) {
    return c.json(
      { error: { code: "VALIDATION", message: "Request tidak valid: name, semester, start_date, end_date wajib diisi" } },
      400,
    );
  }
  try {
    const year = await svc.update(c.req.param("id"), parsed.data);
    return c.json(year, 200);
  } catch (e: any) {
    return c.json({ error: { code: e.code ?? "INTERNAL", message: e.message } }, e.status ?? 500);
  }
});

// PATCH /api/v1/academic-years/:id/activate → set active (single active)
academicYearsRoutes.patch("/:id/activate", async (c) => {
  try {
    const year = await svc.activate(c.req.param("id"));
    return c.json({ message: "Tahun akademik berhasil diaktifkan", data: year }, 200);
  } catch (e: any) {
    return c.json({ error: { code: e.code ?? "INTERNAL", message: e.message } }, e.status ?? 500);
  }
});
