import { Hono } from "hono";
import { z } from "zod";
import { eq } from "drizzle-orm";
import { getDb } from "../db";
import { loadConfig } from "../config";
import { schema } from "@sims/db";
import { Authenticate } from "../middleware/auth";
import { RequireRole } from "../middleware/rbac";
import { listRoles, getRoleById, createRole, updateRole } from "../services/roles";

export const rolesRoutes = new Hono();

const createSchema = z.object({ name: z.string().min(1).max(50) });
const updateSchema = z.object({ name: z.string().min(1).max(50) });

rolesRoutes.get("/", Authenticate(), RequireRole("ADMIN_FAKULTAS"), async (c) => {
  const roles = await listRoles();
  return c.json({ roles }, 200);
});

rolesRoutes.get("/:id", Authenticate(), RequireRole("ADMIN_FAKULTAS"), async (c) => {
  const id = Number(c.req.param("id"));
  if (!Number.isInteger(id)) {
    return c.json({ error: { code: "VALIDATION", message: "Invalid id" } }, 400);
  }
  const role = await getRoleById(id);
  if (!role) return c.json({ error: { code: "NOT_FOUND", message: "Role not found" } }, 404);
  return c.json({ role }, 200);
});

rolesRoutes.post("/", Authenticate(), RequireRole("ADMIN_FAKULTAS"), async (c) => {
  const body = await c.req.json().catch(() => null);
  const parsed = createSchema.safeParse(body);
  if (!parsed.success) {
    return c.json({ error: { code: "VALIDATION", message: "name is required" } }, 400);
  }
  try {
    const role = await createRole(parsed.data.name);
    return c.json({ role }, 201);
  } catch (e: any) {
    if (e?.code === "CONFLICT") {
      return c.json({ error: { code: "CONFLICT", message: "Role name already exists" } }, 409);
    }
    throw e;
  }
});

rolesRoutes.put("/:id", Authenticate(), RequireRole("ADMIN_FAKULTAS"), async (c) => {
  const id = Number(c.req.param("id"));
  if (!Number.isInteger(id)) {
    return c.json({ error: { code: "VALIDATION", message: "Invalid id" } }, 400);
  }
  const body = await c.req.json().catch(() => null);
  const parsed = updateSchema.safeParse(body);
  if (!parsed.success) {
    return c.json({ error: { code: "VALIDATION", message: "name is required" } }, 400);
  }
  try {
    const role = await updateRole(id, parsed.data.name);
    if (!role) return c.json({ error: { code: "NOT_FOUND", message: "Role not found" } }, 404);
    return c.json({ role }, 200);
  } catch (e: any) {
    if (e?.code === "CONFLICT") {
      return c.json({ error: { code: "CONFLICT", message: "Role name already exists" } }, 409);
    }
    throw e;
  }
});
