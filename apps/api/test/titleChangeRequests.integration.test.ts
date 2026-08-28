import { beforeAll, afterAll, beforeEach, describe, expect, it } from "bun:test";
import { eq } from "drizzle-orm";
import { createApp } from "../src/app";
import { loadConfig } from "../src/config";
import { getDb } from "../src/db";
import { schema } from "@sims/db";
import { signAccessToken } from "../src/services/token";
import { titleChangeRequestsRoutes } from "../src/routes/titleChangeRequests";

// Use the temporary postgres started for tests (port 5433) when no env is set.
const TEST_DB_URL = process.env.DATABASE_URL ?? "postgres://postgres@localhost:5433/simtas";
if (!process.env.DATABASE_URL) process.env.DATABASE_URL = TEST_DB_URL;

const app = createApp(loadConfig({ ...process.env, NODE_ENV: "test", DATABASE_URL: TEST_DB_URL } as any));
app.route("/api/v1/title-change-requests", titleChangeRequestsRoutes);

beforeAll(async () => {
  await seedFixtures();
});
afterAll(async () => {
  await clearFixtures();
});

// Each request gets a unique IP so the per-IP rate limiters don't accumulate across tests.
let seq = 0;
const req = (path: string, init: Record<string, any> = {}) =>
  app.request(path, {
    ...init,
    headers: { "x-forwarded-for": `tcr-${seq++}`, ...(init.headers ?? {}) },
  });

const auth = (token: string, init: Record<string, any> = {}) => ({
  ...init,
  headers: { authorization: `Bearer ${token}`, "content-type": "application/json", ...(init.headers ?? {}) },
});

const STUDENT_ID = "11111111-1111-1111-1111-111111111111";
const OTHER_STUDENT_ID = "22222222-2222-2222-2222-222222222222";
const KAPRODI_ID = "33333333-3333-3333-3333-333333333333";
const SUPERVISOR_ID = "44444444-4444-4444-4444-444444444444";
const ACADEMIC_YEAR_ID = "aaaaaaaa-1111-1111-1111-111111111111";
const THESIS_ID = "bbbbbbbb-1111-1111-1111-111111111111";
const ORIGINAL_TITLE =
  "Analisis dan implementasi sistem informasi berbasis web untuk manajemen data skripsi mahasiswa di perguruan tinggi";
const NEW_TITLE =
  "Perancangan dan pengembangan aplikasi berbasis mobile untuk pemantauan progres tugas akhir mahasiswa secara real time";

let studentTok = "";
let kaprodiTok = "";
let otherStudentTok = "";

async function roleIdByName(db: any, name: string): Promise<number> {
  const r = await db.select().from(schema.roles).where(eq(schema.roles.name, name));
  return (r[0] as any).id;
}

async function seedFixtures() {
  const db = getDb(TEST_DB_URL);
  const mhsRole = await roleIdByName(db, "mahasiswa");
  const kaprodiRole = await roleIdByName(db, "kaprodi");

  const DUMMY_HASH = "$2a$12$qMPO1EgF0zmpDh4W49ERVOfOxF28jsItEaiKEKZCWOL9NoKX3U7iC";
  for (const [id, roleId, email, name] of [
    [STUDENT_ID, mhsRole, "tcr-student@filkom.ac.id", "TCR Student"],
    [OTHER_STUDENT_ID, mhsRole, "tcr-other@filkom.ac.id", "TCR Other"],
    [KAPRODI_ID, kaprodiRole, "tcr-kaprodi@filkom.ac.id", "TCR Kaprodi"],
    [SUPERVISOR_ID, mhsRole, "tcr-supervisor@filkom.ac.id", "TCR Supervisor"],
  ] as const) {
    await db.delete(schema.users).where(eq(schema.users.id, id));
    await db.insert(schema.users).values({
      id,
      email,
      fullName: name,
      roleId,
      passwordHash: DUMMY_HASH,
      loginAttemptCount: 0,
      tokenVersion: 0,
      mustChangePassword: false,
      isActive: true,
    } as any);
  }

  await db.delete(schema.academicYears).where(eq(schema.academicYears.id, ACADEMIC_YEAR_ID));
  await db.insert(schema.academicYears).values({
    id: ACADEMIC_YEAR_ID,
    name: "2025/2026",
    semester: "ganjil",
    startDate: "2025-09-01",
    endDate: "2026-01-31",
    isActive: false,
  } as any);

  await db.delete(schema.theses).where(eq(schema.theses.id, THESIS_ID));
  await db.insert(schema.theses).values({
    id: THESIS_ID,
    studentId: STUDENT_ID,
    academicYearId: ACADEMIC_YEAR_ID,
    title: ORIGINAL_TITLE,
    thesisType: "skripsi",
    status: "approved",
  } as any);

  await db.delete(schema.thesisSupervisors).where(eq(schema.thesisSupervisors.thesisId, THESIS_ID));
  await db.insert(schema.thesisSupervisors).values({
    id: "cccccccc-1111-1111-1111-111111111111",
    thesisId: THESIS_ID,
    supervisorId: SUPERVISOR_ID,
    assignedBy: SUPERVISOR_ID,
  } as any);

  studentTok = await signAccessToken(STUDENT_ID, "MAHASISWA", 0);
  kaprodiTok = await signAccessToken(KAPRODI_ID, "KAPRODI", 0);
  otherStudentTok = await signAccessToken(OTHER_STUDENT_ID, "MAHASISWA", 0);
}

async function clearFixtures() {
  const db = getDb(TEST_DB_URL);
  await db.delete(schema.titleChangeRequests).where(eq(schema.titleChangeRequests.thesisId, THESIS_ID));
  await db.delete(schema.thesisSupervisors).where(eq(schema.thesisSupervisors.thesisId, THESIS_ID));
  await db.delete(schema.theses).where(eq(schema.theses.id, THESIS_ID));
  await db.delete(schema.academicYears).where(eq(schema.academicYears.id, ACADEMIC_YEAR_ID));
  for (const id of [STUDENT_ID, OTHER_STUDENT_ID, KAPRODI_ID, SUPERVISOR_ID]) {
    await db.delete(schema.users).where(eq(schema.users.id, id));
  }
}

beforeEach(async () => {
  const db = getDb(TEST_DB_URL);
  // Reset pending state and thesis title so each test starts clean.
  await db.delete(schema.titleChangeRequests).where(eq(schema.titleChangeRequests.thesisId, THESIS_ID));
  await db.update(schema.theses).set({ title: ORIGINAL_TITLE } as any).where(eq(schema.theses.id, THESIS_ID));
});

describe("title change requests", () => {
  it("mahasiswa can create a PENDING request", async () => {
    const res = await req(
      "/api/v1/title-change-requests",
      auth(studentTok, {
        method: "POST",
        body: JSON.stringify({ thesisId: THESIS_ID, requestedTitle: NEW_TITLE, reason: "judul kurang bagus" }),
      }),
    );
    expect(res.status).toBe(201);
    const body = (await res.json()) as any;
    expect(body.data.status).toBe("PENDING");
    expect(body.data.previousTitle).toBe(ORIGINAL_TITLE);
    expect(body.data.requestedTitle).toBe(NEW_TITLE);
  });

  it("kaprodi approve updates the thesis title", async () => {
    const create = await req(
      "/api/v1/title-change-requests",
      auth(studentTok, { method: "POST", body: JSON.stringify({ thesisId: THESIS_ID, requestedTitle: NEW_TITLE }) }),
    );
    const created = (await create.json()) as any;

    const res = await req(
      `/api/v1/title-change-requests/${created.data.id}/review`,
      auth(kaprodiTok, { method: "PATCH", body: JSON.stringify({ decision: "APPROVED" }) }),
    );
    expect(res.status).toBe(200);
    const body = (await res.json()) as any;
    expect(body.data.status).toBe("APPROVED");
    expect(body.data.reviewedById).toBe(KAPRODI_ID);

    const db = getDb(TEST_DB_URL);
    const [thesis] = (await db.select().from(schema.theses).where(eq(schema.theses.id, THESIS_ID))) as any;
    expect(thesis.title).toBe(NEW_TITLE);
  });

  it("kaprodi reject records review notes and sets REJECTED", async () => {
    const create = await req(
      "/api/v1/title-change-requests",
      auth(studentTok, { method: "POST", body: JSON.stringify({ thesisId: THESIS_ID, requestedTitle: NEW_TITLE }) }),
    );
    const created = (await create.json()) as any;

    const res = await req(
      `/api/v1/title-change-requests/${created.data.id}/review`,
      auth(kaprodiTok, { method: "PATCH", body: JSON.stringify({ decision: "REJECTED", reviewNotes: "perlu revisi" }) }),
    );
    expect(res.status).toBe(200);
    const body = (await res.json()) as any;
    expect(body.data.status).toBe("REJECTED");
    expect(body.data.reviewNotes).toBe("perlu revisi");
  });

  it("reject without notes returns VALIDATION 400", async () => {
    const create = await req(
      "/api/v1/title-change-requests",
      auth(studentTok, { method: "POST", body: JSON.stringify({ thesisId: THESIS_ID, requestedTitle: NEW_TITLE }) }),
    );
    const created = (await create.json()) as any;
    const res = await req(
      `/api/v1/title-change-requests/${created.data.id}/review`,
      auth(kaprodiTok, { method: "PATCH", body: JSON.stringify({ decision: "REJECTED" }) }),
    );
    expect(res.status).toBe(400);
  });

  it("mahasiswa can cancel a PENDING request", async () => {
    const create = await req(
      "/api/v1/title-change-requests",
      auth(studentTok, { method: "POST", body: JSON.stringify({ thesisId: THESIS_ID, requestedTitle: NEW_TITLE }) }),
    );
    const created = (await create.json()) as any;

    const res = await req(
      `/api/v1/title-change-requests/${created.data.id}/cancel`,
      auth(studentTok, { method: "PATCH" }),
    );
    expect(res.status).toBe(200);
    const body = (await res.json()) as any;
    expect(body.data.status).toBe("CANCELLED");
    expect(body.data.cancelledById).toBe(STUDENT_ID);
  });

  it("forbidden for another student to create on someone else's thesis (403)", async () => {
    const res = await req(
      "/api/v1/title-change-requests",
      auth(otherStudentTok, { method: "POST", body: JSON.stringify({ thesisId: THESIS_ID, requestedTitle: NEW_TITLE }) }),
    );
    expect(res.status).toBe(403);
  });

  it("returns 404 for a non-existent request", async () => {
    const res = await req(
      "/api/v1/title-change-requests/00000000-0000-0000-0000-000000000000",
      auth(studentTok),
    );
    expect(res.status).toBe(404);
  });

  it("reviewing a non-pending request returns CONFLICT 409", async () => {
    const create = await req(
      "/api/v1/title-change-requests",
      auth(studentTok, { method: "POST", body: JSON.stringify({ thesisId: THESIS_ID, requestedTitle: NEW_TITLE }) }),
    );
    const created = (await create.json()) as any;
    await req(
      `/api/v1/title-change-requests/${created.data.id}/cancel`,
      auth(studentTok, { method: "PATCH" }),
    );
    const res = await req(
      `/api/v1/title-change-requests/${created.data.id}/review`,
      auth(kaprodiTok, { method: "PATCH", body: JSON.stringify({ decision: "APPROVED" }) }),
    );
    expect(res.status).toBe(409);
  });
});
