<script lang="ts">
  import type { Snippet } from "svelte";

  let {
    open = $bindable(false),
    labelledBy,
    initialFocus = "#dialog-initial",
    panelClass = "w-full max-w-lg",
    onClose,
    children,
  }: {
    open?: boolean;
    labelledBy?: string;
    initialFocus?: string;
    panelClass?: string;
    onClose?: () => void;
    children: Snippet;
  } = $props();

  let dialog = $state<HTMLElement>();
  let returnFocus: HTMLElement | null = null;

  const FOCUSABLE =
    "a[href], button:not([disabled]), input:not([disabled]), textarea:not([disabled]), select:not([disabled]), [tabindex]:not([tabindex='-1'])";

  function handleKeydown(event: KeyboardEvent) {
    if (event.key === "Escape") {
      onClose?.();
      return;
    }
    if (event.key !== "Tab" || !dialog) return;

    const focusable = Array.from(dialog.querySelectorAll<HTMLElement>(FOCUSABLE));
    if (focusable.length === 0) return;

    const first = focusable[0];
    const last = focusable[focusable.length - 1];
    if (event.shiftKey && document.activeElement === first) {
      event.preventDefault();
      last.focus();
    } else if (!event.shiftKey && document.activeElement === last) {
      event.preventDefault();
      first.focus();
    }
  }

  $effect(() => {
    if (!open) return;
    returnFocus = document.activeElement as HTMLElement | null;
    const previousOverflow = document.body.style.overflow;
    document.body.style.overflow = "hidden";
    const frame = requestAnimationFrame(() => {
      dialog?.querySelector<HTMLElement>(initialFocus)?.focus();
    });
    return () => {
      cancelAnimationFrame(frame);
      document.body.style.overflow = previousOverflow;
      requestAnimationFrame(() => returnFocus?.focus());
    };
  });
</script>

{#if open}
  <div
    bind:this={dialog}
    class="fixed inset-0 z-50 flex items-center justify-center p-4"
    role="dialog"
    tabindex="-1"
    aria-modal="true"
    aria-labelledby={labelledBy}
    onkeydown={handleKeydown}
  >
    <button
      type="button"
      class="absolute inset-0 bg-black/40"
      aria-label="Tutup dialog"
      onclick={onClose}
    ></button>
    <div class="relative z-10 rounded-2xl border border-st-stroke bg-st-surface p-6 shadow-xl {panelClass}">
      {@render children()}
    </div>
  </div>
{/if}
