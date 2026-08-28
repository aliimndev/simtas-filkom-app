import { eq } from "drizzle-orm";
import { getDb } from "../src/db";
import { loadConfig } from "../src/config";
import { schema } from "@sims/db";
import { hashPassword } from "../src/services/password";

// Ensure tests use the temp postgres started via /tmp/pgrun (port 5433) when no env is set.
// The CI / local dev default in config is localhost:5432, but our temporary DB is on 5433.
const TEST_DB_URL = process.env.DATABASE_URL ?? "postgres://postgres@localhost:5433/simtas";
if (!process.env.DATABASE_URL) process.env.DATABASE_URL = TEST_DB_URL;

export const TEST_USER = {
  email: "admin@filkom.ac.id",
  password: "Admin123!",
  fullName: "Test Admin",
  role: "ADMIN_FAKULTAS",
};

export async function seedTestUser() {
  const cfg = loadConfig({ ...process.env, NODE_ENV: "test", DATABASE_URL: TEST_DB_URL } as any);
  const db = getDb(cfg.databaseUrl);
  // Resolve role_id from roles table; fallback to 1
  let roleId = 1;
  try {
    const r = await db.select().from(schema.roles).where(eq(schema.roles.name, "admin_fakultas"));
    if (r[0]) roleId = (r[0] as any).id;
  } catch {}
  await db.delete(schema.users).where(eq(schema.users.email, TEST_USER.email));
  // Clean up any leftover families as well via raw delete if needed
  await db.insert(schema.users).values({
    id: "00000000-0000-0000-0000-000000000001" as any,
    email: TEST_USER.email,
    fullName: TEST_USER.fullName,
    roleId,
    passwordHash: await hashPassword(TEST_USER.password),
    loginAttemptCount: 0,
    tokenVersion: 0,
    mustChangePassword: false,
    isActive: true,
  } as any);
}

export async function clearTestUser() {
  const cfg = loadConfig({ ...process.env, NODE_ENV: "test", DATABASE_URL: TEST_DB_URL } as any);
  const db = getDb(cfg.databaseUrl);
  // Clean refresh families first due to FK
  try {
    const users = await db.select().from(schema.users).where(eq(schema.users.email, TEST_USER.email));
    if (users[0]) {
      await db.delete(schema.refreshTokenFamilies).where(eq(schema.refreshTokenFamilies.userId, (users[0] as any).id));
    }
  } catch {}
  await db.delete(schema.users).where(eq(schema.users.email, TEST_USER.email));
}
