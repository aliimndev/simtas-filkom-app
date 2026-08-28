import { afterEach, describe, expect, test } from "bun:test";
import { loadConfig, setRuntimeConfig, type Config } from "../../src/config";

const runtimeConfig: Config = {
  databaseUrl: "postgres://runtime/db",
  jwtAccessSecret: "runtime-access",
  jwtRefreshSecret: "runtime-refresh",
  accessTtlSec: 900,
  refreshTtlSec: 90 * 24 * 60 * 60,
  corsOrigin: "http://runtime.test",
  port: 3010,
  isTest: true,
};

afterEach(() => {
  setRuntimeConfig(loadConfig({ NODE_ENV: "test" }));
});

describe("configuration", () => {
  test("uses explicit environment values and test-safe secret defaults", () => {
    const config = loadConfig({
      NODE_ENV: "test",
      DATABASE_URL: "postgres://test/db",
      ACCESS_TTL_SEC: "60",
      CORS_ORIGIN: "http://test.local",
      PORT: "4000",
    });

    expect(config.databaseUrl).toBe("postgres://test/db");
    expect(config.jwtAccessSecret).toBe("test-access-secret");
    expect(config.jwtRefreshSecret).toBe("test-refresh-secret");
    expect(config.accessTtlSec).toBe(60);
    expect(config.corsOrigin).toBe("http://test.local");
    expect(config.port).toBe(4000);
    expect(config.isTest).toBe(true);
  });

  test("returns the runtime configuration when no environment override is supplied", () => {
    setRuntimeConfig(runtimeConfig);

    expect(loadConfig()).toEqual(runtimeConfig);
    expect(loadConfig({ DATABASE_URL: "postgres://explicit/db", NODE_ENV: "test" }).databaseUrl).toBe(
      "postgres://explicit/db",
    );
  });
});
