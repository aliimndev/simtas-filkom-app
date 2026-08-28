import { Hono } from "hono";
import { z } from "zod";
import { Authenticate } from "../middleware/auth";
import { RequireRole } from "../middleware/rbac";
import { UserError, listUsers, getUser, createUser, updateUser, deactivateUser } from "../services/users";

export const usersRoutes = new Hono();
// Note: Go admin user endpoints are admin-scoped; the parity brief lists the five
// roles as the allowed set, so we guard with that set here.
usersRoutes.use(
  "*",
  Authenticate(),
  RequireRole("ADMIN_FAKULTAS", "KAPRODI", "DOSEN_PEMBIMBING", "DOSEN_PENGUJI", "MAHASISWA"),
);

const actorOf = (c: any): { userId: string; ip?: string; ua?: string } => {
  const u = c.get("user");
  return { userId: u.id, ip: c.req.header("x-forwarded-for"), ua: c.req.header("user-agent") };
};

const createSchema = z.object({
  email: z.string().min(1),
  fullName: z.string().min(1),
  role: z.string().min(1),
  nimNidn: z.string().optional(),
  studyProgram: z.string().optional(),
});

const updateSchema = z.object({
  fullName: z.string().optional(),
  nimNidn: z.string().optional(),
  studyProgram: z.string().optional(),
  placeOfBirth: z.string().optional(),
  address: z.string().optional(),
  phone: z.string().optional(),
  birthDate: z.string().optional(),
  faculty: z.string().optional(),
  semester: z.number().int().optional(),
  profilePhotoUrl: z.string().optional(),
});

usersRoutes.get("/", async (c) => {
  const role = c.req.query("role") ?? undefined;
  const isActiveRaw = c.req.query("is_active");
  const isActive = isActiveRaw === undefined ? undefined : isActiveRaw === "true";
  const search = c.req.query("search") ?? undefined;
  const page = Number(c.req.query("page") ?? 1);
  const perPage = Number(c.req.query("per_page") ?? 20);
  const result = await listUsers({ role, isActive, search, page, perPage });
  return c.json(result, 200);
});

usersRoutes.get("/:id", async (c) => {
  const u = await getUser(c.req.param("id"));
  if (!u) return c.json({ error: { code: "NOT_FOUND", message: "User not found" } }, 404);
  return c.json(u, 200);
});

usersRoutes.post("/", async (c) => {
  const body = await c.req.json().catch(() => null);
  const parsed = createSchema.safeParse(body);
  if (!parsed.success) return c.json({ error: { code: "VALIDATION", message: "Invalid request" } }, 400);
  try {
    const u = await createUser(parsed.data, actorOf(c));
    return c.json(u, 201);
  } catch (e: any) {
    if (e?.code === "CONFLICT") return c.json({ error: { code: "CONFLICT", message: e.message } }, 409);
    if (e?.code === "VALIDATION") return c.json({ error: { code: "VALIDATION", message: e.message } }, 400);
    throw e;
  }
});

usersRoutes.put("/:id", async (c) => {
  const id = c.req.param("id");
  const body = await c.req.json().catch(() => null);
  const parsed = updateSchema.safeParse(body);
  if (!parsed.success) return c.json({ error: { code: "VALIDATION", message: "Invalid request" } }, 400);
  const u = await updateUser(id, parsed.data, actorOf(c));
  if (!u) return c.json({ error: { code: "NOT_FOUND", message: "User not found" } }, 404);
  return c.json(u, 200);
});

// Self-profile update (Go: PATCH /users/me). Declared before /:id so it wins the match.
usersRoutes.patch("/me", async (c: any) => {
  const body = await c.req.json().catch(() => null);
  const parsed = updateSchema.safeParse(body);
  if (!parsed.success) return c.json({ error: { code: "VALIDATION", message: "Invalid request" } }, 400);
  const u = await updateUser(c.get("user").id, parsed.data, actorOf(c));
  if (!u) return c.json({ error: { code: "NOT_FOUND", message: "User not found" } }, 404);
  return c.json(u, 200);
});

usersRoutes.patch("/:id/deactivate", async (c) => {
  const id = c.req.param("id");
  try {
    await deactivateUser(id, actorOf(c));
    return c.json({ message: "User deactivated" }, 200);
  } catch (e: any) {
    if (e instanceof UserError) {
      const status = e.code === "FORBIDDEN" ? 403 : 404;
      return c.json({ error: { code: e.code, message: e.message } }, status);
    }
    throw e;
  }
});
