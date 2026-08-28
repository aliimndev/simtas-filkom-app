import { eq } from "drizzle-orm";
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
