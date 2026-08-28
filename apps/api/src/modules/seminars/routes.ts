import { Hono } from "hono";
import { z } from "zod";
import { Authenticate } from "../../middleware/auth";
import { SeminarError, getSeminar, listSeminars, scheduleSeminar, submitSeminar } from "./service";

export const seminarsRoutes = new Hono();
export const seminarSubmissionRoutes = new Hono();
seminarsRoutes.use("*", Authenticate());
seminarSubmissionRoutes.use("*", Authenticate());

const actorFrom = (c: any) => {
  const user = c.get("user");
  return {
    userId: user.id,
    role: user.role,
    ipAddress: c.req.header("x-forwarded-for") ?? null,
    userAgent: c.req.header("user-agent") ?? null,
  };
};

function fail(c: any, error: unknown) {
  if (error instanceof SeminarError) {
    return c.json({ error: { code: error.code, message: error.message } }, error.status as any);
  }
  throw error;
}

const uuid = z.string().uuid();
const scheduleSchema = z.object({
  scheduled_at: z.string().min(1),
  room: z.string().trim().min(1).max(100),
  examiner_ids: z.array(uuid).min(2),
});

// GET /seminars — role-scoped Seminar list.
seminarsRoutes.get("/", async (c: any) => {
  try {
    const status = c.req.query("status");
    if (status && !["pending", "scheduled", "passed", "failed"].includes(status)) {
      return c.json({ error: { code: "VALIDATION", message: "status Seminar tidak valid" } }, 400);
    }
    const result = await listSeminars(
      {
        status,
        page: Number(c.req.query("page") ?? 1),
        perPage: Number(c.req.query("per_page") ?? 20),
      },
      actorFrom(c),
    );
    return c.json(result, 200);
  } catch (error) {
    return fail(c, error);
  }
});

// PUT /seminars/:id/schedule — Kaprodi/Admin schedules and assigns Penguji.
seminarsRoutes.put("/:id/schedule", async (c: any) => {
  try {
    const parsedId = uuid.safeParse(c.req.param("id"));
    if (!parsedId.success) return c.json({ error: { code: "VALIDATION", message: "id Seminar tidak valid" } }, 400);
    const body = await c.req.json().catch(() => null);
    const parsedBody = scheduleSchema.safeParse(body);
    if (!parsedBody.success) {
      return c.json({ error: { code: "VALIDATION", message: "scheduled_at, room, dan minimal 2 examiner_ids wajib diisi" } }, 422);
    }
    const data = await scheduleSeminar(
      parsedId.data,
      {
        scheduledAt: parsedBody.data.scheduled_at,
        room: parsedBody.data.room,
        examinerIds: parsedBody.data.examiner_ids,
      },
      actorFrom(c),
    );
    return c.json({ data }, 200);
  } catch (error) {
    return fail(c, error);
  }
});


// GET /seminars/:id — role-scoped Seminar detail.
seminarsRoutes.get("/:id", async (c: any) => {
  try {
    const parsed = uuid.safeParse(c.req.param("id"));
    if (!parsed.success) return c.json({ error: { code: "VALIDATION", message: "id Seminar tidak valid" } }, 400);
    return c.json({ data: await getSeminar(parsed.data, actorFrom(c)) }, 200);
  } catch (error) {
    return fail(c, error);
  }
});

// POST /theses/:thesisId/seminars — Mahasiswa owner submission.
seminarSubmissionRoutes.post("/:thesisId/seminars", async (c: any) => {
  try {
    const parsed = uuid.safeParse(c.req.param("thesisId"));
    if (!parsed.success) return c.json({ error: { code: "VALIDATION", message: "thesisId tidak valid" } }, 400);
    return c.json({ data: await submitSeminar(parsed.data, actorFrom(c)) }, 201);
  } catch (error) {
    return fail(c, error);
  }
});
