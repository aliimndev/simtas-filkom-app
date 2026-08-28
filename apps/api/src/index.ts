import { loadConfig } from "./config";
import { createApp } from "./app";
export type { AppType } from "./app";

const cfg = loadConfig();
Bun.serve({ port: cfg.port, fetch: createApp(cfg).fetch });
console.log(`api listening on http://localhost:${cfg.port}`);
