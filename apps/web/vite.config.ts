import { sveltekit } from "@sveltejs/kit/vite";
import { defineConfig } from "vite";
import tailwindcss from "@tailwindcss/vite";
import svelteConfig from "./sveltekit.config.ts";

export default defineConfig({
  plugins: [
    tailwindcss(),
    sveltekit({
      ...svelteConfig.kit,
      preprocess: svelteConfig.preprocess,
    }),
  ],
  server: {
    port: 5173,
    proxy: {
      "/api": "http://localhost:3001",
    },
  },
});
