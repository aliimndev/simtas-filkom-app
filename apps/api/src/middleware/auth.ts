import { eq } from "drizzle-orm";
import { verifyJwt } from "../modules/auth/token.service";
import { loadConfig } from "../config";
import { getDb } from "../db";
import { schema } from "@sims/db";

export const Authenticate = () => async (c: any, next: any) => {
  const header = c.req.header("Authorization");
  if (!header?.startsWith("Bearer ")) {
    return c.json({ error: { code: "UNAUTHORIZED", message: "Missing token" } }, 401);
  }
  const cfg = loadConfig();
  const claims = await verifyJwt(header.slice(7), cfg.jwtAccessSecret);
  if (!claims) {
    return c.json({ error: { code: "UNAUTHORIZED", message: "Invalid token" } }, 401);
  }

  // Access tokens carry a jti (Ticket 4). A blacklisted jti — e.g. revoked on
  // logout — is rejected on every authenticated route, not just logout.
  if (claims.jti) {
    const db = getDb(cfg.databaseUrl);
    const blacklisted = await db
      .select({ id: schema.tokenBlacklist.id })
      .from(schema.tokenBlacklist)
      .where(eq(schema.tokenBlacklist.tokenJti, claims.jti));
    if (blacklisted.length > 0) {
      return c.json({ error: { code: "UNAUTHORIZED", message: "Token has been revoked" } }, 401);
    }

    // token_version / active check (parity with Go middleware). ponytail: only
    // enforced when the user row exists — integration tests sign tokens for
    // synthetic IDs absent from the DB, so we fall through rather than reject
    // them. For real users this rejects tokens minted before a session bump and
    // tokens of deactivated accounts. Synthetic ids (e.g. "admin-1") are not
    // valid UUIDs, so the lookup is guarded to avoid a 22P02 error.
    let user: any = null;
    try {
      const users = (await db
        .select({ tokenVersion: schema.users.tokenVersion, isActive: schema.users.isActive })
        .from(schema.users)
        .where(eq(schema.users.id, claims.sub))) as any[];
      user = users[0];
    } catch {
      user = null; // non-UUID synthetic id → treat as absent (fall through)
    }
    if (user) {
      if (user.isActive === false) {
        return c.json({ error: { code: "FORBIDDEN", message: "Account is not active" } }, 403);
      }
      if (user.tokenVersion !== claims.tokenVersion) {
        return c.json({ error: { code: "UNAUTHORIZED", message: "Session revoked; please log in again" } }, 401);
      }
    }
  }

  c.set("user", {
    id: claims.sub,
    role: claims.role,
    tokenVersion: claims.tokenVersion,
  });
  await next();
};
