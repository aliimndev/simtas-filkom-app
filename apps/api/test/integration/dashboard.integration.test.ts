import { beforeAll, afterAll, describe, expect, it } from "bun:test";
import { Hono } from "hono";
import { eq } from "drizzle-orm";
import { loadConfig } from "../../src/config";
import { getDb } from "../../src/db";
import { schema } from "@sims/db";
import { signAccessToken } from "../../src/modules/auth";
import { dashboardRoutes } from "../../src/modules/dashboard";
import { hashPassword } from "../../src/modules/auth";

const TEST_DB_URL = process.env.DATABASE_URL ?? "postgres://postgres@localhost:5433/simtas";
// Packages call loadConfig() (no env arg), so point process.env at the test DB.
if (!process.env.DATABASE_URL) process.env.DATABASE_URL = TEST_DB_URL;
const cfg = loadConfig({ ...process.env, NODE_ENV: "test", DATABASE_URL: TEST_DB_URL } as any);
const db = getDb(cfg.databaseUrl);

const app = new Hono();
app.route("/api/v1/dashboard", dashboardRoutes);

// Unique IP per request so the global per-IP rate limiter never accumulates across tests.
let seq = 0;
const req = (path: string, init: Record<string, any> = {}) =>
  app.request(path, {
    ...init,
    headers: { "x-forwarded-for": `dash-${seq++}`, ...(init.headers ?? {}) },
  });

// ── fixtures (own, not from helpers.ts) ──
const academicYearId = "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbb1";
const studentId = "cccccccc-cccc-cccc-cccc-ccccccccccc1";
const thesisSubmittedId = "dddddddd-dddd-dddd-dddd-dddddddddd01";
const thesisInProgressId = "dddddddd-dddd-dddd-dddd-dddddddddd02";
const thesisGraduatedId = "dddddddd-dddd-dddd-dddd-dddddddddd03";

async function roleId(name: string): Promise<number> {
  const rows: any = await db.select().from(schema.roles).where(eq(schema.roles.name, name));
  if (!rows[0]) throw new Error(`role ${name} missing in test DB`);
  return rows[0].id;
}

async function seed() {
  const mahasiswaRole = await roleId("mahasiswa");
  const kaprodiRole = await roleId("kaprodi");

  await db.insert(schema.academicYears).values({
    id: academicYearId,
    name: "2024/2025",
    semester: "ganjil",
    startDate: "2024-09-01",
    endDate: "2025-01-31",
    isActive: false,
  } as any);

  await db.insert(schema.users).values({
    id: studentId,
    email: "dash-student@filkom.test",
    fullName: "Dash Student",
    roleId: mahasiswaRole,
    passwordHash: await hashPassword("Student123!"),
    isActive: true,
    mustChangePassword: false,
    loginAttemptCount: 0,
    tokenVersion: 0,
  } as any);

  await db.insert(schema.theses).values([
    {
      id: thesisSubmittedId,
      studentId,
      academicYearId,
      title: "Thesis Submitted",
      thesisType: "skripsi",
      status: "submitted",
    },
    {
      id: thesisInProgressId,
      studentId,
      academicYearId,
      title: "Thesis In Progress",
      thesisType: "skripsi",
      status: "in_progress",
    },
    {
      id: thesisGraduatedId,
      studentId,
      academicYearId,
      title: "Thesis Graduated",
      thesisType: "skripsi",
      status: "graduated",
      submittedAt: new Date("2024-01-01T00:00:00Z"),
      graduatedAt: new Date("2024-03-31T00:00:00Z"), // 90 days => 3.0 months
    },
  ] as any);

  return { kaprodiRole };
}

async function clean() {
  await db.delete(schema.theses).where(eq(schema.theses.id, thesisSubmittedId));
  await db.delete(schema.theses).where(eq(schema.theses.id, thesisInProgressId));
  await db.delete(schema.theses).where(eq(schema.theses.id, thesisGraduatedId));
  await db.delete(schema.users).where(eq(schema.users.id, studentId));
  await db.delete(schema.academicYears).where(eq(schema.academicYears.id, academicYearId));
}

beforeAll(async () => {
  await seed();
});
afterAll(async () => {
  await clean();
});

describe("dashboard summary", () => {
  it("returns expected aggregate numbers for seeded fixtures", async () => {
    const adminTok = await signAccessToken("admin-dash", "ADMIN_FAKULTAS", 0);
    const res = await req("/api/v1/dashboard/summary", {
      headers: { authorization: `Bearer ${adminTok}` },
    });
    expect(res.status).toBe(200);
    const body = (await res.json()) as any;

    expect(body.academic_summary.total_active).toBe(2);
    expect(body.academic_summary.total_graduated).toBe(1);
    expect(body.academic_summary.avg_completion_months).toBeCloseTo(3.0, 1);

    const byStatus = body.by_status as { status: string; label: string; count: number }[];
    const counts = Object.fromEntries(byStatus.map((s) => [s.status, s.count]));
    expect(counts).toEqual({ graduated: 1, in_progress: 1, submitted: 1 });
    expect(byStatus.find((s) => s.status === "submitted")?.label).toBe("Menunggu Review");

    const trend = body.graduation_trend as { month: string; count: number }[];
    expect(trend).toEqual([{ month: "2024-03", count: 1 }]);
  });

  it("forbids a MAHASISWA token with 403", async () => {
    const mhsTok = await signAccessToken(studentId, "MAHASISWA", 0);
    const res = await req("/api/v1/dashboard/summary", {
      headers: { authorization: `Bearer ${mhsTok}` },
    });
    expect(res.status).toBe(403);
  });

  it("rejects a missing token with 401", async () => {
    const res = await req("/api/v1/dashboard/summary");
    expect(res.status).toBe(401);
  });
});
