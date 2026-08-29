import { isIP } from "node:net";
import type { Db } from "@sims/db";
import { schema } from "@sims/db";
import type { DefenseActor } from "./types";

function validIp(value?: string | null): string | null {
  const candidate = value?.split(",")[0]?.trim();
  return candidate && isIP(candidate) ? candidate : null;
}

export async function recordAudit(
  connection: Pick<Db, "insert">,
  actor: DefenseActor,
  action: string,
  entityId: string,
  oldValue: unknown,
  newValue: unknown,
) {
  await connection.insert(schema.auditLogs).values({
    userId: actor.userId,
    action,
    entityType: "defense",
    entityId,
    oldValue: oldValue as any,
    newValue: newValue as any,
    ipAddress: validIp(actor.ipAddress),
    userAgent: actor.userAgent ?? null,
  } as any);
}
