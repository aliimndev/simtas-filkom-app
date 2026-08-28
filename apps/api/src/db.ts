import { createDb, type Db } from "@sims/db";

let shared: Db | undefined;

export function getDb(url: string): Db {
  if (!shared) shared = createDb(url);
  return shared;
}
