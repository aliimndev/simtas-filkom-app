import { Hono } from "hono";
import { eq } from "drizzle-orm";
import { loginSchema } from "@sims/shared";
import { getDb } from "../db";
import { loadConfig } from "../config";
import { schema } from "@sims/db";
import { issueTokens, rotateRefresh, revokeRefreshFamily, verifyJwt } from "../services/token";
import { verifyPassword } from "../services/password";
import { rateLimit } from "../middleware/rateLimit";
import { Authenticate } from "../middleware/auth";
import { RequireRole } from "../middleware/rbac";

export const authRoutes = new Hono();
authRoutes.use("/login", rateLimit({ windowMs: 60_000, max: 10 }));

authRoutes.post("/login", async (c) => {
  const body = await c.req.json().catch(() => null);
  const parsed = loginSchema.safeParse(body);
  if (!parsed.success) {
    return c.json({ error: { code: "VALIDATION", message: "Invalid email or password" } }, 400);
  }
  const cfg = loadConfig();
  const db = getDb(cfg.databaseUrl);

  const users = await db.select().from(schema.users).where(eq(schema.users.email, parsed.data.email));
  const user: any = users[0];
  // Timing-equalize: even an unknown email runs a bcrypt compare so response time
  // doesn't reveal whether the email exists (anti-enumeration).
  const DUMMY_HASH = "$2a$12$qMPO1EgF0zmpDh4W49ERVOfOxF28jsItEaiKEKZCWOL9NoKX3U7iC";
  const ok = user ? await verifyPassword(parsed.data.password, user.passwordHash) : await verifyPassword(parsed.data.password, DUMMY_HASH);
  if (!user) return c.json({ error: { code: "UNAUTHORIZED", message: "Invalid credentials" } }, 401);

  if (user.lockedUntil && new Date(user.lockedUntil).getTime() > Date.now()) {
    return c.json({ error: { code: "LOCKED", message: "Account locked" } }, 423);
  }

  if (!ok) {
    const attempts = (user.loginAttemptCount ?? 0) + 1;
    if (attempts >= 5) {
      await db
        .update(schema.users)
        .set({ loginAttemptCount: attempts, lockedUntil: new Date(Date.now() + 15 * 60 * 1000) } as any)
        .where(eq(schema.users.id, user.id));
      return c.json({ error: { code: "LOCKED", message: "Account locked after 5 failures" } }, 423);
    }
    await db.update(schema.users).set({ loginAttemptCount: attempts } as any).where(eq(schema.users.id, user.id));
    return c.json({ error: { code: "UNAUTHORIZED", message: "Invalid credentials" } }, 401);
  }

  if (user.isActive === false) {
    return c.json({ error: { code: "FORBIDDEN", message: "Account not active" } }, 403);
  }

  await db
    .update(schema.users)
    .set({ loginAttemptCount: 0, lastLoginAt: new Date() } as any)
    .where(eq(schema.users.id, user.id));

  // Resolve role name via roles table if needed; fallback to hardcoded map
  let roleName: string = user.role ?? "";
  if (!roleName && user.roleId) {
    const roleRows: any = await db.select().from(schema.roles).where(eq(schema.roles.id, user.roleId));
    roleName = roleRows[0]?.name?.toUpperCase?.() ?? roleRows[0]?.name ?? "MAHASISWA";
    // Normalize: DB stores lower-case with underscore e.g. admin_fakultas -> ADMIN_FAKULTAS
    roleName = roleName.toUpperCase();
    if (roleName === "ADMIN_FAKULTAS") roleName = "ADMIN_FAKULTAS";
  }
  if (!roleName) roleName = "MAHASISWA";

  const tokens = await issueTokens({ userId: user.id, role: roleName, tokenVersion: user.tokenVersion ?? 0 });

  // Fetch fresh role name for response
  const payload = {
    accessToken: tokens.accessToken,
    refreshToken: tokens.refreshToken,
    user: {
      id: user.id,
      email: user.email,
      fullName: user.fullName,
      role: roleName,
      mustChangePassword: Boolean(user.mustChangePassword),
    },
  };
  return c.json(payload, 200);
});

authRoutes.post("/refresh", async (c) => {
  const body = await c.req.json().catch(() => null);
  const refreshToken = body?.refreshToken;
  if (typeof refreshToken !== "string") {
    return c.json({ error: { code: "VALIDATION", message: "refreshToken required" } }, 400);
  }
  try {
    const pair = await rotateRefresh(refreshToken);
    return c.json({ accessToken: pair.accessToken, refreshToken: pair.refreshToken }, 200);
  } catch (e: any) {
    if (e?.code === "TOKEN_REUSE") {
      return c.json({ error: { code: "TOKEN_REUSE", message: "Refresh token reused; family revoked" } }, 401);
    }
    return c.json({ error: { code: "UNAUTHORIZED", message: "Invalid refresh token" } }, 401);
  }
});

authRoutes.post("/logout", async (c) => {
  const body = await c.req.json().catch(() => null);
  if (body?.refreshToken) {
    const claims = await verifyJwt(body.refreshToken, loadConfig().jwtRefreshSecret);
    if (claims?.familyId) await revokeRefreshFamily(claims.familyId as string);
  }
  return c.body(null, 204);
});

// Minimal RBAC demonstration: admin-only. 401 without token (Authenticate), 403 on wrong role (RequireRole).
authRoutes.get("/admin/ping", Authenticate(), RequireRole("ADMIN_FAKULTAS"), (c: any) => c.json({ ok: true }, 200));
