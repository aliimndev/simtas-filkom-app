import { beforeAll, afterAll, describe, expect, it } from "bun:test";
import { eq } from "drizzle-orm";
import { createApp } from "../../src/app";
import { loadConfig } from "../../src/config";
import { signAccessToken } from "../../src/modules/auth";
import { getDb } from "../../src/db";
import { schema } from "@sims/db";
import { documentsRoutes } from "../../src/modules/documents";

const TEST_DB_URL = process.env.DATABASE_URL ?? "postgres://postgres@localhost:5433/simtas";
const cfg = loadConfig({ ...process.env, NODE_ENV: "test", DATABASE_URL: TEST_DB_URL } as any);
const db = getDb(cfg.databaseUrl);
const app = createApp(cfg);
app.route("/api/v1/documents", documentsRoutes);

async function roleIdByName(name: string): Promise<number> {
  const r = await db.select().from(schema.roles).where(eq(schema.roles.name, name));
  return (r[0] as any).id;
}

let studentId: string;
let student2Id: string;
let kaprodiId: string;
let thesisId: string;
let academicYearId: string;

beforeAll(async () => {
  const mhsRole = await roleIdByName("mahasiswa");
  const kaprodiRole = await roleIdByName("kaprodi");

  academicYearId = crypto.randomUUID();
  await db.insert(schema.academicYears).values({
    id: academicYearId as any,
    name: "2025/2026",
    semester: "genap",
    startDate: "2026-01-01",
    endDate: "2026-06-30",
    isActive: false,
  } as any);

  studentId = crypto.randomUUID();
  student2Id = crypto.randomUUID();
  kaprodiId = crypto.randomUUID();
  await db.insert(schema.users).values([
    { id: studentId as any, email: `stu-${studentId}@filkom.ac.id`, fullName: "Student One", roleId: mhsRole, passwordHash: "x", isActive: true },
    { id: student2Id as any, email: `stu2-${student2Id}@filkom.ac.id`, fullName: "Student Two", roleId: mhsRole, passwordHash: "x", isActive: true },
    { id: kaprodiId as any, email: `kap-${kaprodiId}@filkom.ac.id`, fullName: "Kaprodi One", roleId: kaprodiRole, passwordHash: "x", isActive: true },
  ] as any);

  thesisId = crypto.randomUUID();
  await db.insert(schema.theses).values({
    id: thesisId as any,
    studentId: studentId as any,
    academicYearId: academicYearId as any,
    title: "Test Thesis",
    thesisType: "skripsi",
    status: "in_progress",
  } as any);
});

afterAll(async () => {
  await db.delete(schema.documents).where(eq(schema.documents.thesisId, thesisId as any));
  await db.delete(schema.theses).where(eq(schema.theses.id, thesisId as any));
  await db.delete(schema.users).where(eq(schema.users.id, studentId as any));
  await db.delete(schema.users).where(eq(schema.users.id, student2Id as any));
  await db.delete(schema.users).where(eq(schema.users.id, kaprodiId as any));
  await db.delete(schema.academicYears).where(eq(schema.academicYears.id, academicYearId as any));
});

let seq = 0;
const req = (path: string, init: Record<string, any> = {}) =>
  app.request(path, {
    ...init,
    headers: { "x-forwarded-for": `doc-${seq++}`, ...(init.headers ?? {}) },
  });

const studentTok = () => signAccessToken(studentId, "MAHASISWA", 0);
const student2Tok = () => signAccessToken(student2Id, "MAHASISWA", 0);
const kaprodiTok = () => signAccessToken(kaprodiId, "KAPRODI", 0);

const auth = (tok: Promise<string>) => tok.then((t) => ({ authorization: `Bearer ${t}` }));

describe("documents parity", () => {
  it("student uploads with versioning per (thesis, document_type)", async () => {
    const up = async (v?: number) =>
      req("/api/v1/documents", {
        method: "POST",
        headers: { "content-type": "application/json", ...(await auth(studentTok())) },
        body: JSON.stringify({
          thesisId,
          documentType: "draft_chapter",
          chapterNumber: 1,
          fileName: "bab1.pdf",
          fileUrl: "https://storage/bab1.pdf",
        }),
      });

    const r1 = await up();
    expect(r1.status).toBe(201);
    const d1 = (await r1.json()) as any;
    expect(d1.data.version).toBe(1);
    expect(d1.data.status).toBe("pending_review");

    const r2 = await up();
    expect(r2.status).toBe(201);
    const d2 = (await r2.json()) as any;
    expect(d2.data.version).toBe(2);
  });

  it("student upload without chapter_number on draft_chapter is rejected (400)", async () => {
    const r = await req("/api/v1/documents", {
      method: "POST",
      headers: { "content-type": "application/json", ...(await auth(studentTok())) },
      body: JSON.stringify({ thesisId, documentType: "draft_chapter", fileName: "x.pdf", fileUrl: "https://s/x.pdf" }),
    });
    expect(r.status).toBe(400);
  });

  it("other student cannot upload to a thesis they don't own (403)", async () => {
    const r = await req("/api/v1/documents", {
      method: "POST",
      headers: { "content-type": "application/json", ...(await auth(student2Tok())) },
      body: JSON.stringify({ thesisId, documentType: "proposal", fileName: "p.pdf", fileUrl: "https://s/p.pdf" }),
    });
    expect(r.status).toBe(403);
  });

  it("lists documents scoped to the thesis", async () => {
    const r = await req(`/api/v1/documents?thesisId=${thesisId}`, { headers: await auth(studentTok()) });
    expect(r.status).toBe(200);
    const body = (await r.json()) as any;
    expect(body.data.length).toBeGreaterThan(0);
    expect(body.data[0].thesisId).toBe(thesisId);
  });

  it("kaprodi approves a pending document", async () => {
    // grab the latest draft_chapter doc for this thesis
    const list = await req(`/api/v1/documents?thesisId=${thesisId}&document_type=draft_chapter`, {
      headers: await auth(kaprodiTok()),
    });
    const docs = (await list.json()) as any;
    const docId = docs.data[0].id;

    const r = await req(`/api/v1/documents/${docId}/review`, {
      method: "PATCH",
      headers: { "content-type": "application/json", ...(await auth(kaprodiTok())) },
      body: JSON.stringify({ decision: "APPROVED", notes: "ok" }),
    });
    expect(r.status).toBe(200);
    const body = (await r.json()) as any;
    expect(body.data.status).toBe("approved");
    expect(body.data.reviewerId).toBe(kaprodiId);
    expect(body.data.reviewedAt).toBeTruthy();
  });

  it("reviewing a non-pending document is rejected (409)", async () => {
    const list = await req(`/api/v1/documents?thesisId=${thesisId}&document_type=draft_chapter&status=approved`, {
      headers: await auth(kaprodiTok()),
    });
    const docs = (await list.json()) as any;
    const docId = docs.data[0].id;
    const r = await req(`/api/v1/documents/${docId}/review`, {
      method: "PATCH",
      headers: { "content-type": "application/json", ...(await auth(kaprodiTok())) },
      body: JSON.stringify({ decision: "APPROVED" }),
    });
    expect(r.status).toBe(409);
  });

  it("other student cannot view the document (403)", async () => {
    const list = await req(`/api/v1/documents?thesisId=${thesisId}`, { headers: await auth(studentTok()) });
    const docs = (await list.json()) as any;
    const docId = docs.data[0].id;
    const r = await req(`/api/v1/documents/${docId}`, { headers: await auth(student2Tok()) });
    expect(r.status).toBe(403);
  });

  it("returns 404 for a missing document", async () => {
    const r = await req(`/api/v1/documents/${crypto.randomUUID()}`, { headers: await auth(kaprodiTok()) });
    expect(r.status).toBe(404);
  });

  it("download returns the stored file_url (no signing)", async () => {
    const list = await req(`/api/v1/documents?thesisId=${thesisId}`, { headers: await auth(kaprodiTok()) });
    const docs = (await list.json()) as any;
    const docId = docs.data[0].id;
    const r = await req(`/api/v1/documents/${docId}/download`, { headers: await auth(kaprodiTok()) });
    expect(r.status).toBe(200);
    expect((await r.json() as any).fileUrl).toBeTruthy();
  });
});
