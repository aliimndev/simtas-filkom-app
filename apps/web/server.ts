import { join, normalize, resolve } from "node:path";

const port = Number(process.env.PORT ?? 4173);
const buildDirectory = resolve(process.cwd(), "build");

function safePath(pathname: string): string | null {
  let decoded: string;
  try {
    decoded = decodeURIComponent(pathname);
  } catch {
    return null;
  }

  const relativePath = decoded.replace(/^\/+/, "") || "index.html";
  const filePath = normalize(join(buildDirectory, relativePath));
  if (filePath !== buildDirectory && !filePath.startsWith(`${buildDirectory}/`)) return null;
  return filePath;
}

const server = Bun.serve({
  port,
  async fetch(request) {
    const filePath = safePath(new URL(request.url).pathname);
    if (!filePath) return new Response("Not Found", { status: 404 });

    const file = Bun.file(filePath);
    if (await file.exists()) return new Response(file);

    const fallback = Bun.file(join(buildDirectory, "index.html"));
    return (await fallback.exists()) ? new Response(fallback) : new Response("Not Found", { status: 404 });
  },
});

console.log(`web listening on http://localhost:${server.port}`);
