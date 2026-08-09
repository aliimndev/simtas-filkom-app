---
name: make-interfaces-feel-better
description: >-
  Design engineering principles for making interfaces feel polished. Use when building UI components, reviewing frontend code, implementing animations, hover states, shadows, borders, typography, icons, micro-interactions, enter/exit animations, or any visual detail work. Supports quick and full review modes. Triggers on UI polish, design details, "make it feel better", "feels off", stagger animations, border radius, optical alignment, font smoothing, tabular numbers, image outlines, box shadows, icons, icon stroke weight, icon states, motion restraint.
---

# Details that make interfaces feel better

Great interfaces rarely come from a single thing. It's usually a collection of small details that compound into a great experience. Apply these principles when building or reviewing UI code. Before suggesting or writing a fix, identify the project's existing styling system and express the change in that system: Tailwind in a Tailwind project, plain CSS in a CSS project, or the established CSS-in-JS approach. Never introduce a second styling system just to apply a polish fix.

When reviewing, slow the interface down: replay motion at 10% speed in the browser's Animations panel and walk every state: hover, focus, active, loading, empty. What feels off at 10% speed is what's subtly wrong at full speed.

## Quick Reference

| Category | When to Use |
| --- | --- |
| [Typography](typography.md) | Text wrapping, font smoothing, tabular numbers |
| [Surfaces](surfaces.md) | Border radius, optical alignment, shadows, image outlines, hit areas |
| [Animations](animations.md) | Interruptible animations, enter/exit transitions, icon animations, scale on press, motion restraint |
| [Icons](icons.md) | Icon stroke weight, states via `currentColor`, outline vs fill, sizing, RTL flipping |
| [Performance](performance.md) | Transition specificity, `will-change` usage |

## Core Principles

### 1. Concentric Border Radius

Outer radius = inner radius + padding. Mismatched radii on nested elements is the most common thing that makes interfaces feel off.

### 2. Optical Over Geometric Alignment

When geometric centering looks off, align optically. Buttons with icons, play triangles, and asymmetric icons all need manual adjustment.

### 3. Shadows for Elevation, Borders for Structure

For buttons, cards, and containers whose border exists only to create depth, prefer layered transparent `box-shadow` values. Keep borders that communicate structure or state: dividers, layout separators, and selected or focus states.

### 4. Interruptible Animations

Use CSS transitions for interactive state changes — they can be interrupted mid-animation. Reserve keyframes for staged sequences that run once.

### 5. Split and Stagger Enter Animations

For an infrequent staged entrance where sequence helps communicate hierarchy, break content into semantic chunks and stagger them by ~100ms instead of animating one container. Do not stagger routine, high-frequency interactions.

### 6. Subtle Exit Animations

Use a small fixed `translateY` instead of full height. Exits should be softer than enters. Use `ease-out` for both enter and exit transitions.

### 7. Contextual Icon Animations

Animate icons with `opacity`, `scale`, and `blur` instead of toggling visibility. Use exactly these values: scale from `0.25` to `1`, opacity from `0` to `1`, blur from `4px` to `0px`. If the project has `motion` or `framer-motion` in `package.json`, match that package's import path (or the established nearby imports when both exist) and use `transition: { type: "spring", duration: 0.3, bounce: 0 }` — bounce must always be `0`. If no motion library is installed, keep both icons in the DOM (one absolute-positioned) and cross-fade with CSS transitions using `cubic-bezier(0.2, 0, 0, 1)` — this gives both enter and exit animations without any dependency.

### 8. Font Smoothing

Apply `-webkit-font-smoothing: antialiased` to the root layout on macOS for crisper text.

### 9. Tabular Numbers

Use `font-variant-numeric: tabular-nums` for any dynamically updating numbers to prevent layout shift.

### 10. Text Wrapping

Use `text-wrap: balance` on headings. Use `text-wrap: pretty` for body text to avoid orphans.

### 11. Image Outlines

Add a subtle `1px` outline with low opacity to images for consistent depth. The color must be pure black in light mode (`oklch(0 0 0 / 0.1)`) and pure white in dark mode (`oklch(1 0 0 / 0.1)`), never a near-black like slate, zinc, or any tinted neutral. A tinted outline picks up the surface color underneath it and reads as dirt on the image edge.

### 12. Scale on Press

A subtle `scale(0.96)` on click gives buttons tactile feedback. Always use `0.96`. Never use a value smaller than `0.95` — anything below feels exaggerated. Add a `static` prop to disable it when motion would be distracting.

### 13. Skip Animation on Page Load

Use `initial={false}` on `AnimatePresence` to prevent enter animations on first render. Verify it doesn't break intentional entrance animations.

### 14. Never Use `transition: all`

Always specify exact properties: `transition-property: scale, opacity`. Tailwind's `transition-transform` covers `transform, translate, scale, rotate`.

### 15. Use `will-change` Sparingly

Only for `transform`, `opacity`, `filter` — properties the GPU can composite. Never use `will-change: all`. Only add when you notice first-frame stutter.

### 16. Minimum Hit Area

Interactive elements should prefer a 44×44px hit area for touch or mobile contexts. In dense desktop interfaces, use at least 40×40px. Extend with a pseudo-element if the visible element is smaller. Never let hit areas of two elements overlap.

### 17. Match Icon Stroke to Text Weight

An icon next to text carries the text's optical weight: `1.5px` stroke beside regular (400) text, `2px` beside semibold (600). One stroke weight per icon set; never mix libraries on one surface.

### 18. One SVG, Recolored per State

Icons use `currentColor` and get their states (hover, selected, disabled) from CSS color and opacity, never from separate assets. Outline variant is the default; fill variant marks the active state.

### 19. Motion Restraint

No custom animation on high-frequency interactions: the attention cost repeats on every trigger. Motion is never the only feedback channel; every animated state change also needs a static cue such as color, icon, or label.

## Common Mistakes

| Mistake | Fix |
| --- | --- |
| Same border radius on parent and child | Calculate `outerRadius = innerRadius + padding` |
| Icons look off-center | Adjust optically with padding or fix SVG directly |
| Border used only to fake elevation | Use layered `box-shadow` with transparency; keep structural and state borders |
| Jarring staged entrance or contextual exit | Stagger infrequent entrances and keep context-preserving exits subtle |
| Numbers cause layout shift | Apply `tabular-nums` |
| Heavy text on macOS | Apply `antialiased` to root |
| Animation plays on page load | Add `initial={false}` to `AnimatePresence` |
| `transition: all` on elements | Specify exact properties |
| First-frame animation stutter | Add `will-change: transform` (sparingly) |
| Tiny hit areas on small controls | Extend with a pseudo-element to 44×44px for touch/mobile, or at least 40×40px in dense desktop UI |
| Hairline icon beside bold text | Match the stroke width to the text weight |
| Separate icon assets per state | One `currentColor` SVG, states via CSS |
| Filled icons everywhere | Outline as default, fill only for the active state |
| Entrance animation on every hover or keystroke | Instant feedback or ≤150ms opacity/color transition |

## Review Output Format

Use `full` when no review mode is supplied.

| Mode | Coverage | Finding cap |
| --- | --- | --- |
| `quick` | Primary user path and highest-traffic states; report only `HIGH` and `MEDIUM` issues | 5 |
| `full` | Entire requested scope across typography, surfaces, animations, icons, and performance | 15 |

### Scope and Coverage

State the mode, exact scope, framework, styling conventions, and any review boundary. Show what was actually inspected:

| Category | Evidence inspected | Result |
| --- | --- | --- |
| Typography | Files, components, states, or checks | Findings count, `Clear`, or `Not reviewed` with a reason |

Include all five Quick Reference categories. Never imply an uninspected surface was reviewed.

### Findings

Group findings by principle. Use a markdown table with **Severity**, **Location**, **Before**, **After**, and **Why** columns. Include every change made or proposed, not a subset. Never use separate "Before:" / "After:" lines.

- **Severity**: `HIGH` makes an interaction inaccessible, misleading, unreadable, or repeatedly disruptive; `MEDIUM` creates a noticeable usability or consistency problem; `LOW` is isolated polish and appears only in `full` mode.
- **Location**: cite `path/to/file:line`. If the artifact has no source files, cite the exact screen and component instead.
- **Before / After**: show the current implementation and an actionable replacement.
- **Why**: name the violated principle and explain its user impact.

Consolidate a repeated systemic issue into one row and list every affected location. Omit principles with no findings and never pad the report to reach the cap.

### Example

#### Concentric border radius
| Severity | Location | Before | After | Why |
| --- | --- | --- | --- | --- |
| LOW | `src/Card.tsx:28` | `rounded-xl` on card + `rounded-xl` on inner button (`p-2`) | `rounded-2xl` on card (`8 + 8 = 16`), `rounded-lg` on inner button | Nested corners should be concentric |
| LOW | `src/card.css:11` | `border-radius: 16px` on both nested surfaces | Outer `24px`, inner `16px` with `8px` padding | Equal nested radii make the inner surface look pinched |

#### Tabular numbers
| Severity | Location | Before | After | Why |
| --- | --- | --- | --- | --- |
| MEDIUM | `src/Counter.tsx:17` | `<span>{count}</span>` | `<span className="tabular-nums">{count}</span>` | Proportional digits cause changing values to shift |
| LOW | `src/timer.css:8` | Default numerals on a timer | Add `font-variant-numeric: tabular-nums` to the timer | Equal-width digits keep the timer stable |

#### Scale on press
| Severity | Location | Before | After | Why |
| --- | --- | --- | --- | --- |
| LOW | `src/Button.tsx:19` | `<button className="...">` | Add `active:scale-[0.96] transition-transform` | Press feedback makes the control feel responsive |
| MEDIUM | `src/button.css:24` | `scale(0.9)` on press | Raise to `scale(0.96)` | Anything below `0.95` feels exaggerated |

### Considered but Rejected

Include 1–3 real candidates in `quick` mode and 2–5 in `full` mode:

| Location | Candidate | Rejected because |
| --- | --- | --- |
| `src/Card.tsx:28` | Increase the shadow | Existing depth matches the shared surface token; changing one card would reduce consistency |

Do not invent filler. If the scope contains fewer borderline candidates, include the ones that exist and say so.

### Verification and Verdict

After the findings:

1. **Verification**: list the exact commands or interactions run and their observed results. Walk every relevant state and inspect motion at 10% speed when animation is involved. If a check was not run, label it **Not verified** and state what remains.
2. **Verdict**: `Block` if any `HIGH` finding remains, `Needs changes` if only `MEDIUM` or `LOW` findings remain, and `Approve` only when no actionable findings remain. List every unverified check beside the verdict.

When there are no findings, omit the findings table, state "No actionable interface-polish findings", report verification and rejected candidates, and end with `Approve`.

-------------------

---
name: better-colors
description: OKLCH color space and color usage for web projects. Convert hex/rgb/hsl to oklch, generate palettes, check contrast, handle gamut boundaries, theme with Tailwind v4, and apply color with meaning. Triggers on oklch, color conversion, palette generation, contrast ratio, gamut, display p3, design tokens, semantic color tokens, hue drift, chroma, dark mode colors, accent color, color meaning, light and dark appearance, increased contrast.
---

# OKLCH Colors

OKLCH is a perceptually uniform color space where lightness, chroma, and hue are useful design controls. Use it when the project already uses OKLCH, when creating a new color system, or when the user asks for conversion or palette work. Otherwise preserve the project's established tokens and notation: a consistent hex or RGB token system is better than introducing a second color representation for an isolated fix. To explore interactively, visit [oklch.fyi](https://oklch.fyi).

## Quick Reference

| Category | When to use | Reference |
| --- | --- | --- |
| Conversion | Hex/rgb/hsl to oklch | [color-conversion.md](color-conversion.md) |
| Palettes | Generate scales, multi-hue, dark mode | [palette-generation.md](palette-generation.md) |
| Contrast | APCA/WCAG checks, reporting failures, fixing on request | [accessibility-contrast.md](accessibility-contrast.md) |
| Gamut & Tailwind | P3 fallbacks, `@theme` scales, gamut clamping | [gamut-and-tailwind.md](gamut-and-tailwind.md) |
| Usage | Semantic tokens, one meaning per color, primary-action emphasis, appearance variants | [color-usage.md](color-usage.md) |

## Core Principles

### 1. Use a Perceptual Color Space

- **Respect the existing system.** Do not convert notation merely because this skill was loaded. Reuse the project's semantic tokens and authoring format unless the task includes a color-system migration.
- **Perceptual uniformity.** Equal L steps = equal brightness. `oklch(0.5 ...)` is visually mid. HSL's `lightness: 50%` varies wildly by hue.
- **Stable hue.** HSL blue shifts toward purple as lightness changes. OKLCH hue stays constant across the full lightness range.
- **Independent chroma.** Chroma is an absolute measure of colorfulness that doesn't depend on lightness. HSL saturation does.
- **Finite gamut.** Not every oklch value maps to a displayable sRGB color. High-chroma values at certain hues will clip; gamut awareness is required.

### 2. Write and Format OKLCH Consistently

```
oklch(L C H)
oklch(L C H / alpha)
```

| Channel | Range | Description |
| --- | --- | --- |
| L (Lightness) | 0–1 | 0 = black, 1 = white. Perceptually uniform. |
| C (Chroma) | 0–~0.4 | Colorfulness. 0 = gray. Max depends on L and H. |
| H (Hue) | 0–360 | Hue angle in degrees. |
| alpha | 0–1 | Optional transparency. Slash syntax. |

```css
oklch(0.637 0.237 25.331)
oklch(0.8 0.05 200 / 0.5)
```

Use three decimal places for L and C and up to three for H. Drop trailing zeros and format `-0` as `0`. OKLCH is Baseline 2023; when support requirements are unusually broad, check the target project's browser matrix instead of relying on a fixed global-coverage percentage.

### 3. Measure Contrast, Gamut, and Palette Behavior

| Rule | Value |
| --- | --- |
| Light/dark boundary | L > 0.73 = light background → dark text; below it, light text still scores higher |
| Lightness gap (light bg) | Foreground L < 0.35 when background L > 0.9 |
| Lightness gap (dark bg) | Foreground L > 0.9 when background L < 0.25 |
| Hue drift threshold | > 10° spread across palette steps = visible drift |
| APCA body text | \|Lc\| >= 75 minimum, >= 90 preferred |
| APCA non-body text | \|Lc\| >= 60 minimum |
| WCAG 2 normal text | 4.5:1 AA, 7:1 AAA |
| Contrast fix (only when asked) | Adjust L first; preserve C and H when possible, then remeasure the rendered pair |

## Common Mistakes

| Issue | Fix |
| --- | --- |
| Raw color bypasses the project's semantic token system | Reuse or add the correct role token in the project's existing notation |
| Isolated OKLCH value introduced into a hex/RGB codebase | Preserve the established notation unless the task includes a color-system migration |
| HSL palette ramp with hue drift | Rebuild with constant oklch hue |
| Failing contrast (check foreground vs its background using APCA) | Report the pair, its measured Lc and the threshold it misses; change colors only when asked (then adjust L, keep C and H) |
| High chroma without gamut check | Clamp to max chroma for the L/H in sRGB |
| Same absolute C across different hues | Use same C% (percentage of max) for consistent vividness |
| P3 color without sRGB fallback | Add `@media (color-gamut: p3)` pattern |
| Dark mode created by mechanically reversing the light palette | Use the light palette as a starting point, then tune chroma and lightness and recheck every foreground/background pair |
| Hex in Tailwind v4 `@theme` | Convert to oklch values |
| Alpha with comma syntax | Use slash: `oklch(L C H / alpha)` |
| Same hue means two different things (link color reused decoratively) | One color, one meaning; give the second use a neutral |
| Semantic token used outside its role (separator as text) | Add a token for the missing role; never borrow by value |
| Several colored control backgrounds in one view | Fill only the single primary action; secondaries stay neutral |
| Palette verified only in light mode | Recheck every foreground/background pair in both appearances |

## Review Output Format

Use this format only when the user asks for a standalone color review. When `better-interface` orchestrates the review, provide domain evidence and findings to that skill and let its output format, severity scale, consolidation rules, cap, and verdict take precedence.

Present the standalone review in two parts.

### Findings

Group all confirmed findings by principle. Use a markdown table with **Severity**, **Location**, **Before**, **After**, and **Why** columns. Never use separate "Before:" / "After:" lines.

- **Severity**: `HIGH` makes content unreadable or assigns a misleading semantic color; `MEDIUM` creates a noticeable theme, gamut, or consistency failure; `LOW` is isolated polish.
- **Location**: cite `path/to/file:line`. If the artifact has no source files, cite the exact screen and component instead.
- **Before / After**: show the current value or token and the exact replacement.
- **Why**: name the violated principle and include measured contrast or gamut evidence when relevant.

Consolidate a repeated systemic issue into one row and list every affected location. Omit principles with no findings.

| Severity | Location | Before | After | Why |
| --- | --- | --- | --- | --- |
| MEDIUM | `src/theme.css:18` | `color: #3b82f6` | `color: oklch(0.623 0.188 259.815)` | New project colors use OKLCH tokens |
| MEDIUM | `src/palette.ts:31` | Same absolute C across hues | Same C% of each hue's maximum chroma | Equal chroma values do not appear equally vivid across hues |
| HIGH | `src/theme.css:52` | P3 color with no fallback | Add an sRGB fallback before `@media (color-gamut: p3)` | The color fails on non-P3 displays |

### Verification and Verdict

After the findings:

1. **Verification**: list the exact checks run and their observed results, including contrast measurements, gamut checks, and both light and dark appearances when applicable. If a check was not run, state what still needs verification.
2. **Verdict**: `Block` if any `HIGH` finding remains, `Needs changes` if only `MEDIUM` or `LOW` findings remain, and `Approve` only when no actionable findings remain.

When there are no findings, omit the table, state "No actionable color findings", report verification, and end with `Approve`.


----------
---
name: better-layout
description: Layout structure for web interfaces, from grouping and alignment to reading order, progressive disclosure, and adaptive breakpoints. Use when structuring a page or component, spacing or aligning controls, deciding what collapses at small sizes, handling RTL layout direction, or reviewing frontend code for layout. Triggers on layout, spacing, alignment, grouping, negative space, whitespace, visual hierarchy, reading order, progressive disclosure, breakpoints, responsive layout, container queries, safe area, full-bleed, edge-to-edge, layout margins, RTL layout, logical properties.
---

# Layout that communicates structure

Layout communicates before a single word is read: position, spacing, and alignment carry hierarchy on their own, and generous space beats decoration. A good layout also survives stress: resize it, translate it, mirror it for RTL, and it should still hold together. Apply these principles when building or reviewing UI code, and express every change in the project's existing styling system (Tailwind, plain CSS, CSS-in-JS); never introduce a second styling approach.

Hit-area sizes and focus behavior are covered by the `better-accessibility` skill; visual polish (radius, shadows, animation) by the `better-ui` skill; line length and text spacing by the `better-typography` skill.

Treat the numeric values below as starting points for interfaces without an established density or spacing system. Preserve deliberate platform chrome, compact professional tools, and project tokens when they remain usable under hit-area, zoom, localization, and viewport stress tests.

## Quick Reference

| Category | When to Use |
| --- | --- |
| [Grouping & Alignment](grouping-and-alignment.md) | Space vs separators, alignment edges, logical properties, importance ordering |
| [Spacing & Adaptivity](spacing-and-adaptivity.md) | Spacing between targets, layout margins, progressive disclosure, full-bleed content, breakpoints, i18n growth |

## Core Principles

### 1. Group with Space, Not Lines

Negative space is the primary grouping tool; background shapes second; separator lines last, only where space alone can't carry the structure. The gap between groups must be at least 2× the gap within a group (`8px` intra-group → `16px`+ inter-group), or the grouping reads as noise.

### 2. Keep Controls Distinct from Content

Interactive elements must look interactive: a background shape, a border, or a consistent placement zone. Never style a control identically to adjacent static text.

### 3. Align to Shared Edges

Pick alignment edges and stick to them; every stray edge reads as noise. Use one project spacing step for each level of subordination (`16px` is a useful default). Use logical properties (`padding-inline-start`, `margin-inline-end`) for direction-dependent layout; reserve physical left/right for genuinely physical geometry.

### 4. Order by Importance

The most important content sits near the top and the leading edge; reading order flows top-to-bottom, leading-to-trailing. Think in leading/trailing, not left/right.

### 5. Hint at Hidden Content

Progressive disclosure needs a visible affordance. Use the project's established cue; without one, let the next item peek `16–32px` past the scroll edge or show a disclosure control. Content hidden with zero cue may as well not exist.

### 6. Breathing Room Between Targets

Without an established density system, start with `12px` between adjacent bordered or filled controls and `24px` of clearance around borderless text- and icon-only controls. Compact layouts may use less when `better-accessibility` hit areas do not overlap and the controls remain visually distinct.

### 7. Inset Buttons from the Edges

In content layouts, keep full-width buttons inside the layout margins (start near `16px` inline on mobile) with a visible radius. Edge-to-edge actions are acceptable when they intentionally follow established platform or application chrome, account for safe areas, and remain distinguishable from system UI.

### 8. Content Bleeds, Controls Float

Backgrounds and media extend to the viewport edges; controls and text stay inside the layout margins and safe areas (`env(safe-area-inset-*)`). Sticky chrome floats above the content layer, it doesn't dam it.

### 9. Hold Structure Until It Breaks

Breakpoints come from the content, not device presets. Keep the expanded layout as long as it genuinely fits and collapse late; prefer container queries for component-level adaptation. Test the smallest and largest sizes first.

### 10. Plan for Growth and Clipping

Plan for substantial and language-dependent string growth rather than relying on a universal percentage: no fixed widths or heights on text containers, and let rows wrap. Never park critical actions where resizing or scrolling clips them; keep them reachable in the normal flow or stable chrome appropriate to the product.

## Common Mistakes

| Mistake | Fix |
| --- | --- |
| Separator line where spacing would do | Remove the line, double the gap between groups |
| `margin-left` / `padding-right` in a localizable layout | `margin-inline-start` / `padding-inline-end` |
| Content-layout button accidentally touches the viewport | Inset within the project margins; preserve intentional platform chrome |
| Carousel/scroller that looks complete | Let the next item peek `16–32px` past the edge |
| Adjacent controls merge or expanded hit areas overlap | Increase the gap using the project scale; use `12px`/`24px` as starting points |
| Breakpoints at 768/1024 because they're the defaults | Break where the content actually stops fitting |
| Fixed-width text container sized to one language | `max-width` + wrapping; test pseudo-localization and representative locales |
| Primary action at the clip-prone bottom of a pane | Sticky positioning or stable chrome with safe-area padding |

## Review Output Format

Use this format only when the user asks for a standalone layout review. When `better-interface` orchestrates the review, provide domain evidence and findings to that skill and let its output format, severity scale, consolidation rules, cap, and verdict take precedence.

Present the standalone review in two parts.

### Findings

Group all confirmed findings by principle. Use a markdown table with **Severity**, **Location**, **Before**, **After**, and **Why** columns. Never use separate "Before:" / "After:" lines.

- **Severity**: `HIGH` blocks content or an action at a supported viewport; `MEDIUM` harms hierarchy, reading order, or adaptability; `LOW` is isolated alignment or spacing polish.
- **Location**: cite `path/to/file:line`. If the artifact has no source files, cite the exact screen and component instead.
- **Before / After**: show the current layout and an actionable replacement.
- **Why**: name the violated principle and its effect on comprehension or adaptability.

Consolidate a repeated systemic issue into one row and list every affected location. Omit principles with no findings.

### Example

#### Group with space, not lines
| Severity | Location | Before | After | Why |
| --- | --- | --- | --- | --- |
| LOW | `src/Settings.tsx:41` | `border-b` on every settings row | Remove borders; use `space-y-2` within groups and `space-y-8` between groups | Spacing communicates grouping with less visual noise |
| LOW | `src/ProfileForm.tsx:58` | `<hr>` between form sections | Replace with `mt-10` on each section heading | Section hierarchy should not depend on repeated rules |

#### Align to shared edges
| Severity | Location | Before | After | Why |
| --- | --- | --- | --- | --- |
| LOW | `src/Card.tsx:24` | Card text at `pl-4`, card icon at `pl-3` | Align both to the same `pl-4` edge | Shared edges create a legible structure |
| MEDIUM | `src/Nav.css:19` | `margin-left: 16px` | `margin-inline-start: 16px` | Physical properties break direction-aware layouts |

### Verification and Verdict

After the findings:

1. **Verification**: list the exact checks run and their observed results across the relevant viewport widths, reading order, zoom, and RTL state. If a check was not run, state what still needs verification.
2. **Verdict**: `Block` if any `HIGH` finding remains, `Needs changes` if only `MEDIUM` or `LOW` findings remain, and `Approve` only when no actionable findings remain.

When there are no findings, omit the tables, state "No actionable layout findings", report verification, and end with `Approve`.

-----

---
name: better-ui
description: Design engineering principles for making interfaces feel polished. Use when building UI components, reviewing frontend code, implementing animations, hover states, shadows, borders, micro-interactions, enter/exit animations, choosing or reviewing icons, or any visual detail work. Triggers on UI polish, design details, "make it feel better", "feels off", stagger animations, border radius, optical alignment, image outlines, box shadows, icons, icon stroke weight, icon states, motion restraint.
---

# Details that make interfaces feel better

Great interfaces rarely come from a single thing. It's usually a collection of small details that compound into a great experience. Apply these principles when building or reviewing UI code.

When reviewing, slow the interface down: replay motion at 10% speed in the browser's Animations panel and walk every state: hover, focus, active, loading, empty. What feels off at 10% speed is what's subtly wrong at full speed.

Preserve the project's component library, tokens, and density. Match its established motion language except where a principle below prescribes an exact interaction pattern.

Typography (text wrapping, font rendering, tabular numbers, spacing) is covered by the `better-typography` skill; use that for anything text-related. Accessibility (hit areas, focus states, keyboard support, ARIA, reduced motion) is covered by the `better-accessibility` skill. Layout structure (grouping, spacing between sections, breakpoints, spatial RTL) is covered by the `better-layout` skill.

## Quick Reference

| Category | When to Use |
| --- | --- |
| [Surfaces](surfaces.md) | Border radius, optical alignment, shadows, image outlines |
| [Animations](animations.md) | Interruptible animations, enter/exit transitions, icon animations, scale on press, motion restraint |
| [Icons](icons.md) | Icon stroke weight, states via `currentColor`, outline vs fill, sizing, RTL flipping |
| [Performance](performance.md) | Transition specificity, `will-change` usage |

## Core Principles

### 1. Concentric Border Radius

Outer radius = inner radius + padding. Mismatched radii on nested elements is the most common thing that makes interfaces feel off.

### 2. Optical Over Geometric Alignment

When geometric centering looks off, align optically. Buttons with icons, play triangles, and asymmetric icons all need manual adjustment.

### 3. Shadows for Elevation, Borders for Structure

For buttons, cards, and containers whose border exists only to create depth, prefer layered transparent `box-shadow` values. Keep borders that communicate structure or state: dividers, layout separators, and selected or focus states.

### 4. Interruptible Animations

Use CSS transitions for interactive state changes: they can be interrupted mid-animation. Reserve keyframes for staged sequences that run once.

### 5. Split and Stagger Enter Animations

For an infrequent staged entrance where sequence helps communicate hierarchy, break content into semantic chunks and stagger them by ~100ms instead of animating one container. Do not stagger routine, high-frequency interactions.

### 6. Subtle Exit Animations

Use a small fixed `translateY` instead of full height. Exits should be softer than enters. Use `ease-out` for both enter and exit transitions.

### 7. Contextual Icon Animations

Animate icons with `opacity`, `scale`, and `blur` instead of toggling visibility. Use exactly these values: scale from `0.25` to `1`, opacity from `0` to `1`, blur from `4px` to `0px`. If the project has `motion` or `framer-motion` in `package.json`, match that package's import path (or the established nearby imports when both exist) and use `transition: { type: "spring", duration: 0.3, bounce: 0 }`; bounce must always be `0`. If no motion library is installed, keep both icons in the DOM (one absolute-positioned) and cross-fade with CSS transitions using `cubic-bezier(0.2, 0, 0, 1)`; this gives both enter and exit animations without any dependency.

### 8. Image Outlines

Add a subtle `1px` outline with low opacity to images for consistent depth. The color must be pure black in light mode (`oklch(0 0 0 / 0.1)`) and pure white in dark mode (`oklch(1 0 0 / 0.1)`), never a near-black like slate, zinc, or any tinted neutral. A tinted outline picks up the surface color underneath it and reads as dirt on the image edge.

### 9. Scale on Press

A subtle `scale(0.96)` on click gives buttons tactile feedback. Always use `0.96`. Never use a value smaller than `0.95`: anything below feels exaggerated. Add a `static` prop to disable it when motion would be distracting.

### 10. Skip Animation on Page Load

Use `initial={false}` on `AnimatePresence` to prevent enter animations on first render. Verify it doesn't break intentional entrance animations.

### 11. Never Use `transition: all`

Always specify exact properties: `transition-property: scale, opacity`. Tailwind's `transition-transform` covers `transform, translate, scale, rotate`.

### 12. Use `will-change` Sparingly

Only for `transform`, `opacity`, `filter`, the properties the GPU can composite. Never use `will-change: all`. Only add when you notice first-frame stutter.

### 13. Match Icon Stroke to Text Weight

An icon next to text carries the text's optical weight: `1.5px` stroke beside regular (400) text, `2px` beside semibold (600). One stroke weight per icon set; never mix libraries on one surface.

### 14. One SVG, Recolored per State

Icons use `currentColor` and get their states (hover, selected, disabled) from CSS color and opacity, never from separate assets. Outline variant is the default; fill variant marks the active state.

### 15. Motion Restraint

No custom animation on high-frequency interactions: the attention cost repeats on every trigger. Motion is never the only feedback channel; every animated state change also needs a static cue (color, icon, label).

## Common Mistakes

| Mistake | Fix |
| --- | --- |
| Same border radius on closely nested parent and child | Calculate `outerRadius = innerRadius + padding` |
| Icons look off-center | Adjust optically with padding or fix SVG directly |
| Border used only to fake elevation | Use layered `box-shadow` with transparency; keep structural and state borders |
| Jarring staged entrance or contextual exit | Stagger infrequent entrances and keep context-preserving exits subtle |
| Stateful icon or toggle animates its default state on page load | Add `initial={false}` to that `AnimatePresence`; preserve intentional page entrances |
| `transition: all` on elements | Specify exact properties |
| First-frame animation stutter | Add `will-change: transform` (sparingly) |
| Hairline icon beside bold text | Match the stroke width to the text weight |
| Separate icon assets per state | One `currentColor` SVG, states via CSS |
| Filled icons everywhere | Outline as default, fill only for the active state |
| Entrance animation on every hover or keystroke | Instant feedback or ≤150ms opacity/color transition |

## Review Output Format

Use this format only when the user asks for a standalone UI-polish review. When `better-interface` orchestrates the review, provide domain evidence and findings to that skill and let its output format, severity scale, consolidation rules, cap, and verdict take precedence.

Present the standalone review in two parts.

### Findings

Group all confirmed findings by principle. Use a markdown table with **Severity**, **Location**, **Before**, **After**, and **Why** columns. Never use separate "Before:" / "After:" lines.

- **Severity**: `HIGH` makes an interaction misleading, unresponsive, or repeatedly disruptive; `MEDIUM` creates a noticeable craft or consistency problem; `LOW` is isolated polish.
- **Location**: cite `path/to/file:line`. If the artifact has no source files, cite the exact screen and component instead.
- **Before / After**: show the current implementation and an actionable replacement.
- **Why**: name the violated principle and explain how it affects the interface.

Consolidate a repeated systemic issue into one row and list every affected location. Omit principles with no findings.

### Example

#### Concentric border radius
| Severity | Location | Before | After | Why |
| --- | --- | --- | --- | --- |
| LOW | `src/Card.tsx:28` | `rounded-xl` on card + `rounded-xl` on inner button (`p-2`) | `rounded-2xl` on card (`8 + 8 = 16`), `rounded-lg` on inner button | Nested corners should be concentric |
| LOW | `src/card.css:11` | `border-radius: 16px` on both nested surfaces | Outer `24px`, inner `16px` with `8px` padding | Equal nested radii make the inner surface look pinched |

#### Scale on press
| Severity | Location | Before | After | Why |
| --- | --- | --- | --- | --- |
| LOW | `src/Button.tsx:19` | `<button className="...">` | Add `active:scale-[0.96] transition-transform` | Press feedback makes the control feel responsive |
| MEDIUM | `src/button.css:24` | `scale(0.9)` on press | Raise to `scale(0.96)` | Anything below `0.95` feels exaggerated |

### Verification and Verdict

After the findings:

1. **Verification**: list the exact checks run and their observed results. Walk every relevant state and inspect motion at 10% speed when animation is involved. If a check was not run, state what still needs verification.
2. **Verdict**: `Block` if any `HIGH` finding remains, `Needs changes` if only `MEDIUM` or `LOW` findings remain, and `Approve` only when no actionable findings remain.

When there are no findings, omit the tables, state "No actionable UI-polish findings", report verification, and end with `Approve`.
