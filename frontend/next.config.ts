import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  env: {
    NEXT_PUBLIC_API_URL: process.env.NEXT_PUBLIC_API_URL,
  },
  images: {
    remotePatterns: [
      {
        protocol: "https",
        hostname: "**.supabase.co",
      },
      // MinIO / S3-compatible storage presigned URLs
      {
        protocol: "http",
        hostname: "localhost",
      },
      {
        protocol: "https",
        hostname: "**",
      },
    ],
    unoptimized: true,
  },
  // Production security headers (Job 25 — Deployment Frontend)
  async headers() {
    return [
      {
        source: "/(.*)",
        headers: [
          { key: "X-Frame-Options", value: "DENY" },
          { key: "X-Content-Type-Options", value: "nosniff" },
          { key: "Referrer-Policy", value: "strict-origin-when-cross-origin" },
          { key: "Permissions-Policy", value: "camera=(), microphone=(), geolocation=()" },
        ],
      },
    ];
  },
  // ponytail: standalone output was for the Docker image, but it breaks Vercel's
  // build (Vercel generates its own serverless output). Keep it off for Vercel;
  // the backend Dockerfile doesn't use the Next.js standalone build.
  // Root redirect handled by the app router (src/app/page.tsx) instead of
  // next.config redirects so it can react to the auth state at runtime.
};

export default nextConfig;
