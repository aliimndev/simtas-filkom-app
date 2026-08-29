import { afterAll, beforeAll, describe, expect, it } from "bun:test";
import { and, eq, inArray } from "drizzle-orm";
import { createApp } from "../../src/app";
import { loadConfig } from "../../src/config";
import { getDb } from "../../src/db";
import { signAccessToken } from "../../src/modules/auth";
import { schema } from "@sims/db";

const TEST_DB_URL = process.env.DATABASE_URL ?? "postgres://postgres@localhost:5433/simtas";
const cfg = loadConfig({ ...process.env, NODE_ENV: "test", DATABASE_URL: TEST_DB_URL } as any);
const app = createApp(cfg);
const db = getDb(cfg.databaseUrl);
const DUMMY_HASH = "$2a$12$qMPO1EgF0zmpDh4W49ERVOfOxF28jsItEaiKEKZCWOL9NoKX3U7iC";
const unique = (prefix: string) => `${prefix}-${crypto.randomUUID()}@filkom.ac.id`;

let studentId = "";
let otherStudentId = "";
let supervisorId = "";
let unrelatedSupervisorId = "";
let kaprodiId = "";
let adminId = "";
let academicYearId = "";
let passedThesisId = "";
let revisionThesisId = "";
let pendingThesisId = "";
let passedFinalDocumentId = "";
const thesisIds: string[] = [];
const defenseIds: string[] = [];
const documentIds: string[] = [];
const userIds: string[] = [];

async function roleId(name: string): Promise<number> {
  const rows = await db.select().from(schema.roles).where(eq(schema.roles.name, name));
  if (rows[0]) return rows[0].id;
  const [created] = await db.insert(schema.roles).values({ name }).returning({ id: schema.roles.id });
  return created.id;
}

async function user(prefix: string, name: string, roleId: number): Promise<string> {
  const [row] = await db.insert(schema.users).values({
    email: unique(prefix),
    fullName: name,
    roleId,
    passwordHash: DUMMY_HASH,
    isActive: true,
    mustChangePassword: false,
    tokenVersion: 0,
  } as any).returning({ id: schema.users.id });
  userIds.push(row.id);
  return row.id;
}

async function thesis(title: string, status: string, defenseStatus: string): Promise<{ thesisId: string; defenseId: string }> {
  const [row] = await db.insert(schema.theses).values({
    studentId,
    academicYearId,
    title,
    abstract: "Abstrak penelitian untuk pengujian arsip tugas akhir.",
    thesisType: "skripsi",
    status,
  } as any).returning({ id: schema.theses.id });
  thesisIds.push(row.id);
  await db.insert(schema.thesisSupervisors).values({ thesisId: row.id, supervisorId, assignedBy: kaprodiId } as any);
  const [defense] = await db.insert(schema.thesisDefenses).values({
    thesisId: row.id,
    status: defenseStatus,
    finalScore: defenseStatus === "passed" ? "80" : "70",
    revisionNotes: defenseStatus === "revision_required" ? "Revisi selesai melalui final_thesis approved" : null,
  } as any).returning({ id: schema.thesisDefenses.id });
  defenseIds.push(defense.id);
  return { thesisId: row.id, defenseId: defense.id };
}

const auth = async (id: string, role: string) => ({
  authorization: `Bearer ${await signAccessToken(id, role, 0)}`,
});
const request = (path: string, init: Record<string, any> = {}) => app.request(path, {
  ...init,
  headers: { "x-forwarded-for": "127.0.0.1", ...(init.headers ?? {}) },
});
const finalThesisBody = {
  documentType: "final_thesis",
  fileName: "skripsi-final.pdf",
  fileUrl: "https://storage.example/skripsi-final.pdf",
};
const archiveBody = {
  file_url: "https://storage.example/skripsi-final.pdf",
  file_name: "skripsi-final.pdf",
  abstract_id: "Abstrak Bahasa Indonesia yang sudah disetujui.",
  abstract_en: "Approved English abstract.",
  keywords: ["sistem informasi", "tugas akhir"],
  graduation_year: 2026,
};

beforeAll(async () => {
  const [mahasiswa, pembimbing, kaprodi, admin] = await Promise.all([
    roleId("mahasiswa"),
    roleId("dosen_pembimbing"),
    roleId("kaprodi"),
    roleId("admin_fakultas"),
  ]);
  const [year] = await db.insert(schema.academicYears).values({
    name: `2026/${crypto.randomUUID().slice(0, 4)}`,
    semester: "ganjil",
    startDate: "2026-08-01",
    endDate: "2026-12-31",
    isActive: false,
  } as any).returning({ id: schema.academicYears.id });
  academicYearId = year.id;

  [studentId, otherStudentId, supervisorId, unrelatedSupervisorId, kaprodiId, adminId] = await Promise.all([
    user("archive-student", "Archive Student", mahasiswa),
    user("archive-other-student", "Other Archive Student", mahasiswa),
    user("archive-supervisor", "Related Supervisor", pembimbing),
    user("archive-unrelated-supervisor", "Unrelated Supervisor", pembimbing),
    user("archive-kaprodi", "Archive Kaprodi", kaprodi),
    user("archive-admin", "Archive Admin", admin),
  ]);

  passedThesisId = (await thesis("Archive Passed Thesis", "defense_done", "passed")).thesisId;
  revisionThesisId = (await thesis("Archive Revision Thesis", "defense_done", "revision_required")).thesisId;
  pendingThesisId = (await thesis("Archive Pending Thesis", "defense_ready", "scheduled")).thesisId;
});

afterAll(async () => {
  if (documentIds.length) await db.delete(schema.documents).where(inArray(schema.documents.id, documentIds as any));
  if (thesisIds.length) await db.delete(schema.thesisArchives).where(inArray(schema.thesisArchives.thesisId, thesisIds as any));
  if (defenseIds.length) await db.delete(schema.auditLogs).where(inArray(schema.auditLogs.entityId, [...thesisIds, ...documentIds] as any));
  if (defenseIds.length) await db.delete(schema.thesisDefenses).where(inArray(schema.thesisDefenses.id, defenseIds as any));
  if (thesisIds.length) await db.delete(schema.thesisSupervisors).where(inArray(schema.thesisSupervisors.thesisId, thesisIds as any));
  if (thesisIds.length) await db.delete(schema.theses).where(inArray(schema.theses.id, thesisIds as any));
  await db.delete(schema.notifications).where(inArray(schema.notifications.userId, userIds as any));
  await db.delete(schema.users).where(inArray(schema.users.id, userIds as any));
  await db.delete(schema.academicYears).where(eq(schema.academicYears.id, academicYearId));
});

const uploadFinalThesis = async (thesisId: string) => {
  const response = await request("/api/v1/documents", {
    method: "POST",
    headers: { "content-type": "application/json", ...(await auth(studentId, "MAHASISWA")) },
    body: JSON.stringify({ thesisId, ...finalThesisBody }),
  });
  expect(response.status).toBe(201);
  const body = await response.json() as any;
  documentIds.push(body.data.id);
  return body.data.id as string;
};

const approveDocument = async (documentId: string, userId: string, role: string, notes = "Final thesis approved") => request(`/api/v1/documents/${documentId}/review`, {
  method: "PATCH",
  headers: { "content-type": "application/json", ...(await auth(userId, role)) },
  body: JSON.stringify({ decision: "approved", notes }),
});

describe("final thesis approval and graduation archive", () => {
  it("allows only the owner to upload final_thesis and only the related supervisor to review it", async () => {
    const otherUpload = await request("/api/v1/documents", {
      method: "POST",
      headers: { "content-type": "application/json", ...(await auth(otherStudentId, "MAHASISWA")) },
      body: JSON.stringify({ thesisId: passedThesisId, ...finalThesisBody }),
    });
    expect(otherUpload.status).toBe(403);

    passedFinalDocumentId = await uploadFinalThesis(passedThesisId);
    const unrelatedReview = await approveDocument(passedFinalDocumentId, unrelatedSupervisorId, "DOSEN_PEMBIMBING");
    expect(unrelatedReview.status).toBe(403);

    const relatedReview = await approveDocument(passedFinalDocumentId, supervisorId, "DOSEN_PEMBIMBING");
    expect(relatedReview.status).toBe(200);
    expect((await relatedReview.json() as any).data.status).toBe("approved");
  });

  it("requires an eligible Sidang outcome and latest approved final_thesis before archiving", async () => {
    const missingDocument = await request(`/api/v1/theses/${revisionThesisId}/archive`, {
      method: "POST",
      headers: { "content-type": "application/json", ...(await auth(kaprodiId, "KAPRODI")) },
      body: JSON.stringify(archiveBody),
    });
    expect(missingDocument.status).toBe(422);

    const pendingFinal = await uploadFinalThesis(revisionThesisId);
    const pendingArchive = await request(`/api/v1/theses/${revisionThesisId}/archive`, {
      method: "POST",
      headers: { "content-type": "application/json", ...(await auth(kaprodiId, "KAPRODI")) },
      body: JSON.stringify(archiveBody),
    });
    expect(pendingArchive.status).toBe(422);

    const approved = await approveDocument(pendingFinal, adminId, "ADMIN_FAKULTAS");
    expect(approved.status).toBe(200);

    const archived = await request(`/api/v1/theses/${revisionThesisId}/archive`, {
      method: "POST",
      headers: { "content-type": "application/json", ...(await auth(kaprodiId, "KAPRODI")) },
      body: JSON.stringify(archiveBody),
    });
    expect(archived.status).toBe(201);
    const body = await archived.json() as any;
    expect(body.data.thesisId).toBe(revisionThesisId);
    expect(body.data.graduationYear).toBe(2026);
    const [thesisRow] = await db.select().from(schema.theses).where(eq(schema.theses.id, revisionThesisId));
    expect(thesisRow.status).toBe("graduated");
    expect(thesisRow.graduatedAt).toBeTruthy();
  });

  it("rejects invalid archive metadata, duplicate creation, and concurrent archive creation", async () => {
    const invalid = await request(`/api/v1/theses/${passedThesisId}/archive`, {
      method: "POST",
      headers: { "content-type": "application/json", ...(await auth(adminId, "ADMIN_FAKULTAS")) },
      body: JSON.stringify({ ...archiveBody, abstract_id: "", graduation_year: 0 }),
    });
    expect(invalid.status).toBe(422);

    const concurrent = await Promise.all([
      request(`/api/v1/theses/${passedThesisId}/archive`, {
        method: "POST",
        headers: { "content-type": "application/json", ...(await auth(adminId, "ADMIN_FAKULTAS")) },
        body: JSON.stringify(archiveBody),
      }),
      request(`/api/v1/theses/${passedThesisId}/archive`, {
        method: "POST",
        headers: { "content-type": "application/json", ...(await auth(kaprodiId, "KAPRODI")) },
        body: JSON.stringify(archiveBody),
      }),
    ]);
    expect(concurrent.filter((response) => response.status === 201)).toHaveLength(1);
    expect(concurrent.filter((response) => response.status === 409)).toHaveLength(1);
  });

  it("keeps archive and graduated state terminal and scopes archive reads", async () => {
    const studentList = await request("/api/v1/archives", { headers: await auth(studentId, "MAHASISWA") });
    expect(studentList.status).toBe(200);
    expect((await studentList.json() as any).data.some((row: any) => row.thesisId === passedThesisId)).toBe(true);

    const otherList = await request("/api/v1/archives", { headers: await auth(otherStudentId, "MAHASISWA") });
    expect(otherList.status).toBe(200);
    expect((await otherList.json() as any).data.some((row: any) => row.thesisId === passedThesisId)).toBe(false);

    const [archive] = await db.select().from(schema.thesisArchives).where(eq(schema.thesisArchives.thesisId, passedThesisId));
    const detail = await request(`/api/v1/archives/${archive.id}`, { headers: await auth(otherStudentId, "MAHASISWA") });
    expect(detail.status).toBe(403);

    const rearchive = await request(`/api/v1/theses/${passedThesisId}/archive`, {
      method: "POST",
      headers: { "content-type": "application/json", ...(await auth(adminId, "ADMIN_FAKULTAS")) },
      body: JSON.stringify(archiveBody),
    });
    expect(rearchive.status).toBe(409);
  });

  it("rejects final_thesis upload and archive when Thesis is not defense_done", async () => {
    const finalDocument = await request("/api/v1/documents", {
      method: "POST",
      headers: { "content-type": "application/json", ...(await auth(studentId, "MAHASISWA")) },
      body: JSON.stringify({ thesisId: pendingThesisId, ...finalThesisBody }),
    });
    expect(finalDocument.status).toBe(409);

    const response = await request(`/api/v1/theses/${pendingThesisId}/archive`, {
      method: "POST",
      headers: { "content-type": "application/json", ...(await auth(kaprodiId, "KAPRODI")) },
      body: JSON.stringify(archiveBody),
    });
    expect(response.status).toBe(409);
  });
});
