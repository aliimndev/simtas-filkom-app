<!-- BEGIN:nextjs-agent-rules -->
# This is NOT the Next.js you know

This version has breaking changes — APIs, conventions, and file structure may all differ from your training data. Read the relevant guide in `node_modules/next/dist/docs/` before writing any code. Heed deprecation notices.
<!-- END:nextjs-agent-rules -->

<!-- BEGIN:frontend-standards -->
# Frontend Standards (SIMTAS FILKOM)
READ `CLAUDE.MD`
Before writing or modifying any frontend code — especially Tailwind classes — read `FRONTEND_STANDARDS.md` and follow it. Core rules:

- **Canonical Tailwind v4 utilities, never `*-[var(--x)]`** when a canonical form exists: `text-st-muted`, `border-st-stroke`, `bg-st-surface`, `hover:bg-st-surface-hi`, `divide-st-stroke`. For unregistered tokens (`--st-accent-from`, `--st-accent-to`) use the parens syntax: `text-(--st-accent-from)`, `hover:border-(--st-accent-from)/40`.
- **Dynamic scale instead of arbitrary px/rem**: `h-128` (32rem), `w-lg` (32rem), `h-160`/`w-160` (40rem), `h-0.75` (3px), `z-9999` — not `h-[32rem]`, `w-[40rem]`, `h-[3px]`, `z-[9999]`.
- **Public landing page = Server Component** composing colocated sections in `src/components/features/landing/`; client islands only where state/effects require them. Reuse `LandingButton`, `FaqAccordion`, `Reveal`, `RoleCycler` — do not reimplement.
- **Accessibility**: `aria-live="polite"` for rotating text, `aria-expanded`/`aria-controls` on accordions, `aria-hidden` on decorative icons, respect `prefers-reduced-motion`.
- **Zustand**: use granular selectors (`useAuthStore((s) => s.accessToken)`), not full-store destructuring.
- **Validate before finishing**: `npm run type-check`, `npm test`, `npx eslint <changed> --max-warnings 0`, and `npm run build` for styling changes (verify classes land in `.next/static/chunks/*.css`).
<!-- END:frontend-standards -->
