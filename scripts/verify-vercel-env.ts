const apiOrigin = process.env.VITE_API_ORIGIN?.trim();

if (!apiOrigin) {
  console.error(
    "VITE_API_ORIGIN is required for the Vercel frontend build. Set it in Vercel Project Settings > Environment Variables for Preview and Production.",
  );
  process.exit(1);
}

let parsed: URL;
try {
  parsed = new URL(apiOrigin);
} catch {
  console.error("VITE_API_ORIGIN must be an absolute http(s) URL.");
  process.exit(1);
}

if (parsed.protocol !== "http:" && parsed.protocol !== "https:") {
  console.error("VITE_API_ORIGIN must use http:// or https://.");
  process.exit(1);
}

console.log(`VITE_API_ORIGIN configured for ${parsed.origin}`);
