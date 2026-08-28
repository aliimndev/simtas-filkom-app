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

// Rotate: the presented refresh token must be the current valid jti for its family.
// Reuse of an already-rotated jti => revoke the whole family (TOKEN_REUSE).
// ponytail: schema has no revoked column (see migration 000015); rotation keeps exactly one
// current jti per family by deleting the used row before minting the next. Replay then finds no row.
export async function rotateRefresh(refreshToken: string): Promise<{ accessToken: string; refreshToken: string }> {
  const cfg = loadConfig();
  const db = getDb(cfg.databaseUrl);

  const claims = await verifyJwt(refreshToken, cfg.jwtRefreshSecret);
  if (!claims?.sub || !claims.familyId || !claims.jti) {
    throw { code: "INVALID_REFRESH" };
  }
  const familyId = claims.familyId as string;
  const jti = claims.jti as string;

  // The only valid row is the current jti. If it's gone, this token was already used/rotated.
  const current = await db
    .select()
    .from(schema.refreshTokenFamilies)
    .where(and(eq(schema.refreshTokenFamilies.familyId, familyId), eq(schema.refreshTokenFamilies.tokenJti, jti)));

  if (current.length === 0) {
    const familyRows = await db
      .select()
      .from(schema.refreshTokenFamilies)
      .where(eq(schema.refreshTokenFamilies.familyId, familyId));
    if (familyRows.length > 0) {
      await revokeRefreshFamily(familyId);
      throw { code: "TOKEN_REUSE" };
    }
    throw { code: "INVALID_REFRESH" };
  }

  // Rotate: drop the presented (now-spent) jti, mint a fresh one for the same family.
  await db
    .delete(schema.refreshTokenFamilies)
    .where(and(eq(schema.refreshTokenFamilies.familyId, familyId), eq(schema.refreshTokenFamilies.tokenJti, jti)));
  const newJti = crypto.randomUUID();
  await db.insert(schema.refreshTokenFamilies).values({
    userId: claims.sub,
    familyId,
    tokenJti: newJti,
    expiresAt: new Date(Date.now() + cfg.refreshTtlSec * 1000),
  });

  const [accessToken, newRefreshToken] = await Promise.all([
    signAccessToken(claims.sub, claims.role, claims.tokenVersion),
    signRefreshToken(familyId, newJti, claims.sub, claims.role, claims.tokenVersion),
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
