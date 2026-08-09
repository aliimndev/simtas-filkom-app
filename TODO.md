Redesign the existing navbar/header component to closely match the visual style and behavior of the reference navbar I provided.

IMPORTANT:

* Do NOT rebuild the navbar from scratch.
* Do NOT change the existing navigation structure, routes, functionality, or business logic.
* Keep the existing component and modify only its visual design, spacing, styling, and responsive behavior.
* Do not copy the original code. Recreate the same design language using the existing project architecture.

Reference style:

* Minimal, premium, modern SaaS navbar.
* Similar visual direction to the provided reference.
* Clean layout with generous whitespace.
* Subtle borders instead of heavy shadows.
* Rounded container when scrolling.
* Glassmorphism effect using a subtle translucent background + backdrop blur.
* Smooth 300ms transitions.
* Clean typography and muted navigation colors.
* Avoid excessive gradients, animations, or decorative elements.

Header behavior:

* Navbar remains fixed at the top.
* At the top of the page, keep the navbar visually lightweight and transparent/minimal.
* When the user scrolls down:

  * Slightly reduce the navbar width.
  * Add a subtle semi-transparent background.
  * Add backdrop blur.
  * Add a thin border.
  * Add rounded corners.
  * Reduce horizontal padding slightly.
  * Transition everything smoothly.
* Keep the scroll transformation subtle and polished, not dramatic.

Desktop layout:

* Logo/brand on the left.
* Navigation links centered.
* Action buttons on the right.
* Maintain balanced spacing between all elements.
* Navigation links should use muted foreground colors and become stronger on hover.
* Buttons should follow the existing project's design system.

Mobile layout:

* Keep the existing mobile navigation functionality.
* Logo on the left.
* Menu button on the right.
* Use a clean animated Menu/X transition.
* Open the navigation inside a rounded panel below the header.
* Use subtle border, background, and backdrop blur.
* Keep spacing comfortable and touch-friendly.
* Ensure the mobile menu feels like the same design system as desktop.

Visual details:

* Border: subtle 1px border using existing theme variables.
* Background: translucent background when scrolled.
* Backdrop blur: subtle, not excessive.
* Border radius: modern rounded corners.
* Shadow: very subtle or none.
* Hover states: smooth and understated.
* Focus states: accessible and consistent with the existing design system.
* Dark mode: ensure the navbar looks equally polished in dark mode.

Technical requirements:

* Reuse the existing components and utilities.
* Reuse existing Button, cn(), icons, theme variables, and design tokens whenever possible.
* Do not install unnecessary dependencies.
* Do not introduce a new UI library.
* Keep TypeScript strict and clean.
* Preserve all existing routes and links.
* Preserve accessibility attributes.
* Make the component responsive across desktop, tablet, and mobile.

Before editing:

1. Inspect the current navbar implementation.
2. Identify its existing structure and functionality.
3. Apply the reference visual style to that implementation.
4. Do not modify unrelated components.

After editing:

* Verify desktop navbar.
* Verify scroll transformation.
* Verify mobile menu.
* Verify light/dark mode.
* Verify hover/focus states.
* Run TypeScript/lint checks.
* Fix any visual or TypeScript issues you introduce.

The goal is:
"Keep my existing navbar functionality, but make it visually feel like the reference navbar — minimal, premium, clean, responsive, and Vercel-style."
