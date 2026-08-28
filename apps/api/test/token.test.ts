import { describe, expect, it } from "bun:test";
import { signAccessToken, verifyJwt } from "../src/services/token";
import { loadConfig } from "../src/config";
import { MAHASISWA } from "@sims/shared";

const cfg = loadConfig({ NODE_ENV: "test" });

describe("token service", () => {
  it("issues access JWT with sub + role + tokenVersion claims", async () => {
    const accessToken = await signAccessToken(
      "00000000-0000-0000-0000-000000000001",
      MAHASISWA,
      0,
    );
    const claims = await verifyJwt(accessToken, cfg.jwtAccessSecret);
    expect(claims?.role).toBe(MAHASISWA);
    expect(claims?.sub).toBe("00000000-0000-0000-0000-000000000001");
    expect(claims?.tokenVersion).toBe(0);
    // Ticket 4: access tokens must carry a jti so they can be blacklisted on logout.
    expect(typeof claims?.jti).toBe("string");
    expect(claims?.jti?.length).toBeGreaterThan(0);
  });

  it("verifyJwt returns null for a bad token", async () => {
    expect(await verifyJwt("not-a-jwt", cfg.jwtAccessSecret)).toBeNull();
  });
});
