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

let runtimeConfig: Config | undefined;

export function setRuntimeConfig(config: Config): void {
  runtimeConfig = config;
}

export function loadConfig(env?: Record<string, string | undefined>): Config {
  if (!env && runtimeConfig) return runtimeConfig;
  const source = env ?? process.env;
  return {
    databaseUrl: source.DATABASE_URL ?? "postgres://simtas:simtas@localhost:5432/simtas",
    jwtAccessSecret:
      source.JWT_ACCESS_SECRET ??
      (source.NODE_ENV === "test" ? "test-access-secret" : required("JWT_ACCESS_SECRET")),
    jwtRefreshSecret:
      source.JWT_REFRESH_SECRET ??
      (source.NODE_ENV === "test" ? "test-refresh-secret" : required("JWT_REFRESH_SECRET")),
    accessTtlSec: Number(source.ACCESS_TTL_SEC ?? 15 * 60),
    refreshTtlSec: 90 * 24 * 60 * 60,
    corsOrigin: source.CORS_ORIGIN ?? "http://localhost:5173",
    port: Number(source.PORT ?? 3001),
    isTest: source.NODE_ENV === "test",
  };
}

function required(k: string): string {
  throw new Error(`Missing env var ${k}`);
}
