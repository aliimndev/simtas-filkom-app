import { and, eq, inArray } from "drizzle-orm";
import { schema } from "@sims/db";
import { getDb } from "../../db";
import { loadConfig } from "../../config";

function db() {
  return getDb(loadConfig().databaseUrl);
}

export async function notifyKaprodi(seminarId: string) {
  const connection = db();
  const roles = await connection.select({ id: schema.roles.id }).from(schema.roles).where(eq(schema.roles.name, "kaprodi"));
  if (!roles[0]) return;
  const kaprodi = await connection.select({ id: schema.users.id }).from(schema.users).where(eq(schema.users.roleId, roles[0].id));
  if (kaprodi.length === 0) return;
  await connection.insert(schema.notifications).values(
    kaprodi.map((user) => ({
      userId: user.id,
      title: "Pengajuan Seminar Baru",
      message: "Mahasiswa mengajukan Seminar dan menunggu penjadwalan.",
      type: "seminar",
      link: `/dashboard/seminars/${seminarId}`,
    })) as any,
  );
}

export async function notifySchedule(seminarId: string, studentId: string, examinerIds: string[]) {
  const connection = db();
  const recipients = [...new Set([studentId, ...examinerIds])];
  await connection.insert(schema.notifications).values(
    recipients.map((userId) => ({
      userId,
      title: "Seminar Terjadwal",
      message: "Seminar Anda telah dijadwalkan dan Penguji telah ditetapkan.",
      type: "seminar",
      link: `/dashboard/seminars/${seminarId}`,
    })) as any,
  );
}

export async function notifyFinalization(
  seminarId: string,
  studentId: string,
  examinerIds: string[],
  status: "passed" | "failed",
  finalScore: number,
) {
  const connection = db();
  const supervisors = await connection
    .select({ userId: schema.thesisSupervisors.supervisorId })
    .from(schema.thesisSupervisors)
    .innerJoin(schema.seminars, eq(schema.thesisSupervisors.thesisId, schema.seminars.thesisId))
    .where(eq(schema.seminars.id, seminarId as any));
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
  await connection.insert(schema.notifications).values(
    recipients.map((userId) => ({
      userId,
      title: status === "passed" ? "Hasil Seminar Lulus" : "Hasil Seminar Tidak Lulus",
      message: `Seminar telah difinalisasi dengan hasil ${status} dan nilai ${finalScore}.`,
      type: "seminar",
      link: `/dashboard/seminars/${seminarId}`,
    })) as any,
  );
}

export async function notifyCancellation(
  seminarId: string,
  studentId: string,
  examinerIds: string[],
  reason: string,
) {
  const connection = db();
  const supervisors = await connection
    .select({ userId: schema.thesisSupervisors.supervisorId })
    .from(schema.thesisSupervisors)
    .innerJoin(schema.seminars, eq(schema.thesisSupervisors.thesisId, schema.seminars.thesisId))
    .where(eq(schema.seminars.id, seminarId as any));
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
  await connection.insert(schema.notifications).values(
    recipients.map((userId) => ({
      userId,
      title: "Seminar Dibatalkan",
      message: `Seminar dibatalkan. Alasan: ${reason}`,
      type: "seminar",
      link: `/dashboard/seminars/${seminarId}`,
    })) as any,
  );
}
