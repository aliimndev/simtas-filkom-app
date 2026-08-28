import { describe, expect, it } from "bun:test";
import { signAccessToken, verifyJwt } from "../src/services/token";
import { loadConfig } from "../src/config";
import { ROLE_MAHASISWA } from "@sims/shared";

const cfg = loadConfig({ NODE_ENV: "test" });

describe("token service", () => {
  it("issues access JWT with sub + role + tokenVersion claims", async () => {
    const accessToken = await signAccessToken(
      "00000000-0000-0000-0000-000000000001",
      ROLE_MAHASISWA,
      0,
    );
    const claims = await verifyJwt(accessToken, cfg.jwtAccessSecret);
    expect(claims?.role).toBe(ROLE_MAHASISWA);
    expect(claims?.sub).toBe("00000000-0000-0000-0000-000000000001");
    expect(claims?.tokenVersion).toBe(0);
  });

  it("verifyJwt returns null for a bad token", async () => {
    expect(await verifyJwt("not-a-jwt", cfg.jwtAccessSecret)).toBeNull();
  });
});
