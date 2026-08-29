import bcrypt from "bcryptjs";

export function hashPassword(p: string): Promise<string> {
  return bcrypt.hash(p, 12);
}
export function verifyPassword(p: string, hash: string): Promise<boolean> {
  return bcrypt.compare(p, hash);
}
