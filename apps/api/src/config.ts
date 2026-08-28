export type Config = {
  databaseUrl: string;
  jwtAccessSecret: string;
  jwtRefreshSecret: string;
  accessTtlSec: number;
  refreshTtlSec: number;
  corsOrigin: string;
  port: number;
  isTest: boolean;
};

export function loadConfig(env: Record<string, string | undefined> = process.env): Config {
  return {
    databaseUrl: env.DATABASE_URL ?? "postgres://simtas:simtas@localhost:5432/simtas",
    jwtAccessSecret:
      env.JWT_ACCESS_SECRET ?? (env.NODE_ENV === "test" ? "test-access-secret" : required("JWT_ACCESS_SECRET")),
    jwtRefreshSecret:
      env.JWT_REFRESH_SECRET ?? (env.NODE_ENV === "test" ? "test-refresh-secret" : required("JWT_REFRESH_SECRET")),
    accessTtlSec: Number(env.ACCESS_TTL_SEC ?? 15 * 60),
    refreshTtlSec: 90 * 24 * 60 * 60,
    corsOrigin: env.CORS_ORIGIN ?? "http://localhost:5173",
    port: Number(env.PORT ?? 3001),
    isTest: env.NODE_ENV === "test",
  };
}

function required(k: string): string {
  throw new Error(`Missing env var ${k}`);
}
