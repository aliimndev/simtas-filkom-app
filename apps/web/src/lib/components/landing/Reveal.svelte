<script lang="ts">
  import { onMount } from "svelte";
  import type { Snippet } from "svelte";

  let {
    delay = 0,
    class: className = "",
    children,
  }: {
    delay?: number;
    class?: string;
    children: Snippet;
  } = $props();

  let el: HTMLElement;

  onMount(() => {
    const reduce = window.matchMedia("(prefers-reduced-motion: reduce)").matches;
    if (reduce) {
      el.classList.add("is-visible");
      return;
    }
    const io = new IntersectionObserver(
      (entries) => {
        for (const e of entries) {
          if (e.isIntersecting) {
            el.classList.add("is-visible");
            io.disconnect();
          }
        }
      },
      { threshold: 0.15 },
    );
    io.observe(el);
    return () => io.disconnect();
  });
</script>

<div
  bind:this={el}
  class="reveal {className}"
  style={delay ? `animation-delay:${delay}ms` : ""}
>
  {@render children()}
</div>
