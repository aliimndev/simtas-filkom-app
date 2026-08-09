Remove the BootLoader from the PUBLIC AREA.

Context:
This is an internal university application used by lecturers, students, and academic staff.

The current BootLoader displays a full-screen overlay for approximately 2 seconds on the initial visit.

For this type of internal productivity application, I do not want a blocking visual intro or artificial waiting period.

### Goal

The public application should become interactive as soon as the actual UI is ready.

Preferred behavior:

Open SIMTAS
→ Render UI
→ User can immediately interact

NOT:

Open SIMTAS
→ Full-screen BootLoader
→ Artificial wait
→ UI becomes interactive

### Task

1. Identify where the BootLoader is mounted/rendered.
2. Remove it from the public application flow.
3. Remove any related:

   * timer
   * autoplay logic
   * skip handler
   * full-screen overlay
   * unnecessary loading state
   * CSS specifically required only by the BootLoader
4. Remove unused imports after the BootLoader is removed.
5. Do not replace it with another full-screen loader.
6. Do not add artificial delays.
7. Do not modify authenticated/dashboard UI.
8. Do not change business logic or API contracts.

### Preserve useful loading behavior

Do NOT remove legitimate loading states for actual asynchronous operations.

For example:

* API requests → component-level loading state
* Form submission → button pending state
* Navigation → appropriate route loading state
* Data fetching → skeleton/loading state where necessary

The distinction should be:

**Blocking visual intro:** REMOVE

**Real application loading state:** KEEP

### Verification

After removing BootLoader:

* Landing page renders correctly
* Public navbar works immediately
* Mobile hamburger works with one tap
* CTA buttons work with one tap
* Public navigation works
* Login navigation works
* No console errors
* No unused BootLoader imports
* No unnecessary timer remains
* Desktop regression test passes
* Mobile viewport test passes

Also verify that removing BootLoader does not expose an uninitialized or broken UI state.

Do not make unrelated optimizations during this task.
Keep the change minimal and production-safe.
