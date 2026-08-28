import { Hono } from "hono";
import { z } from "zod";
import { Authenticate } from "../middleware/auth";
import { SeminarError, getSeminar, listSeminars, submitSeminar } from "../services/seminars";

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
