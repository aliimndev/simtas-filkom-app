import { drizzle, type PostgresJsDatabase } from "drizzle-orm/postgres-js";
import postgres from "postgres";
import * as schema from "./schema";
export { schema };

export function createDb(url: string) {
  return drizzle(postgres(url, { max: 10 }), { schema });
}
export type Db = PostgresJsDatabase<typeof schema>;
