import { describe, test, expect } from "bun:test";
import { loginSchema, isRoleName, KAPRODI } from "@sims/shared";

describe("auth contract", () => {
  test("accepts valid creds, rejects short password", () => {
    expect(loginSchema.safeParse({ email: "a@example.com", password: "longenough" }).success).toBe(true);
    expect(loginSchema.safeParse({ email: "a@example.com", password: "short" }).success).toBe(false);
  });
  test("role constants are valid", () => {
    expect(isRoleName(KAPRODI)).toBe(true);
    expect(isRoleName("VISITOR")).toBe(false);
  });
});
