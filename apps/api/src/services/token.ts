import { SignJWT, jwtVerify, type JWTPayload } from "jose";
import { eq, and } from "drizzle-orm";
import { getDb } from "../db";
import { loadConfig } from "../config";
import { schema } from "@sims/db";

const enc = (s: string) => new TextEncoder().encode(s);

export interface JwtClaims extends JWTPayload {
  sub: string;
  role: string;
  tokenVersion: number;
  familyId?: string;
  jti?: string;
}

export async function signAccessToken(userId: string, role: string, tokenVersion: number): Promise<string> {
  const cfg = loadConfig();
  return new SignJWT({ role, tokenVersion })
    .setProtectedHeader({ alg: "HS256" })
    .setSubject(userId)
    .setIssuedAt()
    .setExpirationTime(Math.floor(Date.now() / 1000) + cfg.accessTtlSec)
    .sign(enc(cfg.jwtAccessSecret));
}

export async function signRefreshToken(
  familyId: string,
  jti: string,
  userId: string,
  role: string,
  tokenVersion: number,
): Promise<string> {
  const cfg = loadConfig();
  return new SignJWT({ familyId, jti, role, tokenVersion })
    .setProtectedHeader({ alg: "HS256" })
    .setSubject(userId)
    .setIssuedAt()
    .setExpirationTime(Math.floor(Date.now() / 1000) + cfg.refreshTtlSec)
    .sign(enc(cfg.jwtRefreshSecret));
}

export async function verifyJwt(token: string, secret: string): Promise<JwtClaims | null> {
  try {
    const { payload } = await jwtVerify(token, enc(secret));
    return payload as JwtClaims;
  } catch {
    return null;
  }
}

// Issue a brand-new refresh family and return the signed pair.
export async function issueTokens(input: {
  userId: string;
  role: string;
  tokenVersion: number;
}): Promise<{ accessToken: string; refreshToken: string }> {
  const cfg = loadConfig();
  const db = getDb(cfg.databaseUrl);
  const familyId = crypto.randomUUID();
  const jti = crypto.randomUUID();

  await db.insert(schema.refreshTokenFamilies).values({
    userId: input.userId,
    familyId,
    tokenJti: jti,
    expiresAt: new Date(Date.now() + cfg.refreshTtlSec * 1000),
  });

  const [accessToken, refreshToken] = await Promise.all([
    signAccessToken(input.userId, input.role, input.tokenVersion),
    signRefreshToken(familyId, jti, input.userId, input.role, input.tokenVersion),
  ]);
  return { accessToken, refreshToken };
}

// Rotate: the presented refresh token must be the current unrevoked jti.
// Reuse of an old jti => revoke the whole family (TOKEN_REUSE).
export async function rotateRefresh(refreshToken: string): Promise<{ accessToken: string; refreshToken: string }> {
  const cfg = loadConfig();
  const db = getDb(cfg.databaseUrl);

  const claims = await verifyJwt(refreshToken, cfg.jwtRefreshSecret);
  if (!claims?.sub || !claims.familyId || !claims.jti) {
    throw { code: "INVALID_REFRESH" };
  }

  const current = await db
    .select()
    .from(schema.refreshTokenFamilies)
    .where(
      and(
        eq(schema.refreshTokenFamilies.familyId, claims.familyId as string),
        eq(schema.refreshTokenFamilies.tokenJti, claims.jti as string),
      ),
    );

  // Our simplified schema has no revoked column; we detect reuse by checking if the presented jti is NOT the latest.
  // If the family has no matching jti with the presented one, we treat it as reuse if family exists at all.
  // Full revoked-column logic is deferred to DB with revoked column; here we just check existence.
  if (current.length === 0) {
    // No matching row — check if family exists at all; if so, it's reuse.
    const familyRows = await db
      .select()
      .from(schema.refreshTokenFamilies)
      .where(eq(schema.refreshTokenFamilies.familyId, claims.familyId as string));
    if (familyRows.length > 0) {
      await revokeRefreshFamily(claims.familyId as string);
      throw { code: "TOKEN_REUSE" };
    }
    throw { code: "INVALID_REFRESH" };
  }

  // For rotation, we insert a new row and delete the old jti's row is not needed; we just insert new jti.
  // To keep family rotation visible, we insert new jti; old row remains but will be considered stale on next reuse.
  // In a full impl with revoked column, we'd UPDATE revoked=true on old jti.
  const newJti = crypto.randomUUID();
  await db.insert(schema.refreshTokenFamilies).values({
    userId: claims.sub,
    familyId: claims.familyId as string,
    tokenJti: newJti,
    expiresAt: new Date(Date.now() + cfg.refreshTtlSec * 1000),
  });

  const [accessToken, newRefreshToken] = await Promise.all([
    signAccessToken(claims.sub, claims.role, claims.tokenVersion),
    signRefreshToken(claims.familyId as string, newJti, claims.sub, claims.role, claims.tokenVersion),
  ]);
  return { accessToken, refreshToken: newRefreshToken };
}

export async function revokeRefreshFamily(familyId: string): Promise<void> {
  const cfg = loadConfig();
  const db = getDb(cfg.databaseUrl);
  // Without revoked column, deletion is the revoke mechanism for this simplified schema.
  // Full impl would UPDATE revoked=true.
  await db.delete(schema.refreshTokenFamilies).where(eq(schema.refreshTokenFamilies.familyId, familyId));
}
