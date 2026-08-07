---
name: SIMTAS FILKOM — Light Editorial Academic System
description: "A light editorial, software-craft academic interface for the Fakultas Ilmu Komputer Universitas Djuanda thesis platform. Two surfaces: the public landing runs a warm-neutral canvas (#f7f8fa), ink text (#17191d) on white cards with hairline strokes, an Instrument Serif display voice for headings, Inter for body, a monospace micro-label register, and a single institutional FILKOM blue accent gradient (#4a7fb0 → #2b5f9e). The authenticated dashboard is an Operate-mode task console: white cards on the neutral canvas, hairline borders, small radii (8px cards / 6px controls), one sans family (Inter), the FILKOM blue as a single restrained accent, and an 8px spacing grid."
colors:
  ink: "#17191d"
  ink-muted: "#5f6a76"
  canvas: "#f7f8fa"
  surface: "#ffffff"
  surface-hi: "#f0f3f7"
  card: "#ffffff"
  dashboard-surface-hi: "#f3f5f8"
  hairline: "#e3e8ee"
  hairline-strong: "#dde3ea"
  primary: "#4e85bf"
  primary-deep: "#2b5f9e"
  accent-from: "#4a7fb0"
  accent-to: "#2b5f9e"
  success: "#1f9d55"
  warning: "#c7770a"
  danger: "#dc2626"
  success-50: "#ecfdf3"
  warning-50: "#fffaf0"
  danger-50: "#fef2f2"
  primary-50: "#eef4fb"
  primary-100: "#dce9f7"
  primary-200: "#c3d7ee"
  primary-700: "#35618d"
  primary-800: "#2a4e73"
  primary-900: "#203a59"
typography:
  display:
    fontFamily: "Instrument Serif, ui-serif, Georgia, serif"
    fontWeight: 400
    lineHeight: 1.0
    letterSpacing: "-0.02em"
  display-italic:
    fontFamily: "Instrument Serif, ui-serif, Georgia, serif"
    fontWeight: 400
    letterSpacing: "-0.02em"
  body:
    fontFamily: "Inter, ui-sans-serif, system-ui, sans-serif"
    fontSize: "16px"
    fontWeight: 400
    lineHeight: 1.5
  body-muted:
    fontFamily: "Inter, ui-sans-serif, system-ui, sans-serif"
    fontSize: "14px"
    fontWeight: 400
    lineHeight: 1.6
  dashboard-title:
    fontFamily: "Inter, ui-sans-serif, system-ui, sans-serif"
    fontSize: "20px"
    fontWeight: 600
    letterSpacing: "-0.01em"
  dashboard-value:
    fontFamily: "Inter, ui-sans-serif, system-ui, sans-serif"
    fontSize: "24px"
    fontWeight: 600
    fontFeature: "tnum"
  micro-label:
    fontFamily: "ui-monospace, SFMono-Regular, Menlo, monospace"
    fontSize: "12px"
    fontWeight: 500
    letterSpacing: "0.25em"
  button:
    fontFamily: "Inter, ui-sans-serif, system-ui, sans-serif"
    fontSize: "14px"
    fontWeight: 500
rounded:
  sm: "10px"
  md: "16px"
  lg: "20px"
  pill: "9999px"
  dashboard-sm: "6px"
  dashboard-md: "8px"
spacing:
  section-lg: "96px"
  section: "64px"
  section-sm: "40px"
  gap: "16px"
  inner: "24px"
  dashboard-step: "8px"
  dashboard-section: "24px"
components:
  button-solid:
    backgroundColor: "{colors.ink}"
    textColor: "{colors.canvas}"
    typography: "{typography.button}"
    rounded: "{rounded.pill}"
    padding: "12px 28px"
  button-outline:
    backgroundColor: "{colors.surface}"
    textColor: "{colors.ink}"
    typography: "{typography.button}"
    rounded: "{rounded.pill}"
    padding: "12px 28px"
  card:
    backgroundColor: "{colors.surface}"
    textColor: "{colors.ink}"
    rounded: "{rounded.md}"
    padding: "24px"
  dashboard-card:
    backgroundColor: "{colors.card}"
    textColor: "{colors.ink}"
    rounded: "{rounded.dashboard-md}"
    padding: "16px"
  nav-pill:
    backgroundColor: "{colors.surface}"
    textColor: "{colors.ink-muted}"
    rounded: "{rounded.pill}"
  tech-pill:
    backgroundColor: "{colors.surface}"
    textColor: "{colors.ink}"
    rounded: "{rounded.pill}"
    padding: "6px 16px"
---

## Overview

SIMTAS FILKOM ships a **light editorial academic** interface. The canvas is a soft paper tone built around a single institutional blue — not a dark product surface and not a consumer-app wash. The visual language reads like software-craft documentation for a computer-science faculty: a serif display voice, hairline-stroked white cards, monospace micro-labels, and one blue gradient used as the single chromatic accent. Light was chosen deliberately for an academic audience (prolonged reading, formal trust) with a FILKOM blue that never becomes decorative.

The identity is grounded in the product truth: a thesis-management platform for the Fakultas Ilmu Komputer Universitas Djuanda. It deliberately avoids the three generic AI-slop looks — cream+terracotta, near-black acid-green, and broadsheet hairline — in favor of the FILKOM blue + computer-science register, paired with a carefully chosen serif display voice.

The authenticated dashboard is a separate **Operate surface** with its own register: a calm academic task console. White cards sit on the neutral canvas, separated by hairlines rather than shadows; everything is one sans family (Inter) — no serif display and no mono costume in the chrome; the FILKOM blue is a single restrained accent (active nav, primary actions, links, focus, the stage stepper); and a strict 8px spacing grid keeps high information density without crowding. Where the landing persuades, the dashboard operates: the visitor reads status, pending actions, and schedule in seconds and acts. The two surfaces share the neutral ladder and the one accent; they differ in type voice, radius, and ornament budget.

## Colors

Colors are grouped by role, not hue. The neutral ladder carries the page; blue is reserved.

**The Thresholds Rule.** Text sits on a hairline-divided neutral ladder, and every text/background pair clears WCAG AA. Ink on canvas and on white is near-black (#17191d); muted body and captions are a mid-gray blue (#5f6a76) that still clears 4.5:1 on white. The light accent blue is only ever used at 4.5:1+ on a light surface, never as body paragraph fill.

**The Scarce-Blue Rule.** The institutional blue gradient is the single chromatic accent and is used sparingly: brand mark, transition label, link emphasis, focus/hover affordances, icon tint. It never fills a section or covers a card. Pick the deep end (#2b5f9e) for interactive affordances and the light end (#4a7fb0) for icons and bullet markers, and never use the light end for small body text.

**The Dashboard Accent Rule.** Inside the dashboard, blue is an operative color, not a decorative one: it marks active navigation, the primary action, links, keyboard focus, and the completed stage in the progress stepper. It never colors a stat card icon or fills a card. Status lives in the semantic set (success/warning/danger) only; everything else is the neutral ladder.

- **Ink** ({colors.ink} #17191d): headings, primary body, solid-button fill. The default foreground.
- **Ink-muted** ({colors.ink-muted} #5f6a76): secondary text, captions, footer, placeholder register — always on the light neutral ladder, tinted from the blue rather than pure gray.
- **Canvas** ({colors.canvas} #f7f8fa): page background and the background revealed inside cards and the logo mark.
- **Surface** ({colors.surface} #ffffff): cards, floating nav pill, raised panels, tech pills.
- **Surface-hi** ({colors.surface-hi} #f0f3f7): hover-and-lift states above `surface`, subtle emphasis fills.
- **Dashboard card** ({colors.card} #ffffff) and **dashboard surface-hi** ({colors.dashboard-surface-hi} #f3f5f8): the dashboard's white card surface on the neutral canvas, and the faint fill used for list rows, hover states, and inactive stepper steps.
- **Hairline** ({colors.hairline} #e3e8ee): 1px strokes — card borders, dividers, table rules, outline controls.
- **Hairline-strong** ({colors.hairline-strong} #dde3ea): stronger separators and placeholder borders.
- **Primary / Primary-deep** ({colors.primary} #4e85bf / {colors.primary-deep} #2b5f9e): FILKOM institutional blue and its deep anchor.
- **Accent-from / Accent-to** ({colors.accent-from} #4a7fb0 → {colors.accent-to} #2b5f9e): the single blue gradient on the brand mark, CTA gradient ring, and accent text/icons.
- **Semantic**: success (#1f9d55), warning (#c7770a), danger (#dc2626), each with an ultra-light 50-tint for status chips and badge fills. These appear on dashboards and status pills; marketing surfaces stay neutral except the blue accent.

## Typography

Two families, clearly separated by job.

**The Serif-Display Rule.** Display type is `Instrument Serif` (weight 400, tight negative tracking, line-height ~1.0), with an italic variant reserved for one or two emphasis words inside a heading (`<span class="italic">`). The serif is the product's voice; it does not set body text. Display runs up to the 7xl scale on desktop and scales down responsibly; keep max display ~4rem–4.5rem on wide layouts.

**The Mono-Label Register.** Monospace (`ui-monospace`) micro-labels — uppercase, wide letter-spacing (~0.25em), ~12px — introduce sections and mark "system" voice (eyebrow labels, the `SIMTAS://FILKOM` tag, terminal boot lines, "scroll"). This is a deliberate software-craft register, used for structural labels and system chrome, never for paragraphs.

- **Display** ({typography.display}, Instrument Serif 400): landing headings, hero, section openers, the role cycler, marquee words.
- **Display-italic** ({typography.display-italic}): emphasis spans inside headings ("Tugas Akhir", "lakukan", "peran").
- **Body** ({typography.body}, Inter 16/1.5): feature/pengguna copy, FAQ answers, footer prose.
- **Body-muted** ({typography.body-muted}, Inter 14/1.6): card descriptions, captions, meta text — always `ink-muted` or on a controlled surface.
- **Micro-label** ({typography.micro-label}, mono uppercase): section eyebrows, nav/footer category headings, terminal boot labels.
- **Button** ({typography.button}, Inter 14/500).

Headings use obvious weight/size steps. Measure is held between ~65–75ch on max-width containers (`max-w-2xl`/`max-w-3xl`).

**The Dashboard-Type Rule.** The dashboard is a single-family surface: Inter everywhere, fixed rem steps (a compact page title at 20px/600 for greeting headers, values at 24px/600 tabular-nums, labels at 14px/400). No serif display and no monospace labels inside the authenticated chrome; the mono register belongs to the landing only. A fixed scale serves a task surface better than clamp-sized fluid type.

## Layout

A restrained, editorial column:

- Content sits in a **landing container** (`max-width: 80rem`, centered, `padding-inline: 1.5rem`) with generous section rhythm: desktop sections run `py-20`–`py-28`, subsections `py-16`, and the container is narrowed (`max-w-3xl`) for prose pages (about/faq/contact/guide).
- Sections alternate between **canvas** (page background) and **surface** (a lifted white band with a hairline top/bottom rule) to separate passage without full-bleed panels. Most sections are split by a single 1px hairline rule.
- The hero uses a faint blue **grid pattern** (`landing-grid-bg`) and a soft blue **glow** behind the headline — the only atmospheric element, kept at low opacity so the canvas stays clean.
- Pairs/groups use a `gap-4` (16px) grid; distinct sections separate generously. More space above a heading than below it.
- Animated reveals (`st-reveal-up`) fade elements up once on scroll; the hero role title cycles on a timed timeline.

**Grid patterns:** feature bento is 3-up (→2-up →1-up), role/pengguna cards 2-up, info cards 2-up, tech pills wrap. Below 768px everything collapses to a single column and display scales down.

**Dashboard layout.** The authenticated shell is white chrome on the neutral canvas: a 256px white sidebar (hairline right border) with brand, sectioned nav, and a profile footer; a 48px white top bar with the notification bell and a user dropdown; and a centered content column (`max-w-1400px`) on the canvas. Content follows a strict 8px grid — cards `p-6`, list rows `p-4`, `space-y-6` between blocks, `gap-4` stat grids (4-up → 2-up → 1-up). Below 1024px the sidebar collapses to a mobile drawer opened by the top-bar hamburger; the top bar keeps the bell and user menu. Responsive behavior is structural (collapse, reflow), never fluid typography.

## Elevation & Depth

Flat and hairline-based — this is an editorial surface, not a skeuomorphic stack.

**The Hairline-Lift Rule.** Depth reads through standard elevation: surfaces sit one step above the canvas and separate with a 1px hairline stroke; a row of identical cards is a row, not a pyramid. There are no hard offset shadows, no gradient spotlight tiles, and no nested content pyramids. The only shadows are the floating nav pill's soft drop (on scroll) and the subtle default card shadow — both with an offset and blur.

Cards are white on the paper canvas; the boundary is the hairline, not an ambient shadow. Approved emphasis is a soft dark shadow at low alpha for the floating header only.

**The Dashboard Flat Rule.** The dashboard is flat by default: cards and the chrome carry no shadows at rest. Depth comes from the white-on-canvas surface split and the hairline stroke. The only soft shadows are overlay layers — the mobile drawer scrim and dropdowns/popovers — where separation from content demands them.

## Shapes

Radius is generous but not bubbly.

- **Pills** ({rounded.pill} 9999px): all buttons, the floating nav, tech pills, small chips, avatar marks. This is the interactive/button shape.
- **Cards** ({rounded.md} 16px, via `rounded-2xl`): the primary container shape for feature/role/info/FAQ blocks.
- Smaller inner chips and logos use ~10px; section containers stay 16px. Mixed nested radius steps down consistently (a nested chip is smaller-radius than its parent card).

**Dashboard radii.** The dashboard dials radii down to stay quiet and dense: cards at 8px (`rounded-lg`), buttons, inputs, list rows, icon chips, and dropdowns at 6px (`rounded-md`), avatars and status badges remain full pills. A row of small controls reads as tools, not as a bubble-wall.


## Components

- **Solid button** ({components.button-solid}): pulled pill on `ink` with light text — the primary action. Used for "Masuk", "Masuk ke Sistem", "Buka Dashboard", "Hubungi Kami" (solid variants). Optional `accent-ring` gradient ring appears on hover/focus.
- **Outline button** ({components.button-outline}): white pill with a hairline stroke and `ink` text — the secondary action. Uses solid hairline border (never a colored 4px side) on the light canvas so it reads as an interactive control. Hover shifts the fill to `surface-hi` and tints the stroke to the accent.
- **Card** ({components.card}): white `rounded-2xl` panel with a 1px hairline border, `ink` heading and `ink-muted` body. Optionally transition a stroke to the accent gradient on hover (feature cards). The internal icon chip is a smaller-stroke rounded square tinted with the accent.
- **Floating nav pill** ({components.nav-pill}): fixed centered pill, `surface` at ~80% opacity with `backdrop-blur`, hairline border, mono brand tag (`sf` serif mark + "SIMTAS FILKOM"). Links are `ink-muted`, hover to `ink` on `surface-hi`.
- **FAQ accordion**: white bordered `rounded-2xl` rows; question is `ink` 500, chevron rotates, answer sits under a hairline rule in `ink-muted`. Open question carries a top rule, no colored side border.
- **Tech pill** ({components.tech-pill}): outlined pill chip listing stack names.
- **Eyebrow micro-label**: mono uppercase `micro-label` in `ink-muted`, sometimes with a short hairline dash, above display headings; part of the terminal register.
- **Boot loader**: a full-screen canvas overlay (`SIMTAS://FILKOM`), mono init lines, a large serif percentage counter, a hairline track filled by the accent gradient. Runs once per browser session, skipped under reduced-motion.

### Dashboard components

- **Sidebar / top bar**: white surfaces with a hairline boundary. Sidebar width 256px; active nav item is a light primary tint (`primary-50` fill, `primary-800` text), inactive is muted text with a faint hover fill. Section labels are small uppercase sans (never mono). The top bar is 48px: notification bell + a compact user dropdown (avatar initial, profile, logout).
- **Stat card**: a quiet 8px-radius card — uppercase 11px muted label, a 24px/600 tabular value, and a small 16px muted icon. No colored icon tiles and no accent on the value; hover (when it is a link) is a primary hairline tint only.
- **List row**: the dashboard's workhorse — `rounded-md`, hairline border, faint neutral fill (`surface-hi`/`muted` at low alpha), 16px padding, 8px column gap. Warning rows swap the fill/border to the warning tint.
- **Dashboard header**: compact greeting (`20px` semibold), a muted role/meta line, and a quiet date on the right. No emoji, no kicker, no hero.

Every interactive dashboard component carries default, hover, focus (visible ring or outline), disabled, loading, and empty/error states; loading uses skeletons, empty states teach the action (e.g. the no-thesis CTA).

## Do's and Don'ts

### Do

- Keep the paper-light canvas and reserve the FILKOM blue gradient for brand, focus, links, and icon tint — never decorate a whole section.
- Give every secondary control a visible hairline boundary on the light canvas; a borderless white button on paper is a defect, not minimalism.
- Use the serif only for display, the mono register only for micro-labels/system chrome, and Inter for body.
- Alternate canvas and lifted white sections with hairline rules to establish rhythm.
- Keep every text/surface pair at WCAG AA (≥4.5:1 for body; ≥3:1 for large display).
- Preserve the terminal/software-craft accent (boot loader, mono labels) — it is the faculty's character.

### Don't

- Don't ship a dark or near-black canvas; this product is deliberately light.
- Don't use gradient text for emphasis — emphasis comes from the serif italic + weight.
- Don't add a second chromatic accent (green, orange, pink) to marketing surfaces.
- Don't use hard offset shadows, spotlight cards, or atmospheric gradients beyond the single hero glow and grid.
- Don't render cards as same-size icon+heading+text rows by default — vary the composition (list + pair + bento) as sections do.
- Don't let the light accent blue carry small body text; keep it at AA contrast or push to the deep end.
- Don't remove the hairline from outline controls.

### Dashboard

- **Do** keep the dashboard flat and hairline-based: white cards on the canvas, no rest-state shadows, no colored icon tiles, no gradient fills.
- **Do** keep the single Inter family and the fixed rem scale; a dashboard with display serifs or mono labels reads as a landing page in costume.
- **Do** follow the 8px spacing grid — rows at 16px padding, 8px list gaps, 24px section rhythm — and keep stat values tabular for scanning.
- **Do** keep the FILKOM blue for operative meaning only: active nav, primary action, links, focus, the completed stage. Status colors belong to badges, never to card decoration.
- **Don't** use emoji as decoration (👋 in greetings, 🎉 in empty states) or glowing status dots; the calm register is the product's trust signal.
- **Don't** invent charts or sparklines where the data is real but thin — the admin funnel renders real counts with a single primary hue, graduated stage in success.
- **Don't** ship a modal, a gradient, or a glass effect for a task a plain card with a hairline can carry.

## Responsive Behavior

| Width | Behavior |
|---|---|
| ≥1024px | Feature bento 3-up; hero display at full scale; content in 80rem container |
| 768–1024px | Bento → 2-up; role/info cards 2-up |
| <768px | Single column; hamburger nav; display scales toward ~2.6rem; grids collapse |

Touch controls hold ≥40px tap height; nav pill collapses to a menu button below 768px.

