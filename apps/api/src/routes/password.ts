import { Hono } from "hono";
import { z } from "zod";
import { eq, sql } from "drizzle-orm";
import { getDb } from "../db";
import { loadConfig } from "../config";
import { schema } from "@sims/db";
import { hashPassword } from "../services/password";

// ponytail: forgot/reset business logic kept inline (services/password.ts is
// locked — it holds the hashPassword primitive — so we don't re-create it).

export const passwordRoutes = new Hono();

const forgotSchema = z.object({ email: z.string().email() });

const resetSchema = z.object({
  token: z.string().min(1),
  newPassword: z.string().min(8),
  confirmPassword: z.string().min(1),
});

const db = () => getDb(loadConfig().databaseUrl);

// POST /api/v1/auth/password/forgot
passwordRoutes.post("/forgot", async (c) => {
  const body = await c.req.json().catch(() => null);
  const parsed = forgotSchema.safeParse(body);
  if (!parsed.success) return c.json({ error: { code: "VALIDATION", message: "Invalid email" } }, 400);

  const email = parsed.data.email.toLowerCase().trim();
  const d = db();
  const users: any = await d.select().from(schema.users).where(eq(schema.users.email, email));

  // Anti-enumeration: always 200; only when the email exists do we mint a token + email log.
  if (users[0]) {
    const token = crypto.randomBytes(32).toString("hex");
    await d.insert(schema.passwordResetTokens).values({
      userId: users[0].id,
      token,
      expiresAt: new Date(Date.now() + 60 * 60 * 1000),
    });
    await d.insert(schema.emailLogs).values({
      recipientEmail: email,
      eventType: "password_reset",
      subject: "Reset Password SIMTAS",
      status: "sent",
    });
  }

  return c.json({ message: "If the email is registered, a reset link has been sent" }, 200);
});

// POST /api/v1/auth/password/reset
passwordRoutes.post("/reset", async (c) => {
  const body = await c.req.json().catch(() => null);
  const parsed = resetSchema.safeParse(body);
  if (!parsed.success) return c.json({ error: { code: "VALIDATION", message: "Invalid request" } }, 400);
  if (parsed.data.newPassword !== parsed.data.confirmPassword) {
    return c.json({ error: { code: "VALIDATION", message: "Passwords do not match" } }, 400);
  }
  if (!/[A-Z]/.test(parsed.data.newPassword) || !/[0-9]/.test(parsed.data.newPassword)) {
    return c.json({ error: { code: "VALIDATION", message: "Password must contain an uppercase letter and a digit" } }, 400);
  }

  const d = db();
  const tokens: any = await d
    .select()
    .from(schema.passwordResetTokens)
    .where(eq(schema.passwordResetTokens.token, parsed.data.token));
  const prt = tokens[0];
  if (!prt || prt.usedAt || new Date(prt.expiresAt).getTime() < Date.now()) {
    return c.json({ error: { code: "VALIDATION", message: "Invalid or expired token" } }, 400);
  }

  const passwordHash = await hashPassword(parsed.data.newPassword);
  await d
    .update(schema.users)
    .set({ passwordHash, mustChangePassword: false } as any)
    .where(eq(schema.users.id, prt.userId));
  await d
    .update(schema.users)
    .set({ tokenVersion: sql`token_version + 1` } as any)
    .where(eq(schema.users.id, prt.userId));
  await d
    .update(schema.passwordResetTokens)
    .set({ usedAt: new Date() })
    .where(eq(schema.passwordResetTokens.id, prt.id));

  return c.json({ message: "Password reset successful" }, 200);
});
