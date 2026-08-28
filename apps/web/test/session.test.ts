import { beforeEach, describe, expect, test } from "bun:test";
import { clearSession, getAccessToken, getRefreshToken, setSession } from "../src/lib/session";

beforeEach(() => {
  clearSession();
});

describe("frontend session state", () => {
  test("stores and reads both access and refresh tokens", () => {
    setSession({ accessToken: "access-token", refreshToken: "refresh-token" });

    expect(getAccessToken()).toBe("access-token");
    expect(getRefreshToken()).toBe("refresh-token");
  });

  test("preserves the refresh token when only the access token changes", () => {
    setSession({ accessToken: "first-access", refreshToken: "refresh-token" });
    setSession({ accessToken: "second-access" });

    expect(getAccessToken()).toBe("second-access");
    expect(getRefreshToken()).toBe("refresh-token");
  });

  test("clears the complete session", () => {
    setSession({ accessToken: "access-token", refreshToken: "refresh-token" });
    clearSession();

    expect(getAccessToken()).toBeNull();
    expect(getRefreshToken()).toBeNull();
  });
});
