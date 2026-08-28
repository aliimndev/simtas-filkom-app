<script lang="ts">
  import type { Snippet } from "svelte";

  let {
    href,
    variant = "solid",
    class: className = "",
    children,
  }: {
    href: string;
    variant?: "solid" | "outline";
    class?: string;
    children: Snippet;
  } = $props();

  const cls = $derived(
    "accent-ring inline-flex h-12 items-center justify-center gap-2 rounded-full px-7 text-sm font-medium transition " +
    (variant === "solid"
      ? "bg-st-text text-st-bg hover:opacity-90"
      : "border border-st-stroke bg-st-surface text-st-text hover:border-(--st-accent-from)/40 hover:bg-st-surface-hi") +
    " " +
    className
  );
</script>

<a {href} class={cls}>{@render children()}</a>
