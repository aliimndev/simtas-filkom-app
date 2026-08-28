import { Hono } from "hono";
import { eq } from "drizzle-orm";
import { getDb } from "../db";
import { loadConfig } from "../config";
import { schema } from "@sims/db";
import { Authenticate } from "../middleware/auth";

export const meRoutes = new Hono();
meRoutes.use("*", Authenticate());
meRoutes.get("/me", async (c: any) => {
  const u = c.get("user") as any;
  const db = getDb(loadConfig().databaseUrl);
  const rows: any = await db.select().from(schema.users).where(eq(schema.users.id, u.id));
  const user = rows[0];
  if (!user) return c.json({ error: { code: "NOT_FOUND", message: "User not found" } }, 404);
  let roleName: string = user.role ?? "";
  if (!roleName && user.roleId) {
    const roleRows: any = await db.select().from(schema.roles).where(eq(schema.roles.id, user.roleId));
    roleName = (roleRows[0]?.name ?? "").toUpperCase();
  }
  return c.json(
    {
      id: user.id,
      email: user.email,
      fullName: user.fullName,
      role: roleName,
      mustChangePassword: Boolean(user.mustChangePassword),
    },
    200,
  );
});
