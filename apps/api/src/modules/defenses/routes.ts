import { Hono } from "hono";
import { z } from "zod";
import { Authenticate } from "../../middleware/auth";
import {
  cancelSuratTugas,
  createSuratTugas,
  DefenseError,
  getDefense,
  issueSuratTugas,
  listDefenses,
  finalizeDefense,
  saveDefenseScores,
  scheduleDefense,
  submitDefense,
} from "./service";

export const defensesRoutes = new Hono();
export const defenseSubmissionRoutes = new Hono();
export const suratTugasRoutes = new Hono();
defensesRoutes.use("*", Authenticate());
defenseSubmissionRoutes.use("*", Authenticate());
suratTugasRoutes.use("*", Authenticate());

const actorFrom = (c: any) => ({
  userId: c.get("user").id,
  role: c.get("user").role,
  ipAddress: c.req.header("x-forwarded-for") ?? null,
  userAgent: c.req.header("user-agent") ?? null,
});

function fail(c: any, error: unknown) {
  if (error instanceof DefenseError) return c.json({ error: { code: error.code, message: error.message } }, error.status as any);
  throw error;
}

const uuid = z.string().uuid();
const suratTugasSchema = z.object({
  letter_number: z.string().trim().min(1).max(100),
  issue_date: z.string().min(1),
  file_name: z.string().trim().min(1).max(255),
  file_url: z.string().trim().min(1),
});
const scheduleSchema = z.object({
  scheduled_at: z.string().min(1),
  room: z.string().trim().min(1).max(100),
  examiner_ids: z.array(uuid).min(2),
});
const cancellationSchema = z.object({ reason: z.string().trim().min(1) });
const scoreSchema = z.object({
  scores: z.array(z.object({
    component_name: z.string(),
    component_weight: z.number(),
    score: z.number(),
    notes: z.string().nullable().optional(),
  })),
});
const finalizeSchema = z.object({ revision_notes: z.string().trim().nullable().optional() });

defensesRoutes.get("/", async (c: any) => {
  try {
    const thesisId = c.req.query("thesis_id");
    if (thesisId && !uuid.safeParse(thesisId).success) return c.json({ error: { code: "VALIDATION", message: "thesis_id tidak valid" } }, 400);
    const data = await listDefenses({
      thesisId,
      page: Number(c.req.query("page") ?? 1),
      perPage: Number(c.req.query("per_page") ?? 20),
    }, actorFrom(c));
    return c.json(data, 200);
  } catch (error) {
    return fail(c, error);
  }
});

defensesRoutes.get("/:id", async (c: any) => {
  try {
    const parsed = uuid.safeParse(c.req.param("id"));
    if (!parsed.success) return c.json({ error: { code: "VALIDATION", message: "id Sidang tidak valid" } }, 400);
    return c.json({ data: await getDefense(parsed.data, actorFrom(c)) }, 200);
  } catch (error) {
    return fail(c, error);
  }
});

defensesRoutes.put("/:id/schedule", async (c: any) => {
  try {
    const parsedId = uuid.safeParse(c.req.param("id"));
    const parsedBody = scheduleSchema.safeParse(await c.req.json().catch(() => null));
    if (!parsedId.success) return c.json({ error: { code: "VALIDATION", message: "id Sidang tidak valid" } }, 400);
    if (!parsedBody.success) return c.json({ error: { code: "VALIDATION", message: "scheduled_at, room, dan minimal 2 examiner_ids wajib diisi" } }, 422);
    const data = await scheduleDefense(parsedId.data, {
      scheduledAt: parsedBody.data.scheduled_at,
      room: parsedBody.data.room,
      examinerIds: parsedBody.data.examiner_ids,
    }, actorFrom(c));
    return c.json({ data }, 200);
  } catch (error) {
    return fail(c, error);
  }
});

defensesRoutes.post("/:id/surat-tugas", async (c: any) => {
  try {
    const parsedId = uuid.safeParse(c.req.param("id"));
    const parsedBody = suratTugasSchema.safeParse(await c.req.json().catch(() => null));
    if (!parsedId.success) return c.json({ error: { code: "VALIDATION", message: "id Sidang tidak valid" } }, 400);
    if (!parsedBody.success) return c.json({ error: { code: "VALIDATION", message: "letter_number, issue_date, file_name, dan file_url wajib diisi" } }, 422);
    const data = await createSuratTugas(parsedId.data, {
      letterNumber: parsedBody.data.letter_number,
      issueDate: parsedBody.data.issue_date,
      fileName: parsedBody.data.file_name,
      fileUrl: parsedBody.data.file_url,
    }, actorFrom(c));
    return c.json({ data }, 201);
  } catch (error) {
    return fail(c, error);
  }
});

defensesRoutes.put("/:id/scores", async (c: any) => {
  try {
    const parsedId = uuid.safeParse(c.req.param("id"));
    const parsedBody = scoreSchema.safeParse(await c.req.json().catch(() => null));
    if (!parsedId.success) return c.json({ error: { code: "VALIDATION", message: "id Sidang tidak valid" } }, 400);
    if (!parsedBody.success) return c.json({ error: { code: "VALIDATION", message: "scores wajib berisi rubric Sidang yang lengkap" } }, 422);
    const data = await saveDefenseScores(parsedId.data, parsedBody.data.scores.map((score) => ({
      componentName: score.component_name,
      componentWeight: score.component_weight,
      score: score.score,
      notes: score.notes,
    })), actorFrom(c));
    return c.json({ data }, 200);
  } catch (error) {
    return fail(c, error);
  }
});

defensesRoutes.post("/:id/finalize", async (c: any) => {
  try {
    const parsedId = uuid.safeParse(c.req.param("id"));
    const parsedBody = finalizeSchema.safeParse(await c.req.json().catch(() => ({})));
    if (!parsedId.success) return c.json({ error: { code: "VALIDATION", message: "id Sidang tidak valid" } }, 400);
    if (!parsedBody.success) return c.json({ error: { code: "VALIDATION", message: "revision_notes tidak valid" } }, 422);
    const data = await finalizeDefense(parsedId.data, { revisionNotes: parsedBody.data.revision_notes }, actorFrom(c));
    return c.json({ data }, 200);
  } catch (error) {
    return fail(c, error);
  }
});

suratTugasRoutes.post("/:id/issue", async (c: any) => {
  try {
    const parsedId = uuid.safeParse(c.req.param("id"));
    if (!parsedId.success) return c.json({ error: { code: "VALIDATION", message: "id Surat Tugas tidak valid" } }, 400);
    return c.json({ data: await issueSuratTugas(parsedId.data, actorFrom(c)) }, 200);
  } catch (error) {
    return fail(c, error);
  }
});

suratTugasRoutes.post("/:id/cancel", async (c: any) => {
  try {
    const parsedId = uuid.safeParse(c.req.param("id"));
    const parsedBody = cancellationSchema.safeParse(await c.req.json().catch(() => null));
    if (!parsedId.success) return c.json({ error: { code: "VALIDATION", message: "id Surat Tugas tidak valid" } }, 400);
    if (!parsedBody.success) return c.json({ error: { code: "VALIDATION", message: "reason wajib diisi" } }, 422);
    return c.json({ data: await cancelSuratTugas(parsedId.data, parsedBody.data.reason, actorFrom(c)) }, 200);
  } catch (error) {
    return fail(c, error);
  }
});

defenseSubmissionRoutes.post("/:thesisId/defenses", async (c: any) => {
  try {
    const parsed = uuid.safeParse(c.req.param("thesisId"));
    if (!parsed.success) return c.json({ error: { code: "VALIDATION", message: "thesisId tidak valid" } }, 400);
    return c.json({ data: await submitDefense(parsed.data, actorFrom(c)) }, 201);
  } catch (error) {
    return fail(c, error);
  }
});
