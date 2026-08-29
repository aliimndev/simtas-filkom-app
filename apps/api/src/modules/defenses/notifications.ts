import { and, eq, inArray } from "drizzle-orm";
import { schema } from "@sims/db";
import { getDb } from "../../db";
import { loadConfig } from "../../config";

function db() {
  return getDb(loadConfig().databaseUrl);
}

export async function notifySchedule(defenseId: string, studentId: string, examinerIds: string[]) {
  const connection = db();
  const recipients = [...new Set([studentId, ...examinerIds])];
  await connection.insert(schema.notifications).values(
    recipients.map((userId) => ({
      userId,
      title: "Sidang Terjadwal",
      message: "Sidang Anda telah dijadwalkan dan Penguji telah ditetapkan.",
      type: "defense",
      link: `/dashboard/defenses/${defenseId}`,
    })) as any,
  );
}

export async function notifyFinalization(
  defenseId: string,
  studentId: string,
  examinerIds: string[],
  status: "passed" | "revision_required" | "failed",
  finalScore: number,
) {
  const connection = db();
  const supervisors = await connection
    .select({ userId: schema.thesisSupervisors.supervisorId })
    .from(schema.thesisSupervisors)
    .innerJoin(schema.theses, eq(schema.thesisSupervisors.thesisId, schema.theses.id))
    .innerJoin(schema.thesisDefenses, eq(schema.thesisDefenses.thesisId, schema.theses.id))
    .where(eq(schema.thesisDefenses.id, defenseId as any));
  const staffRoles = await connection
    .select({ id: schema.roles.id })
    .from(schema.roles)
    .where(inArray(schema.roles.name, ["kaprodi", "admin_fakultas"]));
  const staff = staffRoles.length === 0
    ? []
    : await connection
      .select({ userId: schema.users.id })
      .from(schema.users)
      .where(and(inArray(schema.users.roleId, staffRoles.map((role) => role.id)), eq(schema.users.isActive, true)));
  const recipients = [...new Set([
    studentId,
    ...examinerIds,
    ...supervisors.map((row) => row.userId),
    ...staff.map((row) => row.userId),
  ])];
  const label = status === "passed" ? "Lulus" : status === "revision_required" ? "Perlu Revisi" : "Tidak Lulus";
  await connection.insert(schema.notifications).values(
    recipients.map((userId) => ({
      userId,
      title: `Hasil Sidang ${label}`,
      message: `Sidang telah difinalisasi dengan hasil ${status} dan nilai ${finalScore}.`,
      type: "defense",
      link: `/dashboard/defenses/${defenseId}`,
    })) as any,
  );
}

export async function notifySubmission(defenseId: string, studentId: string) {
  const connection = db();
  const supervisors = await connection
    .select({ userId: schema.thesisSupervisors.supervisorId })
    .from(schema.thesisSupervisors)
    .innerJoin(schema.theses, eq(schema.thesisSupervisors.thesisId, schema.theses.id))
    .innerJoin(schema.thesisDefenses, eq(schema.thesisDefenses.thesisId, schema.theses.id))
    .where(eq(schema.thesisDefenses.id, defenseId as any));
  const kaprodiRoles = await connection
    .select({ id: schema.roles.id })
    .from(schema.roles)
    .where(inArray(schema.roles.name, ["kaprodi", "admin_fakultas"]));
  const staff = kaprodiRoles.length === 0
    ? []
    : await connection
      .select({ userId: schema.users.id })
      .from(schema.users)
      .where(and(
        inArray(schema.users.roleId, kaprodiRoles.map((role) => role.id)),
        eq(schema.users.isActive, true),
      ));
  const recipients = [...new Set([
    studentId,
    ...supervisors.map((row) => row.userId),
    ...staff.map((row) => row.userId),
  ])];
  await connection.insert(schema.notifications).values(
    recipients.map((userId) => ({
      userId,
      title: "Pengajuan Sidang Baru",
      message: "Mahasiswa mengajukan Sidang dan menunggu penjadwalan.",
      type: "defense",
      link: `/dashboard/defenses/${defenseId}`,
    })) as any,
  );
}
