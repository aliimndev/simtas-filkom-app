# FRONTEND_STANDARDS — SIMTAS FILKOM

Every developer / AI agent **must read and follow this document before writing or
modifying any frontend code**. Its purpose is to prevent the same problems from recurring —
most notably Tailwind IntelliSense warnings (`suggestCanonicalClasses`) and inconsistent
public-landing-page architecture. A concise summary of the core rules is also wired into
`AGENTS.md`.

---

## 1. Tailwind CSS v4 — Always Use Canonical Classes

Styling stack: **Tailwind v4** (utilities generated from `@theme inline` in
`src/app/globals.css`). Never write `*-[var(--x)]` when a canonical form exists — it triggers
IDE warnings and makes the CSS inconsistent.

### 1.1 Registered color tokens → use theme utilities

These tokens are registered as theme colors, so canonical utilities are available:

| CSS variable | Canonical utilities used in the codebase |
|---|---|
| `--st-bg` | `bg-st-bg`, `text-st-bg` |
| `--st-surface` | `bg-st-surface`, `border-st-surface` |
| `--st-surface-hi` | `bg-st-surface-hi`, `hover:bg-st-surface-hi` |
| `--st-text` | `text-st-text`, `hover:text-st-text`, `group-hover:text-st-text` |
| `--st-muted` | `text-st-muted` |
| `--st-stroke` | `border-st-stroke`, `bg-st-stroke`, `text-st-stroke`, `divide-st-stroke` |

✅ **Correct:**

```tsx
<p className="text-st-muted">description</p>
<div className="border border-st-stroke bg-st-surface">card</div>
<span className="hover:bg-st-surface-hi hover:text-st-text">link</span>
```

❌ **Incorrect** (triggers `suggestCanonicalClasses`):

```tsx
<p className="text-[var(--st-muted)]">description</p>
<div className="border border-[var(--st-stroke)] bg-[var(--st-surface)]">card</div>
<span className="hover:bg-[var(--st-surface-hi)] hover:text-[var(--st-text)]">link</span>
```

### 1.2 Unregistered tokens → Tailwind v4 parens syntax `(--var)`

`--st-accent-from` and `--st-accent-to` are **not** registered as theme colors. Use the
Tailwind v4 parens syntax, including with opacity modifiers / variants:

✅ **Correct:** `text-(--st-accent-from)`, `group-hover:text-(--st-accent-to)`,
`hover:border-(--st-accent-from)/40`, `bg-(--st-surface)/80`

❌ **Incorrect:** `text-[var(--st-accent-from)]`, `hover:border-[var(--st-accent-from)]/40`

### 1.3 Spacing & z-index → dynamic scale, not arbitrary

Tailwind v4 ships a dynamic spacing scale and container sizes. Do not write arbitrary px/rem
when a canonical equivalent exists:

| Intent | ✅ Canonical | ❌ Arbitrary |
|---|---|---|
| 32rem | `h-128`, `w-lg` (container) | `h-[32rem]`, `w-[32rem]` |
| 40rem | `h-160`, `w-160` | `h-[40rem]`, `w-[40rem]` |
| 640px min-width | `min-w-160` | `min-w-[640px]` |
| 400px min-height | `min-h-100` | `min-h-[400px]` |
| 1400px max-width | `max-w-350` | `max-w-[1400px]` |
| 80px / 40px / 120px min-height | `min-h-20` / `min-h-10` / `min-h-30` | `min-h-[80px]` / `min-h-[40px]` / `min-h-[120px]` |
| 18px icon | `h-4.5 w-4.5` | `h-[18px] w-[18px]` |
| 3px | `h-0.75` | `h-[3px]` |
| 2px translate | `-translate-y-0.5` | `translate-y-[-2px]` |
| 16.25rem inset | `left-65` | `left-[16.25rem]` |
| z-index 9999 | `z-9999` | `z-[9999]` |

Rule: any `px` value divisible by 4 → `value / 4` (e.g. `640px → 160`, `18px → 4.5`); any `rem` value divisible by `0.25` → `value × 4` (e.g. `16.25rem → 65`).

**Arbitrary values are only justified when no canonical form exists** — examples already
present in the codebase: `h-[calc(100%-3px)]`, `text-[0.7rem]`, `text-[2.6rem]`,
`blur-[120px]`, `top-[-15%]`, `tracking-[0.25em]`.

### 1.4 Rules of thumb

1. If Tailwind IntelliSense suggests `suggestCanonicalClasses` → **follow the suggestion**.
2. When unsure whether a canonical class is generated: run `npm run build`, then confirm the
   class exists in `.next/static/chunks/*.css`.
3. Do not add new utilities/classes in `globals.css` for things that already exist; use tokens.
4. Do not change classes in other files without a clear visual intent — preserve markup when
   refactoring (a refactor is not a redesign).

---

## 2. Public Landing Page Architecture

- `src/app/(public)/page.tsx` is a **Server Component** (no `'use client'`) that only
  composes sections + `metadata`. Do not turn it back into a client component.
- Interactivity is isolated as **client islands** in `src/components/features/landing/`:
  `role-cycler.tsx`, `faq-accordion.tsx`, `redirect-authenticated.tsx`, `reveal.tsx`,
  `boot-loader.tsx`.
- **One section = one file** (Server Component): `hero.tsx`, `stats.tsx`, `pipeline.tsx`,
  `features.tsx`, `roles.tsx`, `faq-section.tsx`, `cta.tsx`.
- Static data (arrays + types) is colocated with the section that owns it — **not** on the page.
- **DO NOT reimplement existing components — reuse them:**
  - `LandingButton` — pill CTA `solid`/`outline`; `href` starting with `#` renders `<a>`,
    otherwise `next/link`.
  - `FaqAccordion` — accordion with state + ARIA; home uses `variant="list"`, the `/faq`
    page uses `variant="cards"`.
  - `Reveal` — scroll reveal (IntersectionObserver, respects reduced motion).
  - `RoleCycler` — hero role animation.
- Keep client islands minimal: only components with client state/effects get `'use client'`.

---

## 3. Accessibility (Required)

- Rotating/dynamic text → `aria-live="polite"`; do not run intervals when
  `prefers-reduced-motion: reduce` is set.
- Accordions → the button must have `aria-expanded` + `aria-controls`; decorative icons get
  `aria-hidden`.
- Icon-only buttons without visible text → `aria-label`.
- One `h1` per page; headings in order (`h1 → h2 → h3`), never skip levels.
- Interactive elements must be focusable and keyboard-operable (`<button>`/`<a>`, not `<div>`).
- Click-to-dismiss modal/dropdown backdrop overlays must be `aria-hidden` (the real close
  button is the accessible dismissal path).

---

## 4. State & Store

- Use **granular Zustand selectors**: `useAuthStore((s) => s.accessToken)`.
- Do not destructure the whole store when only part of it is needed
  (`const { accessToken, isHydrated } = useAuthStore()` causes unnecessary re-renders).

---

## 5. Mandatory Validation Before Finishing Work

```bash
npm run type-check                        # tsc --noEmit
npm test                                  # jest
npx eslint <changed files/folders> --max-warnings 0
npm run build                             # for styling/RSC changes
```

For styling changes, confirm the canonical classes you used actually appear in
`.next/static/chunks/*.css` (not `.next/static/css/` — Turbopack uses this location).

---

## 6. Audit Findings (2026-08-08) — Anomalies Found & Fixed

Full `src/` audit against this standard + `CLAUDE.md`. Every error class found is documented
below as a rule — do not reintroduce them.

### 6.1 Zustand: full-store destructuring — FIXED (13 occurrences)

`const { user } = useAuthStore()` subscribes to the entire store and re-renders on any change.

✅ `const user = useAuthStore((s) => s.user)`
❌ `const { user } = useAuthStore()`

Fixed in: `use-thesis-picker`, `require-auth`, `sidebar` (Sidebar + TopBar), `thesis`,
`supervision`, `profile`, `documents`, `dashboard/{page,student,admin,examiner,supervisor}-page`.

### 6.2 Inline styles — 1 FIXED, 6 allowed

Only convert to a class when the value is static.

- FIXED: boot-loader overlay `style={{ background: 'var(--st-bg)' }}` → `bg-st-bg` class.
- ALLOWED (dynamic / not expressible as a static class): upload progress width
  (`documents/page.tsx`), data-driven bar height (`admin-page.tsx`), `animationDelay`
  (`reveal.tsx`), radial-gradient glow (login / hero), progress width + boxShadow
  (`boot-loader.tsx`).

### 6.3 Backdrop click-targets missing `aria-hidden` — FIXED (2)

Backdrop `<div onClick={...}>` overlays were announceable/odd for assistive tech. Add
`aria-hidden` when dismissal already exists via a real close button.

Fixed: `(dashboard)/layout.tsx` mobile-drawer scrim, `notification-bell.tsx` dropdown scrim.

### 6.4 Components over 200 lines — SPLIT (done 2026-08-08)

`CLAUDE.md` flags components over 200 lines. Six dashboard pages exceeded it and were split
into colocated section components (same pattern as the landing page). The page keeps only
data fetching, state, and mutations; everything presentational lives in `src/components/features/`:

| Page (before → after) | Extracted to |
|---|---|
| `(dashboard)/thesis/page.tsx` (357 → 132) | `my-thesis/` (detail card, history, submit dialog, cancel dialog, no-thesis state) |
| `(dashboard)/documents/page.tsx` (295 → 171) | `documents-page/` (upload controls, filter, row, shared types) |
| `(dashboard)/dashboard/student-page.tsx` (245 → 121) | `dashboard/` (stage stepper, pending actions, thesis info, schedule card) |
| `(dashboard)/admin/users/page.tsx` (236 → 115) | `admin-users/` (filters, table, mobile cards, shared actions, create form, pagination) |
| `(dashboard)/title-change-reviews/page.tsx` (226 → 89) | `title-change-review/` (queue, review dialog) |
| `(dashboard)/supervision/page.tsx` (212 → 127) | `supervision-page/` (form, row) |

**Rules learned from this split:**
- Extract sections, but do not change behavior/markup while refactoring (pixel-identical).
- A conditional `{data && …}` render already unmounts the subtree on close — do **not** add a
  `useEffect` + `setState` to reset form state (triggers `react-hooks/set-state-in-effect`).
  Prefer `key={data.id}` / unmount-based reset instead.
- Forms own their state (`react-hook-form`, zod schema) inside the extracted component;
  the page only passes `onSubmit` + loading/error props.

### 6.5 Decorative emoji in empty states — FIXED (1)

`CLAUDE.md` bans emoji as decoration (🎉 in empty states, 👋 in greetings). One was found in
the title-change-review empty state and removed. Empty states communicate with text + icon,
never emoji.

### 6.6 Verified clean — no action needed

- No `any` / `as any` / `@ts-ignore` anywhere.
- No `TODO` / `FIXME` / `HACK` comments.
- No hardcoded hex colors in TSX — all styling goes through design tokens.
- `role="button"` elements (file-dropzone, notification items) already implement
  `tabIndex={0}` + `Enter`/`Space` handling.

---

## 7. References

- `DESIGN.md` — design-system spec (palette, typography, radii, spacing).
- `CLAUDE.md` — UI-engineering principles (components, state, accessibility).
- `src/app/globals.css` — source of the `--st-*` tokens and the `@theme inline` mapping.
- `src/components/features/landing/` — live examples of every rule above.
