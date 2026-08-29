import { Hono } from "hono";
import { z } from "zod";
import { Authenticate } from "../../middleware/auth";
import { ArchiveError } from "./errors";
import { createArchive, getArchive, listArchives } from "./service";

export const archivesRoutes = new Hono();
export const archiveCreationRoutes = new Hono();
archivesRoutes.use("*", Authenticate());
archiveCreationRoutes.use("*", Authenticate());

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
  if (error instanceof ArchiveError) return c.json({ error: { code: error.code, message: error.message } }, error.status as any);
  throw error;
}

const uuid = z.string().uuid();
const archiveSchema = z.object({
  file_url: z.string().trim().min(1),
  file_name: z.string().trim().min(1).max(255),
  abstract_id: z.string().trim().min(1),
  abstract_en: z.string().trim().nullable().optional(),
  keywords: z.array(z.string().trim().min(1)).min(1),
  graduation_year: z.number().int(),
});

archivesRoutes.get("/", async (c: any) => {
  try {
    return c.json(await listArchives(actorFrom(c)), 200);
  } catch (error) {
    return fail(c, error);
  }
});

archivesRoutes.get("/:id", async (c: any) => {
  try {
    const parsedId = uuid.safeParse(c.req.param("id"));
    if (!parsedId.success) return c.json({ error: { code: "VALIDATION", message: "id Arsip tidak valid" } }, 400);
    return c.json({ data: await getArchive(parsedId.data, actorFrom(c)) }, 200);
  } catch (error) {
    return fail(c, error);
  }
});

archiveCreationRoutes.post("/:id/archive", async (c: any) => {
  try {
    const parsedId = uuid.safeParse(c.req.param("id"));
    const parsedBody = archiveSchema.safeParse(await c.req.json().catch(() => null));
    if (!parsedId.success) return c.json({ error: { code: "VALIDATION", message: "id Thesis tidak valid" } }, 400);
    if (!parsedBody.success) return c.json({ error: { code: "VALIDATION", message: "metadata Arsip tidak valid" } }, 422);
    const data = await createArchive(parsedId.data, {
      fileUrl: parsedBody.data.file_url,
      fileName: parsedBody.data.file_name,
      abstractId: parsedBody.data.abstract_id,
      abstractEn: parsedBody.data.abstract_en,
      keywords: parsedBody.data.keywords,
      graduationYear: parsedBody.data.graduation_year,
    }, actorFrom(c));
    return c.json({ data }, 201);
  } catch (error) {
    return fail(c, error);
  }
});
