import { eq } from "drizzle-orm";
import { getDb } from "../../db";
import { loadConfig } from "../../config";
import { schema } from "@sims/db";

export async function listRoles() {
  const db = getDb(loadConfig().databaseUrl);
  return db.select().from(schema.roles).orderBy(schema.roles.id);
}

export async function getRoleById(id: number) {
  const db = getDb(loadConfig().databaseUrl);
  const rows = await db.select().from(schema.roles).where(eq(schema.roles.id, id));
  return (rows[0] as any) ?? null;
}

export async function createRole(name: string) {
  const db = getDb(loadConfig().databaseUrl);
  const existing = await db.select().from(schema.roles).where(eq(schema.roles.name, name));
  if (existing.length > 0) {
    const err: any = new Error("Role name already exists");
    err.code = "CONFLICT";
    throw err;
  }
  const [row] = await db.insert(schema.roles).values({ name }).returning();
  return row as any;
}

export async function updateRole(id: number, name: string) {
  const db = getDb(loadConfig().databaseUrl);
  const current = await getRoleById(id);
  if (!current) return null;
  const clash = await db.select().from(schema.roles).where(eq(schema.roles.name, name));
  if (clash.length > 0 && (clash[0] as any).id !== id) {
    const err: any = new Error("Role name already exists");
    err.code = "CONFLICT";
    throw err;
  }
  const [row] = await db.update(schema.roles).set({ name }).where(eq(schema.roles.id, id)).returning();
  return row as any;
}
