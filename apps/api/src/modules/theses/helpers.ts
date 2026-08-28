import { isIP } from "node:net";
import { eq } from "drizzle-orm";
import { getDb } from "../../db";
import { loadConfig } from "../../config";
import { schema } from "@sims/db";
import type { Actor, ThesisDetail } from "./types";

export const db_ = () => getDb(loadConfig().databaseUrl);

export async function getDetail(db: any, thesis: any): Promise<ThesisDetail> {
  const student = (await db.select().from(schema.users).where(eq(schema.users.id, thesis.studentId)))[0];
  const ay = (await db.select().from(schema.academicYears).where(eq(schema.academicYears.id, thesis.academicYearId)))[0];
  const supervisors = await db
    .select({ id: schema.users.id, fullName: schema.users.fullName })
    .from(schema.users)
    .innerJoin(schema.thesisSupervisors, eq(schema.thesisSupervisors.supervisorId, schema.users.id))
    .where(eq(schema.thesisSupervisors.thesisId, thesis.id));
  return {
    id: thesis.id,
    title: thesis.title,
    abstract: thesis.abstract,
    fieldOfStudy: thesis.fieldOfStudy,
    thesisType: thesis.thesisType,
    status: thesis.status,
    kaprodiNotes: thesis.kaprodiNotes,
    student: { id: thesis.studentId, fullName: student?.fullName ?? "", nim: student?.nimNidn ?? null },
    supervisors,
    academicYear: { name: ay?.name ?? "", semester: ay?.semester ?? "" },
    submittedAt: thesis.submittedAt?.toISOString?.() ?? String(thesis.submittedAt),
    approvedAt: thesis.approvedAt ? (thesis.approvedAt.toISOString?.() ?? String(thesis.approvedAt)) : null,
  };
}

export function validIp(value?: string | null): string | null {
  const candidate = value?.split(",")[0]?.trim();
  return candidate && isIP(candidate) ? candidate : null;
}

export async function logAudit(db: any, actor: Actor, action: string, entityId: string, oldValue?: any, newValue?: any) {
  await db.insert(schema.auditLogs).values({
    userId: actor.userId,
    action,
    entityType: "thesis",
    entityId,
    oldValue: oldValue ?? null,
    newValue: newValue ?? null,
    ipAddress: validIp(actor.ipAddress),
    userAgent: actor.userAgent || null,
  });
}
