import { and, eq, inArray } from "drizzle-orm";
import { schema } from "@sims/db";
import { getDb } from "../../db";
import { loadConfig } from "../../config";

function db() {
  return getDb(loadConfig().databaseUrl);
}

export async function notifyGraduation(thesisId: string, studentId: string, supervisorIds: string[]) {
  const connection = db();
  const staffRoles = await connection
    .select({ id: schema.roles.id })
    .from(schema.roles)
    .where(inArray(schema.roles.name, ["kaprodi", "admin_fakultas"]));
  const staff = staffRoles.length === 0
    ? []
    : await connection
      .select({ userId: schema.users.id })
      .from(schema.users)
      .where(and(
        inArray(schema.users.roleId, staffRoles.map((role) => role.id)),
        eq(schema.users.isActive, true),
      ));
  const recipients = [...new Set([
    studentId,
    ...supervisorIds,
    ...staff.map((row) => row.userId),
  ])];
  await connection.insert(schema.notifications).values(
    recipients.map((userId) => ({
      userId,
      title: "Tugas Akhir Diarsipkan",
      message: "Tugas Akhir telah diarsipkan dan Thesis berstatus graduated.",
      type: "archive",
      link: `/dashboard/archives/${thesisId}`,
    })) as any,
  );
}
