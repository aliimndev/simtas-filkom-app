import { createDb, type Db } from "@sims/db";

// ponytail: cache one connection per URL so health can probe an alternate DB
// (e.g. an unreachable URL in tests) instead of reusing the first caller's pool.
const pool = new Map<string, Db>();

export function getDb(url: string): Db {
  let db = pool.get(url);
  if (!db) {
    db = createDb(url);
    pool.set(url, db);
  }
  return db;
}
