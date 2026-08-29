import { describe, expect, test } from "bun:test";
import { hashPassword, verifyPassword } from "../../src/modules/auth";

describe("password service", () => {
  test("hashes a password and verifies the original value", async () => {
    const hash = await hashPassword("CorrectHorseBatteryStaple1!");

    expect(hash).not.toBe("CorrectHorseBatteryStaple1!");
    expect(await verifyPassword("CorrectHorseBatteryStaple1!", hash)).toBe(true);
  });

  test("rejects a different password", async () => {
    const hash = await hashPassword("CorrectHorseBatteryStaple1!");

    expect(await verifyPassword("WrongPassword1!", hash)).toBe(false);
  });
});
