import adapter from "@sveltejs/adapter-static";
import { vitePreprocess } from "@sveltejs/vite-plugin-svelte";
import type { Config } from "@sveltejs/kit";

const config: Config = {
  preprocess: vitePreprocess(),
  kit: {
    adapter: adapter({ fallback: "index.html" }),
    prerender: { handleUnseenRoutes: "ignore" },
    alias: {
      $lib: "./src/lib",
      "@sims/api": "../api/src/index.ts",
      "@sims/shared": "../../packages/shared/src/index.ts",
      "@sims/db": "../../packages/db/src/index.ts",
    },
  },
};

export default config;
