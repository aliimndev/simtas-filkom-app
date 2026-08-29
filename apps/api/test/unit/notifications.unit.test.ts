import { describe, expect, test } from "bun:test";
import { normalizeLimit } from "../../src/modules/notifications";

describe("notification query normalization", () => {
  test.each([
    [undefined, 20],
    [null, 20],
    ["not-a-number", 20],
    [0, 20],
    [-5, 20],
    [1.9, 1],
    ["25", 25],
    [101, 100],
  ])("normalizes %j to %j", (input, expected) => {
    expect(normalizeLimit(input)).toBe(expected);
  });
});
