import { eq, and, ilike, or, sql, desc, count } from "drizzle-orm";
import { randomUUID } from "crypto";
import { getDb } from "../../db";
import { loadConfig } from "../../config";
import { schema } from "@sims/db";
import { hashPassword } from "../auth/password.service";

// ponytail: thin business-logic layer over drizzle; routes call these. No extra layers.

const db = () => getDb(loadConfig().databaseUrl);

export type UserResponse = {
  id: string;
  email: string;
  fullName: string;
  role: string;
  nimNidn: string | null;
  studyProgram: string | null;
  placeOfBirth: string | null;
  address: string | null;
  phone: string | null;
  birthDate: string | null;
  faculty: string | null;
  semester: number | null;
  profilePhotoUrl: string | null;
  isActive: boolean;
  mustChangePassword: boolean;
  createdAt: Date;
  updatedAt: Date;
};

function toResponse(u: any, roleName: string | null): UserResponse {
  return {
    id: u.id,
    email: u.email,
    fullName: u.fullName,
    role: (roleName ?? "").toUpperCase(),
    nimNidn: u.nimNidn ?? null,
    studyProgram: u.studyProgram ?? null,
    placeOfBirth: u.placeOfBirth ?? null,
    address: u.address ?? null,
    phone: u.phone ?? null,
    birthDate: u.birthDate ?? null,
    faculty: u.faculty ?? null,
    semester: u.semester ?? null,
    profilePhotoUrl: u.profilePhotoUrl ?? null,
    isActive: u.isActive,
    mustChangePassword: u.mustChangePassword,
    createdAt: u.createdAt,
    updatedAt: u.updatedAt,
  };
}

export type Actor = { userId: string; ip?: string; ua?: string };

function emailRegex(v: string) {
  return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(v);
}

function randomPassword(n = 12): string {
  const chars = "ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnpqrstuvwxyz23456789";
  let out = "";
  const buf = new Uint8Array(n);
  crypto.getRandomValues(buf);
  for (let i = 0; i < n; i++) out += chars[buf[i] % chars.length];
  return out + "1A"; // satisfy complexity (digit + uppercase)
}

export async function listUsers(filter: {
  role?: string;
  isActive?: boolean;
  search?: string;
  page?: number;
  perPage?: number;
}): Promise<{ data: UserResponse[]; total: number; page: number; perPage: number }> {
  const d = db();
  const page = filter.page && filter.page > 0 ? filter.page : 1;
  let perPage = filter.perPage && filter.perPage > 0 ? filter.perPage : 20;
  if (perPage > 100) perPage = 100;

  const conditions: any[] = [];
  if (filter.role) conditions.push(eq(schema.roles.name, filter.role.toLowerCase()));
  if (filter.isActive !== undefined) conditions.push(eq(schema.users.isActive, filter.isActive));
  if (filter.search) {
    const like = `%${filter.search}%`;
    conditions.push(
      or(
        ilike(schema.users.fullName, like),
        ilike(schema.users.email, like),
        ilike(schema.users.nimNidn, like),
      ),
    );
  }
  const where = conditions.length ? and(...conditions) : undefined;

  const rows = await d
    .select({ user: schema.users, roleName: schema.roles.name })
    .from(schema.users)
    .leftJoin(schema.roles, eq(schema.users.roleId, schema.roles.id))
    .where(where)
    .orderBy(desc(schema.users.createdAt))
    .limit(perPage)
    .offset((page - 1) * perPage);

  const [{ value }] = await d
    .select({ value: count() })
    .from(schema.users)
    .leftJoin(schema.roles, eq(schema.users.roleId, schema.roles.id))
    .where(where);

  return { data: rows.map((r) => toResponse(r.user, r.roleName)), total: Number(value), page, perPage };
}

export async function getUser(id: string): Promise<UserResponse | null> {
  const d = db();
  const rows = await d
    .select({ user: schema.users, roleName: schema.roles.name })
    .from(schema.users)
    .leftJoin(schema.roles, eq(schema.users.roleId, schema.roles.id))
    .where(eq(schema.users.id, id));
  const row = rows[0];
  if (!row) return null;
  return toResponse(row.user, row.roleName);
}

export async function createUser(
  input: { email: string; fullName: string; role: string; nimNidn?: string; studyProgram?: string },
  actor: Actor,
): Promise<UserResponse> {
  const d = db();
  const email = input.email.toLowerCase().trim();
  const fullName = (input.fullName ?? "").trim();
  if (!email || !fullName) throw { code: "VALIDATION", message: "Email and full name are required" };
  if (!emailRegex(email)) throw { code: "VALIDATION", message: "Invalid email format" };

  const roleRows: any = await d.select().from(schema.roles).where(eq(schema.roles.name, input.role.toLowerCase()));
  if (!roleRows[0]) throw { code: "VALIDATION", message: "Invalid role" };
  const roleId = roleRows[0].id;

  const existing: any = await d.select().from(schema.users).where(eq(schema.users.email, email));
  if (existing[0]) throw { code: "CONFLICT", message: "Email already registered" };

  const passwordHash = await hashPassword(randomPassword(12));
  const id = randomUUID();
  await d.insert(schema.users).values({
    id,
    email,
    fullName,
    roleId,
    passwordHash,
    isActive: true,
    mustChangePassword: true,
    loginAttemptCount: 0,
    tokenVersion: 0,
  } as any);

  await d.insert(schema.auditLogs).values({
    userId: actor.userId,
    action: "user.created",
    entityType: "user",
    entityId: id,
    newValue: { email, full_name: fullName, role: input.role },
    ipAddress: null,
    userAgent: actor.ua ?? null,
  } as any);

  return (await getUser(id))!;
}

export async function updateUser(
  id: string,
  input: {
    fullName?: string;
    nimNidn?: string;
    studyProgram?: string;
    placeOfBirth?: string;
    address?: string;
    phone?: string;
    birthDate?: string;
    faculty?: string;
    semester?: number;
    profilePhotoUrl?: string;
  },
  actor: Actor,
): Promise<UserResponse | null> {
  const d = db();
  const current: any = await d.select().from(schema.users).where(eq(schema.users.id, id));
  if (!current[0]) return null;

  const set: any = {};
  // Email & role are immutable on update (Go: UpdateUserRequest omits them).
  for (const k of [
    "fullName",
    "nimNidn",
    "studyProgram",
    "placeOfBirth",
    "address",
    "phone",
    "birthDate",
    "faculty",
    "semester",
    "profilePhotoUrl",
  ] as const) {
    if (input[k] !== undefined) set[k] = input[k];
  }
  if (Object.keys(set).length) {
    await d.update(schema.users).set(set).where(eq(schema.users.id, id));
  }

  const updated = (await getUser(id))!;
  await d.insert(schema.auditLogs).values({
    userId: actor.userId,
    action: "user.updated",
    entityType: "user",
    entityId: id,
    oldValue: { full_name: current[0].fullName, study_program: current[0].studyProgram },
    newValue: { full_name: updated.fullName, study_program: updated.studyProgram },
    ipAddress: null,
    userAgent: actor.ua ?? null,
  } as any);

  return updated;
}

export class UserError extends Error {
  constructor(public code: string, message: string) {
    super(message);
  }
}

export async function deactivateUser(id: string, actor: Actor): Promise<void> {
  const d = db();
  if (id === actor.userId) throw new UserError("FORBIDDEN", "Cannot deactivate your own account");
  const current: any = await d.select().from(schema.users).where(eq(schema.users.id, id));
  if (!current[0]) throw new UserError("NOT_FOUND", "User not found");

  await d.update(schema.users).set({ isActive: false } as any).where(eq(schema.users.id, id));
  await d.update(schema.users).set({ tokenVersion: sql`token_version + 1` } as any).where(eq(schema.users.id, id));

  await d.insert(schema.auditLogs).values({
    userId: actor.userId,
    action: "user.deactivated",
    entityType: "user",
    entityId: id,
    newValue: { is_active: false },
    ipAddress: null,
    userAgent: actor.ua ?? null,
  } as any);
}
