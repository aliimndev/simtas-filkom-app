<script lang="ts">
  import type { Snippet } from "svelte";

  let {
    kind = "error",
    children,
  }: {
    kind?: "error" | "success";
    children: Snippet;
  } = $props();

  const cls = $derived(
    kind === "success"
      ? "border-success/40 bg-success-50 text-success"
      : "border-danger-700/40 bg-danger-50 text-danger-700",
  );
</script>

<div role={kind === "error" ? "alert" : "status"} class="rounded-md {cls} px-3 py-2 text-sm">
  {@render children()}
</div>
